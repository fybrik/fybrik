// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	tmpl "text/template"

	"emperror.dev/errors"
	"github.com/Masterminds/sprig/v3"
	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fappv1 "fybrik.io/fybrik/manager/apis/app/v1beta1"
	fappv2 "fybrik.io/fybrik/manager/apis/app/v1beta2"
	managerUtils "fybrik.io/fybrik/manager/controllers/utils"
	storage "fybrik.io/fybrik/pkg/connectors/storagemanager/clients"
	"fybrik.io/fybrik/pkg/datapath"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/storagemanager"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
	"fybrik.io/fybrik/pkg/utils"
	"fybrik.io/fybrik/pkg/vault"
)

// NewAssetInfo points to the provisioned storage and holds information about the new asset
type NewAssetInfo struct {
	StorageAccount *fappv2.FybrikStorageAccountSpec
	Details        *fappv1.DataStore
	Persistent     bool
}

// PlotterGenerator constructs a plotter based on the requirements (governance actions, data location) and the existing set of FybrikModules
type PlotterGenerator struct {
	Client             client.Client
	Log                *zerolog.Logger
	UUID               string
	Owner              types.NamespacedName
	StorageManager     storage.StorageManagerInterface
	ProvisionedStorage map[string]NewAssetInfo
}

// Provision allocates storage based on the selected account and generates the destination data store for the plotter
func (p *PlotterGenerator) Provision(item *datapath.DataInfo, destinationInterface *taxonomy.Interface,
	account *fappv2.FybrikStorageAccountSpec) (*fappv1.DataStore, error) {
	// provisioned storage
	secretRef := &taxonomy.SecretRef{Name: account.SecretRef, Namespace: environment.GetAdminCRsNamespace()}
	allocateRequest := &storagemanager.AllocateStorageRequest{
		AccountType:       account.Type,
		AccountProperties: taxonomy.StorageAccountProperties{Properties: account.AdditionalProperties},
		Secret:            *secretRef,
		Opts: storagemanager.Options{
			AppDetails:        storagemanager.ApplicationDetails{Name: p.Owner.Name, Namespace: p.Owner.Namespace, UUID: p.UUID},
			DatasetProperties: storagemanager.DatasetDetails{Name: item.Context.DataSetID},
			ConfigurationOpts: storagemanager.ConfigOptions{},
		},
	}
	response, err := p.StorageManager.AllocateStorage(allocateRequest)
	if err != nil {
		return nil, err
	}

	vaultSecretPath := vault.PathForReadingKubeSecret(secretRef.Namespace, secretRef.Name)
	vaultMap := make(map[string]fappv1.Vault)
	if environment.IsVaultEnabled() {
		vaultMap[string(taxonomy.WriteFlow)] = fappv1.Vault{
			SecretPath: vaultSecretPath,
			Role:       environment.GetModulesRole(),
			Address:    environment.GetVaultAddress(),
		}
		// The copied asset needs creds for later to be read
		vaultMap[string(taxonomy.ReadFlow)] = fappv1.Vault{
			SecretPath: vaultSecretPath,
			Role:       environment.GetModulesRole(),
			Address:    environment.GetVaultAddress(),
		}
	} else {
		vaultMap[string(taxonomy.WriteFlow)] = fappv1.Vault{}
		vaultMap[string(taxonomy.ReadFlow)] = fappv1.Vault{}
	}
	datastore := &fappv1.DataStore{
		Vault:      vaultMap,
		Connection: *response.Connection,
		Format:     destinationInterface.DataFormat,
	}
	assetInfo := NewAssetInfo{
		StorageAccount: account,
		Details:        datastore,
	}
	p.ProvisionedStorage[item.Context.DataSetID] = assetInfo
	logging.LogStructure("ProvisionedStorage element", assetInfo, p.Log, zerolog.DebugLevel, false, true)
	return datastore, nil
}

func (p *PlotterGenerator) getAssetDataStore(item *datapath.DataInfo) *fappv1.DataStore {
	return &fappv1.DataStore{
		Connection: item.DataDetails.Details.Connection,
		Vault:      getDatasetCredentials(item),
		Format:     item.DataDetails.Details.DataFormat,
	}
}

// store all available credentials in the plotter
// only relevant credentials will be sent to modules
func getDatasetCredentials(item *datapath.DataInfo) map[string]fappv1.Vault {
	vaultMap := make(map[string]fappv1.Vault)
	// credentials for read, write, delete
	// currently, one is used for all flows
	// TODO: store multiple secrets with credentials depending on the flow
	flows := []string{string(taxonomy.ReadFlow), string(taxonomy.WriteFlow), string(taxonomy.DeleteFlow)}
	for _, flow := range flows {
		if environment.IsVaultEnabled() {
			// Set the value received from the catalog connector.
			vaultSecretPath := item.DataDetails.Credentials
			vaultMap[flow] = fappv1.Vault{
				SecretPath: vaultSecretPath,
				Role:       environment.GetModulesRole(),
				Address:    environment.GetVaultAddress(),
			}
		} else {
			vaultMap[flow] = fappv1.Vault{}
		}
	}
	return vaultMap
}

