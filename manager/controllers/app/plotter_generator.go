// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"strings"
	"text/template"

	"emperror.dev/errors"
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	modules "fybrik.io/fybrik/manager/controllers/app/modules"
	"fybrik.io/fybrik/manager/controllers/utils"
	connectors "fybrik.io/fybrik/pkg/connectors/clients"
	pb "fybrik.io/fybrik/pkg/connectors/protobuf"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/serde"
	"fybrik.io/fybrik/pkg/storage"
	vault "fybrik.io/fybrik/pkg/vault"
	"github.com/Masterminds/sprig/v3"
	"github.com/go-logr/logr"
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
	Client             client.Client
	Log                logr.Logger
	Modules            map[string]*app.FybrikModule
	Clusters           []multicluster.Cluster
	Owner              types.NamespacedName
	PolicyManager      connectors.PolicyManager
	Provision          storage.ProvisionInterface
	VaultConnection    vault.Interface
	ProvisionedStorage map[string]NewAssetInfo
}

// GetCopyDestination creates a Dataset for bucket allocation by implicit copies or ingest.
func (p *PlotterGenerator) GetCopyDestination(item modules.DataInfo, destinationInterface *app.InterfaceDetails, geo string) (*app.DataStore, error) {
	// provisioned storage for COPY
	originalAssetName := item.DataDetails.Name
	var bucket *storage.ProvisionedBucket
	var err error
	if bucket, err = AllocateBucket(p.Client, p.Log, p.Owner, originalAssetName, geo); err != nil {
		p.Log.Info("Bucket allocation failed: " + err.Error())
		return nil, err
	}
	bucketRef := &types.NamespacedName{Name: bucket.Name, Namespace: utils.GetSystemNamespace()}
	if err = p.Provision.CreateDataset(bucketRef, bucket, &p.Owner); err != nil {
		p.Log.Info("Dataset creation failed: " + err.Error())
		return nil, err
	}
	var endpoint string
	if strings.HasPrefix(bucket.Endpoint, "http://") {
		endpoint = bucket.Endpoint[7:]
	} else {
		endpoint = bucket.Endpoint
	}
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
			Metadata:   item.DataDetails.Metadata,
		}}
	p.ProvisionedStorage[item.Context.DataSetID] = assetInfo
	utils.PrintStructure(&assetInfo, p.Log, "ProvisionedStorage element")

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
func (p *PlotterGenerator) AddFlowInfoForAsset(item modules.DataInfo, appContext *app.FybrikApplication, plotterSpec *app.PlotterSpec) error {
	datasetID := item.Context.DataSetID
	p.Log.Info("Select modules for " + datasetID)

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

	selectors, err := p.SelectModules(item, appContext)
	if err != nil {
		return err
	}
	p.Log.V(0).Info("Generating a plotter")
	readSelector := selectors[app.Read]
	copySelector := selectors[app.Copy]
	// sanity
	if readSelector == nil && copySelector == nil {
		return errors.New("Can't generate a plotter - no modules available")
	}
	var copyCluster, readCluster string
	if copySelector != nil {
		p.Log.Info("Found copy module " + copySelector.GetModule().Name + " for " + datasetID)
		// copy should be applied - allocate storage
		if sinkDataStore, err = p.GetCopyDestination(item, copySelector.Destination, copySelector.StorageAccountRegion); err != nil {
			p.Log.Info("Allocation failed: " + err.Error())
			return err
		}
		var copyDataAssetID = datasetID + "-copy"
		actions := actionsToArbitrary(copySelector.Actions)
		if len(item.Configuration.ConfigDecisions[app.Copy].Clusters) > 0 {
			copyCluster = item.Configuration.ConfigDecisions[app.Copy].Clusters[0]
		} else {
			msg := "Coud not determine the cluster for copy"
			p.Log.Info(msg)
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
		p.Log.Info("Add subflow")
		subFlow := app.SubFlow{
			Name:     "",
			FlowType: app.CopyFlow,
			Triggers: []app.SubFlowTrigger{app.InitTrigger},
			Steps:    [][]app.DataFlowStep{steps},
		}
		subflows = append(subflows, subFlow)
		p.Log.Info("Adding copy module")
	}
	if readSelector != nil {
		p.Log.Info("Adding read path")
		var readAssetID string
		if sinkDataStore == nil {
			readAssetID = datasetID
		} else {
			readAssetID = datasetID + "-copy"
		}
		actions := actionsToArbitrary(readSelector.Actions)
		if len(item.Configuration.ConfigDecisions[app.Read].Clusters) > 0 {
			readCluster = item.Configuration.ConfigDecisions[app.Read].Clusters[0]
		} else {
			msg := "Coud not determine the cluster for read"
			p.Log.Info(msg)
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
		p.Log.Info("Add subflow")
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
	flowType := subflows[0].FlowType
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
	blueprintNamespace := utils.GetBlueprintNamespace()

	values := map[string]interface{}{
		"Release": map[string]interface{}{
			"Name":      releaseName,
			"Namespace": blueprintNamespace,
		},
		"Values": map[string]interface{}{
			"labels": appContext.Labels,
		},
	}

	// the following is required for proper types (e.g., labels must be map[string]interface{})
	values, err = utils.StructToMap(values)
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
