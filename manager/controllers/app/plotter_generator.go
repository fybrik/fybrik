// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"strings"
	tmpl "text/template"

	"emperror.dev/errors"
	"github.com/Masterminds/sprig/v3"
	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	dcclient "fybrik.io/fybrik/pkg/connectors/datacatalog/clients"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
	"fybrik.io/fybrik/pkg/storage"
	"fybrik.io/fybrik/pkg/vault"
)

const (
	objectKeyHashLength  = 10
	bucketNameHashLength = 10
)

// NewAssetInfo points to the provisoned storage and hold information about the new asset
type NewAssetInfo struct {
	Storage *storage.ProvisionedBucket
}

// PlotterGenerator constructs a plotter based on the requirements (governance actions, data location) and the existing set of FybrikModules
type PlotterGenerator struct {
	Client             client.Client
	Log                *zerolog.Logger
	Owner              types.NamespacedName
	Provision          storage.ProvisionInterface
	ProvisionedStorage map[string]NewAssetInfo
	DataCatalog        dcclient.DataCatalog
}

// AllocateStorage creates a Dataset for bucket allocation
func (p *PlotterGenerator) AllocateStorage(item *DataInfo, destinationInterface *taxonomy.Interface,
	account *app.FybrikStorageAccountSpec) (*app.DataStore, error) {
	// provisioned storage
	var genBucketName, genObjectKeyName string
	if item.DataDetails.ResourceMetadata.Name != "" {
		genObjectKeyName = item.DataDetails.ResourceMetadata.Name + utils.Hash(p.Owner.Name+p.Owner.Namespace, objectKeyHashLength)
	} else {
		genObjectKeyName = p.Owner.Name + utils.Hash(item.Context.DataSetID, objectKeyHashLength)
	}
	genBucketName = generateBucketName(p.Owner, item.Context.DataSetID)
	bucket := &storage.ProvisionedBucket{
		Name:      genBucketName,
		Endpoint:  account.Endpoint,
		SecretRef: types.NamespacedName{Name: account.SecretRef, Namespace: utils.GetSystemNamespace()},
	}
	bucketRef := &types.NamespacedName{Name: bucket.Name, Namespace: utils.GetSystemNamespace()}
	if err := p.Provision.CreateDataset(bucketRef, bucket, &p.Owner); err != nil {
		p.Log.Error().Err(err).Msg("Dataset creation failed")
		return nil, err
	}

	connection := taxonomy.Connection{
		Name: app.S3,
		AdditionalProperties: serde.Properties{
			Items: map[string]interface{}{
				string(app.S3): map[string]interface{}{
					"endpoint":   bucket.Endpoint,
					"bucket":     bucket.Name,
					"object_key": genObjectKeyName,
				},
			},
		},
	}

	assetInfo := NewAssetInfo{
		Storage: bucket,
	}
	p.ProvisionedStorage[item.Context.DataSetID] = assetInfo
	logging.LogStructure("ProvisionedStorage element", assetInfo, p.Log, zerolog.DebugLevel, false, true)

	vaultSecretPath := vault.PathForReadingKubeSecret(bucket.SecretRef.Namespace, bucket.SecretRef.Name)
	vaultMap := make(map[string]app.Vault)
	if utils.IsVaultEnabled() {
		vaultMap[string(taxonomy.WriteFlow)] = app.Vault{
			SecretPath: vaultSecretPath,
			Role:       utils.GetModulesRole(),
			Address:    utils.GetVaultAddress(),
		}
		// The copied asset needs creds for later to be read
		vaultMap[string(taxonomy.ReadFlow)] = app.Vault{
			SecretPath: vaultSecretPath,
			Role:       utils.GetModulesRole(),
			Address:    utils.GetVaultAddress(),
		}
	} else {
		vaultMap[string(taxonomy.WriteFlow)] = app.Vault{}
		vaultMap[string(taxonomy.ReadFlow)] = app.Vault{}
	}
	return &app.DataStore{
		Vault:      vaultMap,
		Connection: connection,
		Format:     destinationInterface.DataFormat,
	}, nil
}

func (p *PlotterGenerator) getAssetDataStore(item *DataInfo) *app.DataStore {
	return &app.DataStore{
		Connection: item.DataDetails.Details.Connection,
		Vault:      getDatasetCredentials(item),
		Format:     item.DataDetails.Details.DataFormat,
	}
}