func (p *PlotterGenerator) addTemplate(element *datapath.ResolvedEdge, plotterSpec *fappv1.PlotterSpec, templateName string) {
	moduleCapability := element.Module.Spec.Capabilities[element.CapabilityIndex]
	template := fappv1.Template{
		Name: templateName,
		Modules: []fappv1.ModuleInfo{{
			Name:       element.Module.Name,
			Type:       element.Module.Spec.Type,
			Chart:      element.Module.Spec.Chart,
			Scope:      moduleCapability.Scope,
			Capability: moduleCapability.Capability,
		}},
	}
	plotterSpec.Templates[template.Name] = template
}

func (p *PlotterGenerator) addInMemoryStep(element *datapath.ResolvedEdge, datasetID string, api *datacatalog.ResourceDetails,
	steps []fappv1.DataFlowStep, templateName string) []fappv1.DataFlowStep {
	if steps == nil {
		steps = []fappv1.DataFlowStep{}
	}
	var lastStepAPI *datacatalog.ResourceDetails
	if len(steps) > 0 {
		lastStepAPI = steps[len(steps)-1].Parameters.API
	}
	assetID := ""
	if lastStepAPI == nil {
		assetID = datasetID
	}
	steps = append(steps, fappv1.DataFlowStep{
		Cluster:  element.Cluster,
		Template: templateName,
		Parameters: &fappv1.StepParameters{
			Arguments: []*fappv1.StepArgument{{
				AssetID: assetID,
				API:     lastStepAPI,
			}},
			API:     api,
			Actions: element.Actions,
		},
	})
	return steps
}

func (p *PlotterGenerator) addStep(element *datapath.ResolvedEdge, datasetID string, api *datacatalog.ResourceDetails,
	steps []fappv1.DataFlowStep, templateName string) []fappv1.DataFlowStep {
	if steps == nil {
		steps = []fappv1.DataFlowStep{}
	}
	steps = append(steps, fappv1.DataFlowStep{
		Cluster:  element.Cluster,
		Template: templateName,
		Parameters: &fappv1.StepParameters{
			Arguments: []*fappv1.StepArgument{{AssetID: datasetID}, {AssetID: datasetID + "-copy"}},
			API:       api,
			Actions:   element.Actions,
		},
	})
	return steps
}

// getSupportedFormat returns the first dataformat supported by the module's capability sink interface that matches the protocol
func (p *PlotterGenerator) getSupportedFormat(capability *fappv1.ModuleCapability, protocol taxonomy.ConnectionType) taxonomy.DataFormat {
	for _, inter := range capability.SupportedInterfaces {
		if inter.Sink != nil && inter.Sink.Protocol == protocol {
			return inter.Sink.DataFormat
		}
	}
	return ""
}

// Handle a new asset: allocate storage and update its metadata. Used when the
// IsNewDataSet flag is true.
func (p *PlotterGenerator) handleNewAsset(item *datapath.DataInfo, selection *datapath.Solution) error {
	var err error
	if item.DataDetails != nil && item.DataDetails.Details.DataFormat != "" {
		return nil
	}
	p.Log.Trace().Str(logging.DATASETID, item.Context.DataSetID).Msg("Handle new dataset")

	var sinkDataStore *fappv1.DataStore
	var element *datapath.ResolvedEdge

	needToAllocateStorage := false
	for _, element = range selection.DataPath {
		if element.StorageAccount.Geography != "" {
			needToAllocateStorage = true
			break
		}
	}
	if !needToAllocateStorage {
		return nil
	}

	// Fill in the empty dataFormat in the sink node
	capability := element.Module.Spec.Capabilities[element.CapabilityIndex]
	element.Sink.Connection.DataFormat = p.getSupportedFormat(&capability, element.StorageAccount.Type)

	// allocate storage
	if sinkDataStore, err = p.Provision(item, element.Sink.Connection, &element.StorageAccount); err != nil {
		p.Log.Error().Err(err).Str(logging.DATASETID, item.Context.DataSetID).Msg("Storage allocation failed")
		return err
	}

	resourceMetadata := datacatalog.ResourceMetadata{
		Name:      item.Context.DataSetID,
		Geography: string(element.StorageAccount.Geography),
	}

	// Reset StorageAccount to prevent re-allocation
	element.StorageAccount.Geography = ""

	// Update item with details of the asset
	// the asset will registered in the catalog later however
	// there are details that are already known like the asset
	// secret path
	item.DataDetails = &datacatalog.GetAssetResponse{
		ResourceMetadata: resourceMetadata,
	}
	if environment.IsVaultEnabled() {
		secretPath :=
			vault.PathForReadingKubeSecret(environment.GetAdminCRsNamespace(), element.StorageAccount.SecretRef)

		item.DataDetails.Credentials = secretPath
	}
	item.DataDetails.Details.DataFormat = sinkDataStore.Format
	item.DataDetails.Details.Connection = sinkDataStore.Connection

	return nil
}

