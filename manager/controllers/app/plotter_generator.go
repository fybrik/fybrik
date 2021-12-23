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
	pmclient "fybrik.io/fybrik/pkg/connectors/policymanager/clients"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
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
}

// PlotterGenerator constructs a plotter based on the requirements (governance actions, data location) and the existing set of FybrikModules
type PlotterGenerator struct {
	Client                client.Client
	Log                   zerolog.Logger
	Modules               map[string]*app.FybrikModule
	Clusters              []multicluster.Cluster
	Owner                 types.NamespacedName
	PolicyManager         pmclient.PolicyManager
	Provision             storage.ProvisionInterface
	VaultConnection       vault.Interface
	ProvisionedStorage    map[string]NewAssetInfo
	StorageAccountRegions []string
}

// GetCopyDestination creates a Dataset for bucket allocation by implicit copies or ingest.
func (p *PlotterGenerator) GetCopyDestination(item DataInfo, destinationInterface *app.InterfaceDetails, geo string) (*app.DataStore, error) {
	// provisioned storage for COPY
	originalAssetName := item.DataDetails.Metadata.Name
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

	connection := taxonomy.Connection{
		Name: "s3",
		AdditionalProperties: serde.Properties{
			Items: map[string]interface{}{
				"s3": map[string]interface{}{
					"endpoint":   url.Host,
					"bucket":     bucket.Name,
					"object_key": originalAssetName + utils.Hash(p.Owner.Name+p.Owner.Namespace, 10),
				},
			},
		},
	}

	assetInfo := NewAssetInfo{
		Storage: bucket,
	}
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
		Format:     string(destinationInterface.DataFormat),
	}, nil
}

// Adds the asset details, flows and templates to the given plotter spec.
func (p *PlotterGenerator) AddFlowInfoForAsset(item DataInfo, appContext *app.FybrikApplication, plotterSpec *app.PlotterSpec) error {
	p.Log.Trace().Str(logging.DATASETID, item.Context.DataSetID).Msg("Choose modules for dataset")
	var err error
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
		Connection: item.DataDetails.Connection,
		Vault:      vaultMap,
		Format:     string(item.DataDetails.Interface.DataFormat),
	}

	assets[item.Context.DataSetID] = app.AssetDetails{
		AdvertisedAssetID: "",
		DataStore:         *sourceDataStore,
	}

	// DataStore for destination will be determined if an implicit copy is required
	var sinkDataStore *app.DataStore

	solutions := p.FindPaths(&item, appContext)
	if len(solutions) == 0 {
		return errors.New("Data path could not be constructed")
	}
	p.Log.Trace().Msg("Generating a plotter")
	selection := solutions[0]
	datasetID := item.Context.DataSetID
	for _, element := range selection.DataPath {
		moduleCapability := element.Module.Spec.Capabilities[element.CapabilityIndex]
		p.Log.Trace().Msg("Adding module for " + moduleCapability.Capability)
		actions := element.Actions
		template := app.Template{
			Name: moduleCapability.Capability,
			Modules: []app.ModuleInfo{{
				Name:  element.Module.Name,
				Type:  element.Module.Spec.Type,
				Chart: element.Module.Spec.Chart,
				Scope: moduleCapability.Scope,
			}},
		}
		templates = append(templates, template)
		var api *app.Service
		if moduleCapability.API != nil {
			api, err = moduleAPIToService(moduleCapability.API, moduleCapability.Scope,
				appContext, element.Module.Name, datasetID)
			if err != nil {
				return err
			}
		}
		var subFlow app.SubFlow
		if !element.Sink.Virtual {
			// allocate storage and create a temoprary asset
			if sinkDataStore, err = p.GetCopyDestination(item, element.Sink.Connection, element.StorageAccountRegion); err != nil {
				p.Log.Error().Err(err).Msg("Storage allocation failed")
				return err
			}
			copyAssetID := datasetID + "-copy"
			copyAsset := app.AssetDetails{
				AdvertisedAssetID: datasetID,
				DataStore:         *sinkDataStore,
			}
			assets[copyAssetID] = copyAsset
			steps := []app.DataFlowStep{
				{
					Name:     "",
					Cluster:  element.Cluster,
					Template: moduleCapability.Capability,
					Parameters: &app.StepParameters{
						Source: &app.StepSource{
							AssetID: datasetID,
							API:     nil,
						},
						Sink: &app.StepSink{
							AssetID: copyAssetID,
						},
						API:     api,
						Actions: actions,
					},
				},
			}
			datasetID = copyAssetID
			subFlow = app.SubFlow{
				Name:     "",
				FlowType: app.CopyFlow,
				Triggers: []app.SubFlowTrigger{app.InitTrigger},
				Steps:    [][]app.DataFlowStep{steps},
			}
		} else {
			steps := []app.DataFlowStep{
				{
					Name:     "",
					Cluster:  element.Cluster,
					Template: moduleCapability.Capability,
					Parameters: &app.StepParameters{
						Source: &app.StepSource{
							AssetID: datasetID,
							API:     nil,
						},
						API:     api,
						Actions: actions,
					},
				},
			}
			subFlow = app.SubFlow{
				Name:     "",
				FlowType: app.ReadFlow,
				Triggers: []app.SubFlowTrigger{app.WorkloadTrigger},
				Steps:    [][]app.DataFlowStep{steps},
			}
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
		Format: string(api.DataFormat),
	}

	return service, nil
}
