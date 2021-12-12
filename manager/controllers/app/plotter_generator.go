// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/url"
	"text/template"

	"emperror.dev/errors"
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	connectors "fybrik.io/fybrik/pkg/connectors/clients"
	pb "fybrik.io/fybrik/pkg/connectors/protobuf"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/serde"
	"fybrik.io/fybrik/pkg/storage"
	vault "fybrik.io/fybrik/pkg/vault"
	"github.com/Masterminds/sprig/v3"
	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewAssetInfo points to the provisoned storage and hold information about the new asset
type NewAssetInfo struct {
	Storage *storage.ProvisionedBucket
	Details *pb.DatasetDetails
}

// PlotterGenerator constructs a plotter based on the requirements (governance actions, data location) and the existing set of FybrikModules
type PlotterGenerator struct {
	Client                client.Client
	Log                   zerolog.Logger
	Modules               map[string]*app.FybrikModule
	Owner                 types.NamespacedName
	PolicyManager         connectors.PolicyManager
	Provision             storage.ProvisionInterface
	VaultConnection       vault.Interface
	ProvisionedStorage    map[string]NewAssetInfo
	StorageAccountRegions []string
}

// GetCopyDestination creates a Dataset for bucket allocation by implicit copies or ingest.
func (p *PlotterGenerator) GetCopyDestination(item DataInfo, destinationInterface *app.InterfaceDetails, geo string) (*app.DataStore, error) {
	// provisioned storage for COPY
	originalAssetName := item.DataDetails.Name
	var bucket *storage.ProvisionedBucket
	var err error
	if bucket, err = AllocateBucket(p.Client, p.Log, p.Owner, originalAssetName, geo); err != nil {
		p.Log.Error().Err(err).Msg("Bucket allocation failed")
		return nil, err
	}
	bucketRef := &types.NamespacedName{Name: bucket.Name, Namespace: utils.GetSystemNamespace()}
	if err = p.Provision.CreateDataset(bucketRef, bucket, &p.Owner); err != nil {
		p.Log.Error().Err(err).Msg("Dataset creation failed")
		return nil, err
	}

	// S3 endpoint should not include the url scheme only the host name
	// thus ignoring it if such exists.
	url, err := url.Parse(bucket.Endpoint)
	if err != nil {
		return nil, err
	}
	endpoint := url.Host
	datastore := &pb.DataStore{
		Type: pb.DataStore_S3,
		Name: "S3",
		S3: &pb.S3DataStore{
			Bucket:    bucket.Name,
			Endpoint:  endpoint,
			ObjectKey: originalAssetName + utils.Hash(p.Owner.Name+p.Owner.Namespace, 10),
		},
	}
	connection := serde.NewArbitrary(datastore)
	assetInfo := NewAssetInfo{
		Storage: bucket,
		Details: &pb.DatasetDetails{
			Name:       originalAssetName,
			Geo:        item.DataDetails.Geography,
			DataFormat: destinationInterface.DataFormat,
			DataStore:  datastore,
			Metadata:   item.DataDetails.TagMetadata,
		}}
	p.ProvisionedStorage[item.Context.DataSetID] = assetInfo
	logging.LogStructure("ProvisionedStorage element", assetInfo, p.Log, false, true)

	vaultSecretPath := vault.PathForReadingKubeSecret(bucket.SecretRef.Namespace, bucket.SecretRef.Name)
	vaultMap := make(map[string]app.Vault)
	vaultMap[string(app.WriteFlow)] = app.Vault{
		SecretPath: vaultSecretPath,
		Role:       utils.GetModulesRole(),
		Address:    utils.GetVaultAddress(),
	}
	// The copied asset needs creds for later to be read
	vaultMap[string(app.ReadFlow)] = app.Vault{
		SecretPath: vaultSecretPath,
		Role:       utils.GetModulesRole(),
		Address:    utils.GetVaultAddress(),
	}
	return &app.DataStore{
		Vault:      vaultMap,
		Connection: connection,
		Format:     destinationInterface.DataFormat,
	}, nil
}