// Adds the asset details, flows and templates to the given plotter spec.
func (p *PlotterGenerator) AddFlowInfoForAsset(item *datapath.DataInfo, application *fappv1.FybrikApplication,
	selection *datapath.Solution, plotterSpec *fappv1.PlotterSpec) error {
	var err error
	p.Log.Trace().Str(logging.DATASETID, item.Context.DataSetID).Msg("Generating a plotter")
	datasetID := item.Context.DataSetID
	subflows := make([]fappv1.SubFlow, 0)

	plotterSpec.Assets[item.Context.DataSetID] = fappv1.AssetDetails{
		DataStore: *p.getAssetDataStore(item),
	}
	// DataStore for destination will be determined if an implicit copy is required
	var steps []fappv1.DataFlowStep
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
		if element.Sink != nil && !element.Sink.Virtual && element.StorageAccount.Geography != "" {
			// allocate storage and create a temporary asset
			var sinkDataStore *fappv1.DataStore
			if sinkDataStore, err = p.Provision(item, element.Sink.Connection, &element.StorageAccount); err != nil {
				p.Log.Error().Err(err).Str(logging.DATASETID, item.Context.DataSetID).Msg("Storage allocation for copy failed")
				return err
			}
			steps = p.addStep(element, datasetID, api, steps, templateName)
			copyAssetID := steps[len(steps)-1].Parameters.Arguments[1].AssetID
			copyAsset := fappv1.AssetDetails{
				AdvertisedAssetID: datasetID,
				DataStore:         *sinkDataStore,
			}
			plotterSpec.Assets[copyAssetID] = copyAsset
			datasetID = copyAssetID
			subflows = append(subflows, fappv1.SubFlow{
				FlowType: taxonomy.CopyFlow,
				Triggers: []fappv1.SubFlowTrigger{fappv1.InitTrigger},
				Steps:    [][]fappv1.DataFlowStep{steps},
			})

			// clear steps
			steps = nil
		} else {
			steps = p.addInMemoryStep(element, datasetID, api, steps, templateName)
		}
	}
	if steps != nil {
		subflows = append(subflows, fappv1.SubFlow{
			FlowType: flowType,
			Triggers: []fappv1.SubFlowTrigger{fappv1.WorkloadTrigger},
			Steps:    [][]fappv1.DataFlowStep{steps},
		})
	}
	// If everything finished without errors build the flow and add it to the plotter spec
	// Also add new assets as well as templates
	flowName := item.Context.DataSetID + "-" + string(flowType)
	flow := fappv1.Flow{
		Name:     flowName,
		FlowType: flowType,
		AssetID:  item.Context.DataSetID,
		SubFlows: subflows,
	}
	plotterSpec.Flows = append(plotterSpec.Flows, flow)
	return nil
}

func moduleAPIToService(api *datacatalog.ResourceDetails, scope fappv1.CapabilityScope, appContext *fappv1.FybrikApplication,
	moduleName, assetID string) (*datacatalog.ResourceDetails, error) {
	instanceName := moduleName
	if scope == fappv1.Asset {
		// if the scope of the module is asset then concat its id to the module name
		// to create the instance name.
		instanceName = managerUtils.CreateStepName(moduleName, assetID)
	}
	releaseName := managerUtils.GetReleaseName(appContext.Name, string(appContext.UID), instanceName)
	releaseNamespace := environment.GetDefaultModulesNamespace()

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
	// The connection may have some of the string fields templated
	// These templates should be resolved using APITemplateArgs
	newConnection := taxonomy.Connection{Name: api.Connection.Name,
		AdditionalProperties: serde.Properties{Items: make(map[string]interface{})}}
	for key, val := range api.Connection.AdditionalProperties.Items {
		newVal, err := resolveTemplates(val, key, values)
		if err != nil {
			return nil, err
		}
		newConnection.AdditionalProperties.Items[key] = newVal
	}
	var service = &datacatalog.ResourceDetails{
		Connection: newConnection,
		DataFormat: api.DataFormat,
	}
	return service, nil
}

// resolve string fields that are templated using the values map
func resolveTemplates(val interface{}, key string, values map[string]interface{}) (interface{}, error) {
	if s, ok := val.(string); ok {
		return resolveString(s, key, values)
	}
	if m, err := utils.StructToMap(val); err == nil {
		newMap := make(map[string]interface{}, 0)
		for k, v := range m {
			var newVal interface{}
			if newVal, err = resolveTemplates(v, key+"."+k, values); err != nil {
				return nil, err
			}
			newMap[k] = newVal
		}
		return newMap, nil
	}
	return val, nil
}

// substitute templates with their actual values
func resolveString(templateStr, key string, values map[string]interface{}) (interface{}, error) {
	fieldTemplate, err := tmpl.New(key).Funcs(sprig.TxtFuncMap()).Parse(templateStr)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse %s as a template", templateStr)
	}
	var newValue bytes.Buffer
	if err = fieldTemplate.Execute(&newValue, values); err != nil {
		return nil, errors.Wrapf(err, "could not process template %s", templateStr)
	}
	return newValue.String(), nil
}