// store all available credentials in the plotter
// only relevant credentials will be sent to modules
func getDatasetCredentials(item *DataInfo) map[string]app.Vault {
	vaultMap := make(map[string]app.Vault)
	// credentials for read, write, delete
	// currently, one is used for all flows
	// TODO: store multiple secrets with credentials depending on the flow
	flows := []string{string(taxonomy.ReadFlow), string(taxonomy.WriteFlow), string(taxonomy.DeleteFlow)}
	for _, flow := range flows {
		if utils.IsVaultEnabled() {
			// Set the value received from the catalog connector.
			vaultSecretPath := item.DataDetails.Credentials
			vaultMap[flow] = app.Vault{
				SecretPath: vaultSecretPath,
				Role:       utils.GetModulesRole(),
				Address:    utils.GetVaultAddress(),
			}
		} else {
			vaultMap[flow] = app.Vault{}
		}
	}
	return vaultMap
}

func (p *PlotterGenerator) addTemplate(element *ResolvedEdge, plotterSpec *app.PlotterSpec, templateName string) {
	moduleCapability := element.Module.Spec.Capabilities[element.CapabilityIndex]
	template := app.Template{
		Name: templateName,
		Modules: []app.ModuleInfo{{
			Name:       element.Module.Name,
			Type:       element.Module.Spec.Type,
			Chart:      element.Module.Spec.Chart,
			Scope:      moduleCapability.Scope,
			Capability: moduleCapability.Capability,
		}},
	}
	plotterSpec.Templates[template.Name] = template
}

func (p *PlotterGenerator) addInMemoryStep(element *ResolvedEdge, datasetID string, api *datacatalog.ResourceDetails,
	steps []app.DataFlowStep, templateName string) []app.DataFlowStep {
	if steps == nil {
		steps = []app.DataFlowStep{}
	}
	var lastStepAPI *datacatalog.ResourceDetails
	if len(steps) > 0 {
		lastStepAPI = steps[len(steps)-1].Parameters.API
	}
	assetID := ""
	if lastStepAPI == nil {
		assetID = datasetID
	}
	steps = append(steps, app.DataFlowStep{
		Cluster:  element.Cluster,
		Template: templateName,
		Parameters: &app.StepParameters{
			Arguments: []*app.StepArgument{{
				AssetID: assetID,
				API:     lastStepAPI,
			}},
			API:     api,
			Actions: element.Actions,
		},
	})
	return steps
}

func (p *PlotterGenerator) addStep(element *ResolvedEdge, datasetID string, api *datacatalog.ResourceDetails,
	steps []app.DataFlowStep, templateName string) []app.DataFlowStep {
	if steps == nil {
		steps = []app.DataFlowStep{}
	}
	steps = append(steps, app.DataFlowStep{
		Cluster:  element.Cluster,
		Template: templateName,
		Parameters: &app.StepParameters{
			Arguments: []*app.StepArgument{{AssetID: datasetID}, {AssetID: datasetID + "-copy"}},
			API:       api,
			Actions:   element.Actions,
		},
	})
	return steps
}

// Create a new asset and register it to the catalog. Used when the IsNewDataSet flag is true
func (p *PlotterGenerator) HandleNewAsset(item *DataInfo, application *app.FybrikApplication, selection *Solution,
	plotterSpec *app.PlotterSpec) error {
	var err error
	if item.DataDetails != nil && item.DataDetails.Details.DataFormat != "" {
		return nil
	}
	var sinkDataStore *app.DataStore
	var element *ResolvedEdge

	for _, element = range selection.DataPath {
		if element.StorageAccount.Region != "" {
			break
		}
	}
	// allocate storage
	if sinkDataStore, err = p.AllocateStorage(item, element.Sink.Connection, &element.StorageAccount); err != nil {
		p.Log.Error().Err(err).Str(logging.DATASETID, item.Context.DataSetID).Msg("Storage allocation failed")
		return err
	}

	// Reset StorageAccount to prevent re-allocation
	element.StorageAccount.Region = ""

	resourceMetadata := datacatalog.ResourceMetadata{
		Name:      item.Context.DataSetID,
		Geography: string(element.StorageAccount.Region),
	}
	// Update item with details of the asset
	// the asset will registered in the catalog later however
	// there are details that are already known like the asset
	// secret path
	item.DataDetails = &datacatalog.GetAssetResponse{
		ResourceMetadata: resourceMetadata,
	}
	if utils.IsVaultEnabled() {
		secretPath :=
			vault.PathForReadingKubeSecret(utils.GetSystemNamespace(), element.StorageAccount.SecretRef)

		item.DataDetails.Credentials = secretPath
	}
	item.DataDetails.Details.DataFormat = sinkDataStore.Format
	item.DataDetails.Details.Connection = sinkDataStore.Connection

	return nil
}