// Adds the asset details, flows and templates to the given plotter spec.
// Write path is not yet implemented
func (p *PlotterGenerator) AddFlowInfoForAsset(item DataInfo, appContext *app.FybrikApplication, plotterSpec *app.PlotterSpec) error {
	datasetID := item.Context.DataSetID
	p.Log.Trace().Str(logging.DATASETID, datasetID).Msg("Choose modules for dataset")

	subflows := make([]app.SubFlow, 0)
	assets := map[string]app.AssetDetails{}
	templates := []app.Template{}

	// Set the value received from the catalog connector.
	vaultSecretPath := item.VaultSecretPath
	vaultMap := make(map[string]app.Vault)
	vaultMap[string(app.ReadFlow)] = app.Vault{
		SecretPath: vaultSecretPath,
		Role:       utils.GetModulesRole(),
		Address:    utils.GetVaultAddress(),
	}
	sourceDataStore := &app.DataStore{
		Connection: &item.DataDetails.Connection,
		Vault:      vaultMap,
		Format:     item.DataDetails.Interface.DataFormat,
	}

	assets[datasetID] = app.AssetDetails{
		AdvertisedAssetID: "",
		DataStore:         *sourceDataStore,
	}

	// DataStore for destination will be determined if an implicit copy is required
	var sinkDataStore *app.DataStore

	selectors, err := p.SelectModules(&item, appContext)
	if err != nil {
		return err
	}
	p.Log.Trace().Str(logging.DATASETID, datasetID).Msg("Generating a plotter")
	readSelector := selectors[app.Read]
	copySelector := selectors[app.Copy]
	// sanity
	if readSelector == nil && copySelector == nil {
		return errors.New("Can't generate a plotter - no modules available")
	}
	var copyCluster, readCluster string
	if copySelector != nil {
		p.Log.Trace().Str(logging.DATASETID, datasetID).Msg("Found copy module " + copySelector.GetModule().Name + " for dataset")
		// copy should be applied - allocate storage
		if sinkDataStore, err = p.GetCopyDestination(item, copySelector.Destination, copySelector.StorageAccountRegion); err != nil {
			p.Log.Error().Err(err).Msg("Allocation of storage for copy failed")
			return err
		}
		var copyDataAssetID = datasetID + "-copy"
		actions := createActionStructure(copySelector.Actions)
		if len(item.Configuration.ConfigDecisions[app.Copy].Clusters) > 0 {
			copyCluster = item.Configuration.ConfigDecisions[app.Copy].Clusters[0]
		} else {
			msg := "Coud not determine the cluster for copy"
			p.Log.Error().Str(logging.DATASETID, datasetID).Msg(msg)
			return errors.New(msg)
		}
		// The default capability scope is of type Asset
		scope := copySelector.ModuleCapability.Scope
		template := app.Template{
			Name: "copy",
			Modules: []app.ModuleInfo{{
				Name:  "copy",
				Type:  copySelector.Module.Spec.Type,
				Chart: copySelector.Module.Spec.Chart,
				Scope: scope,
			}},
		}
		templates = append(templates, template)
		copyAsset := app.AssetDetails{
			AdvertisedAssetID: datasetID,
			DataStore:         *sinkDataStore,
		}
		assets[copyDataAssetID] = copyAsset
		steps := []app.DataFlowStep{
			{
				Name:     "",
				Cluster:  copyCluster,
				Template: "copy",
				Parameters: &app.StepParameters{
					Source: &app.StepSource{
						AssetID: datasetID,
						API:     nil,
					},
					Sink: &app.StepSink{
						AssetID: copyDataAssetID,
					},
					API:     nil,
					Actions: actions,
				},
			},
		}
		p.Log.Trace().Str(logging.DATASETID, datasetID).Msg("Add subflow")
		subFlow := app.SubFlow{
			Name:     "",
			FlowType: app.CopyFlow,
			Triggers: []app.SubFlowTrigger{app.InitTrigger},
			Steps:    [][]app.DataFlowStep{steps},
		}
		subflows = append(subflows, subFlow)
		p.Log.Trace().Str(logging.DATASETID, datasetID).Msg("Adding copy module")
	}
	if readSelector != nil {
		p.Log.Trace().Str(logging.DATASETID, datasetID).Msg("Adding read path")
		var readAssetID string
		if sinkDataStore == nil {
			readAssetID = datasetID
		} else {
			readAssetID = datasetID + "-copy"
		}
		actions := createActionStructure(readSelector.Actions)
		if len(item.Configuration.ConfigDecisions[app.Read].Clusters) > 0 {
			readCluster = item.Configuration.ConfigDecisions[app.Read].Clusters[0]
		} else {
			msg := "Coud not determine the cluster for read"
			p.Log.Error().Str(logging.DATASETID, datasetID).Msg(msg)
			return errors.New(msg)
		}
		// The default capability scope is of type Asset
		scope := readSelector.ModuleCapability.Scope
		template := app.Template{
			Name: "read",
			Modules: []app.ModuleInfo{
				{
					Name:  readSelector.Module.Name,
					Type:  readSelector.Module.Spec.Type,
					Chart: readSelector.Module.Spec.Chart,
					Scope: scope,
				},
			},
		}
		templates = append(templates, template)
		api, err := moduleAPIToService(readSelector.ModuleCapability.API, readSelector.ModuleCapability.Scope,
			appContext, readSelector.Module.Name, readAssetID)
		if err != nil {
			return err
		}
		steps := []app.DataFlowStep{
			{
				Name:     "",
				Cluster:  readCluster,
				Template: "read",
				Parameters: &app.StepParameters{
					Source: &app.StepSource{
						AssetID: readAssetID,
						API:     nil,
					},
					API:     api,
					Actions: actions,
				},
			},
		}
		p.Log.Trace().Str(logging.DATASETID, datasetID).Msg("Add subflow")
		subFlow := app.SubFlow{
			Name:     "",
			FlowType: app.ReadFlow,
			Triggers: []app.SubFlowTrigger{app.WorkloadTrigger},
			Steps:    [][]app.DataFlowStep{steps},
		}

		subflows = append(subflows, subFlow)
	}

	// If everything finished without errors build the flow and add it to the plotter spec
	// Also add new assets as well as templates
	flowType := subflows[len(subflows)-1].FlowType
	flowName := item.Context.DataSetID + "-" + string(flowType)
	flow := app.Flow{
		Name:     flowName,
		FlowType: flowType,
		AssetID:  item.Context.DataSetID,
		SubFlows: subflows,
	}

	plotterSpec.Flows = append(plotterSpec.Flows, flow)
	for key, details := range assets {
		plotterSpec.Assets[key] = details
	}

	for _, template := range templates {
		plotterSpec.Templates[template.Name] = template
	}

	return nil
}

func moduleAPIToService(api *app.ModuleAPI, scope app.CapabilityScope, appContext *app.FybrikApplication, moduleName string, assetID string) (*app.Service, error) {
	hostnameTemplateString := api.Endpoint.Hostname
	if hostnameTemplateString == "" {
		hostnameTemplateString = "{{ .Release.Name }}.{{ .Release.Namespace }}"
	}
	hostnameTemplate, err := template.New("hostname").Funcs(sprig.TxtFuncMap()).Parse(hostnameTemplateString)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse hostname %s as a template", hostnameTemplateString)
	}

	instanceName := moduleName
	if scope == app.Asset {
		// if the scope of the module is asset then concat its id to the module name
		// to create the instance name.
		instanceName = utils.CreateStepName(moduleName, assetID)
	}
	releaseName := utils.GetReleaseName(appContext.Name, appContext.Namespace, instanceName)
	releaseNamespace := utils.GetDefaultModulesNamespace()

	type Release struct {
		Name      string `json:"Name"`
		Namespace string `json:"Namespace"`
	}

	type Values struct {
		Labels map[string]string `json:"labels,omitempty"`
	}

	type HostnameTemplateArgs struct {
		Release Release `json:"Release"`
		Values  Values  `json:"Values,omitempty"`
	}

	args := HostnameTemplateArgs{
		Release: Release{
			Name:      releaseName,
			Namespace: releaseNamespace,
		},
		Values: Values{
			Labels: appContext.Labels,
		},
	}

	// the following is required for proper types (e.g., labels must be map[string]interface{})
	values, err := utils.StructToMap(args)
	if err != nil {
		return nil, errors.Wrap(err, "could not serialize values for hostname field")
	}

	var hostname bytes.Buffer
	if err := hostnameTemplate.Execute(&hostname, values); err != nil {
		return nil, errors.Wrapf(err, "could not process template %s", hostnameTemplateString)
	}

	var service = &app.Service{
		Endpoint: app.EndpointSpec{
			Hostname: hostname.String(),
			Port:     api.Endpoint.Port,
			Scheme:   api.Endpoint.Scheme,
		},
		Format: api.DataFormat,
	}

	return service, nil
}