// Adds the asset details, flows and templates to the given plotter spec.
func (p *PlotterGenerator) AddFlowInfoForAsset(item *DataInfo, application *app.FybrikApplication, selection *Solution,
	plotterSpec *app.PlotterSpec) error {
	var err error
	p.Log.Trace().Str(logging.DATASETID, item.Context.DataSetID).Msg("Generating a plotter")
	datasetID := item.Context.DataSetID
	subflows := make([]app.SubFlow, 0)

	plotterSpec.Assets[item.Context.DataSetID] = app.AssetDetails{
		DataStore: *p.getAssetDataStore(item),
	}
	// DataStore for destination will be determined if an implicit copy is required
	var steps []app.DataFlowStep
	flowType := item.Context.Flow
	if flowType == "" {
		flowType = taxonomy.ReadFlow
	}
	for _, element := range selection.DataPath {
		moduleCapability := element.Module.Spec.Capabilities[element.CapabilityIndex]
		p.Log.Trace().Str(logging.DATASETID, item.Context.DataSetID).Msgf("Adding module %s for capability %s", element.Module.Name,
			moduleCapability.Capability)
		templateName := element.Module.Name + "-" + string(moduleCapability.Capability)
		p.addTemplate(element, plotterSpec, templateName)
		var api *datacatalog.ResourceDetails
		if moduleCapability.API != nil {
			if api, err = moduleAPIToService(moduleCapability.API, moduleCapability.Scope,
				application, element.Module.Name, datasetID); err != nil {
				return err
			}
		}
		if element.Sink != nil && !element.Sink.Virtual && element.StorageAccount.Region != "" {
			// allocate storage and create a temoprary asset
			var sinkDataStore *app.DataStore
			if sinkDataStore, err = p.AllocateStorage(item, element.Sink.Connection, &element.StorageAccount); err != nil {
				p.Log.Error().Err(err).Str(logging.DATASETID, item.Context.DataSetID).Msg("Storage allocation for copy failed")
				return err
			}
			steps = p.addStep(element, datasetID, api, steps, templateName)
			copyAssetID := steps[len(steps)-1].Parameters.Arguments[1].AssetID
			copyAsset := app.AssetDetails{
				AdvertisedAssetID: datasetID,
				DataStore:         *sinkDataStore,
			}
			plotterSpec.Assets[copyAssetID] = copyAsset
			datasetID = copyAssetID
			subflows = append(subflows, app.SubFlow{
				FlowType: taxonomy.CopyFlow,
				Triggers: []app.SubFlowTrigger{app.InitTrigger},
				Steps:    [][]app.DataFlowStep{steps},
			})
			// clear steps
			steps = nil
		} else {
			steps = p.addInMemoryStep(element, datasetID, api, steps, templateName)
		}
	}
	if steps != nil {
		subflows = append(subflows, app.SubFlow{
			FlowType: flowType,
			Triggers: []app.SubFlowTrigger{app.WorkloadTrigger},
			Steps:    [][]app.DataFlowStep{steps},
		})
	}
	// If everything finished without errors build the flow and add it to the plotter spec
	// Also add new assets as well as templates
	flowName := item.Context.DataSetID + "-" + string(flowType)
	flow := app.Flow{
		Name:     flowName,
		FlowType: flowType,
		AssetID:  item.Context.DataSetID,
		SubFlows: subflows,
	}
	plotterSpec.Flows = append(plotterSpec.Flows, flow)
	return nil
}

func moduleAPIToService(api *datacatalog.ResourceDetails, scope app.CapabilityScope, appContext *app.FybrikApplication,
	moduleName, assetID string) (*datacatalog.ResourceDetails, error) {
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

	type APITemplateArgs struct {
		Release Release `json:"Release"`
		Values  Values  `json:"Values,omitempty"`
	}

	args := APITemplateArgs{
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
		return nil, errors.Wrap(err, "could not serialize values")
	}
	newConnection := taxonomy.Connection{Name: api.Connection.Name,
		AdditionalProperties: serde.Properties{Items: make(map[string]interface{})}}
	newProps := make(map[string]interface{})
	props := api.Connection.AdditionalProperties.Items[string(api.Connection.Name)].(map[string]interface{})
	for key, val := range props {
		if templateStr, ok := val.(string); ok {
			fieldTemplate, err := tmpl.New(key).Funcs(sprig.TxtFuncMap()).Parse(templateStr)
			if err != nil {
				return nil, errors.Wrapf(err, "could not parse %s as a template", templateStr)
			}
			var newValue bytes.Buffer
			if err = fieldTemplate.Execute(&newValue, values); err != nil {
				return nil, errors.Wrapf(err, "could not process template %s", templateStr)
			}
			newProps[key] = newValue.String()
		} else {
			newProps[key] = val
		}
	}
	newConnection.AdditionalProperties.Items[string(api.Connection.Name)] = newProps
	var service = &datacatalog.ResourceDetails{
		Connection: newConnection,
		DataFormat: api.DataFormat,
	}
	return service, nil
}

func generateBucketName(owner types.NamespacedName, id string) string {
	name := owner.Name + "-" + owner.Namespace + utils.Hash(id, bucketNameHashLength)
	name = strings.ReplaceAll(name, ".", "-")
	return utils.K8sConformName(name)
}
