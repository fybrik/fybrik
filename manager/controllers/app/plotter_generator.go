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

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	managerUtils "fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/datapath"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/serde"
	"fybrik.io/fybrik/pkg/storage"
	sa "fybrik.io/fybrik/pkg/storage/apis/storageaccount/v1alpha1"
	"fybrik.io/fybrik/pkg/utils"
	"fybrik.io/fybrik/pkg/vault"
)

const (
	objectKeyHashLength  = 10
	bucketNameHashLength = 10
)

// NewAssetInfo points to the provisoned storage and hold information about the new asset
type NewAssetInfo struct {
	Storage *storage.ProvisionedBucket
	Details *fapp.DataStore
}

// PlotterGenerator constructs a plotter based on the requirements (governance actions, data location) and the existing set of FybrikModules
type PlotterGenerator struct {
	Client             client.Client
	Log                *zerolog.Logger
	Owner              types.NamespacedName
	Provision          storage.ProvisionInterface
	ProvisionedStorage map[string]NewAssetInfo
}

// AllocateStorage creates a Dataset for bucket allocation
func (p *PlotterGenerator) AllocateStorage(item *datapath.DataInfo, destinationInterface *taxonomy.Interface,
	account *sa.FybrikStorageAccountSpec) (*fapp.DataStore, error) {
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
		Endpoint:  account.Properties[storage.EndpointKey],
		SecretRef: types.NamespacedName{Name: account.SecretRef, Namespace: environment.GetSystemNamespace()},
		Geography: string(account.Geography),
	}
	bucketRef := &types.NamespacedName{Name: bucket.Name, Namespace: environment.GetSystemNamespace()}
	if err := p.Provision.CreateDataset(bucketRef, bucket, &p.Owner); err != nil {
		p.Log.Error().Err(err).Msg("Dataset creation failed")
		return nil, err
	}

	cType := managerUtils.GetDefaultConnectionType()
	connection := taxonomy.Connection{
		Name: cType,
		AdditionalProperties: serde.Properties{
			Items: map[string]interface{}{
				string(cType): map[string]interface{}{
					"endpoint":   bucket.Endpoint,
					"bucket":     bucket.Name,
					"object_key": genObjectKeyName,
				},
			},
		},
	}

	vaultSecretPath := vault.PathForReadingKubeSecret(bucket.SecretRef.Namespace, bucket.SecretRef.Name)
	vaultMap := make(map[string]fapp.Vault)
	if environment.IsVaultEnabled() {
		vaultMap[string(taxonomy.WriteFlow)] = fapp.Vault{
			SecretPath: vaultSecretPath,
			Role:       environment.GetModulesRole(),
			Address:    environment.GetVaultAddress(),
		}
		// The copied asset needs creds for later to be read
		vaultMap[string(taxonomy.ReadFlow)] = fapp.Vault{
			SecretPath: vaultSecretPath,
			Role:       environment.GetModulesRole(),
			Address:    environment.GetVaultAddress(),
		}
	} else {
		vaultMap[string(taxonomy.WriteFlow)] = fapp.Vault{}
		vaultMap[string(taxonomy.ReadFlow)] = fapp.Vault{}
	}
	datastore := &fapp.DataStore{
		Vault:      vaultMap,
		Connection: connection,
		Format:     destinationInterface.DataFormat,
	}
	assetInfo := NewAssetInfo{
		Storage: bucket,
		Details: datastore,
	}
	p.ProvisionedStorage[item.Context.DataSetID] = assetInfo
	logging.LogStructure("ProvisionedStorage element", assetInfo, p.Log, zerolog.DebugLevel, false, true)
	return datastore, nil
}

func (p *PlotterGenerator) getAssetDataStore(item *datapath.DataInfo) *fapp.DataStore {
	return &fapp.DataStore{
		Connection: item.DataDetails.Details.Connection,
		Vault:      getDatasetCredentials(item),
		Format:     item.DataDetails.Details.DataFormat,
	}
}

// store all available credentials in the plotter
// only relevant credentials will be sent to modules
func getDatasetCredentials(item *datapath.DataInfo) map[string]fapp.Vault {
	vaultMap := make(map[string]fapp.Vault)
	// credentials for read, write, delete
	// currently, one is used for all flows
	// TODO: store multiple secrets with credentials depending on the flow
	flows := []string{string(taxonomy.ReadFlow), string(taxonomy.WriteFlow), string(taxonomy.DeleteFlow)}
	for _, flow := range flows {
		if environment.IsVaultEnabled() {
			// Set the value received from the catalog connector.
			vaultSecretPath := item.DataDetails.Credentials
			vaultMap[flow] = fapp.Vault{
				SecretPath: vaultSecretPath,
				Role:       environment.GetModulesRole(),
				Address:    environment.GetVaultAddress(),
			}
		} else {
			vaultMap[flow] = fapp.Vault{}
		}
	}
	return vaultMap
}

func (p *PlotterGenerator) addTemplate(element *datapath.ResolvedEdge, plotterSpec *fapp.PlotterSpec, templateName string) {
	moduleCapability := element.Module.Spec.Capabilities[element.CapabilityIndex]
	template := fapp.Template{
		Name: templateName,
		Modules: []fapp.ModuleInfo{{
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
	steps []fapp.DataFlowStep, templateName string) []fapp.DataFlowStep {
	if steps == nil {
		steps = []fapp.DataFlowStep{}
	}
	var lastStepAPI *datacatalog.ResourceDetails
	if len(steps) > 0 {
		lastStepAPI = steps[len(steps)-1].Parameters.API
	}
	assetID := ""
	if lastStepAPI == nil {
		assetID = datasetID
	}
	steps = append(steps, fapp.DataFlowStep{
		Cluster:  element.Cluster,
		Template: templateName,
		Parameters: &fapp.StepParameters{
			Arguments: []*fapp.StepArgument{{
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
	steps []fapp.DataFlowStep, templateName string) []fapp.DataFlowStep {
	if steps == nil {
		steps = []fapp.DataFlowStep{}
	}
	steps = append(steps, fapp.DataFlowStep{
		Cluster:  element.Cluster,
		Template: templateName,
		Parameters: &fapp.StepParameters{
			Arguments: []*fapp.StepArgument{{AssetID: datasetID}, {AssetID: datasetID + "-copy"}},
			API:       api,
			Actions:   element.Actions,
		},
	})
	return steps
}

// getSupportedFormat returns the first dataformat supported by the module's capability sink interface
func (p *PlotterGenerator) getSupportedFormat(capability *fapp.ModuleCapability) taxonomy.DataFormat {
	for _, inter := range capability.SupportedInterfaces {
		if inter.Sink != nil {
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

	var sinkDataStore *fapp.DataStore
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
	element.Sink.Connection.DataFormat = p.getSupportedFormat(&capability)

	// allocate storage
	if sinkDataStore, err = p.AllocateStorage(item, element.Sink.Connection, &element.StorageAccount); err != nil {
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
			vault.PathForReadingKubeSecret(environment.GetSystemNamespace(), element.StorageAccount.SecretRef)

		item.DataDetails.Credentials = secretPath
	}
	item.DataDetails.Details.DataFormat = sinkDataStore.Format
	item.DataDetails.Details.Connection = sinkDataStore.Connection

	return nil
}

// Adds the asset details, flows and templates to the given plotter spec.
func (p *PlotterGenerator) AddFlowInfoForAsset(item *datapath.DataInfo, application *fapp.FybrikApplication,
	selection *datapath.Solution, plotterSpec *fapp.PlotterSpec) error {
	var err error
	p.Log.Trace().Str(logging.DATASETID, item.Context.DataSetID).Msg("Generating a plotter")
	datasetID := item.Context.DataSetID
	subflows := make([]fapp.SubFlow, 0)

	plotterSpec.Assets[item.Context.DataSetID] = fapp.AssetDetails{
		DataStore: *p.getAssetDataStore(item),
	}
	// DataStore for destination will be determined if an implicit copy is required
	var steps []fapp.DataFlowStep
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
			// allocate storage and create a temoprary asset
			var sinkDataStore *fapp.DataStore
			if sinkDataStore, err = p.AllocateStorage(item, element.Sink.Connection, &element.StorageAccount); err != nil {
				p.Log.Error().Err(err).Str(logging.DATASETID, item.Context.DataSetID).Msg("Storage allocation for copy failed")
				return err
			}
			steps = p.addStep(element, datasetID, api, steps, templateName)
			copyAssetID := steps[len(steps)-1].Parameters.Arguments[1].AssetID
			copyAsset := fapp.AssetDetails{
				AdvertisedAssetID: datasetID,
				DataStore:         *sinkDataStore,
			}
			plotterSpec.Assets[copyAssetID] = copyAsset
			datasetID = copyAssetID
			subflows = append(subflows, fapp.SubFlow{
				FlowType: taxonomy.CopyFlow,
				Triggers: []fapp.SubFlowTrigger{fapp.InitTrigger},
				Steps:    [][]fapp.DataFlowStep{steps},
			})

			// clear steps
			steps = nil
		} else {
			steps = p.addInMemoryStep(element, datasetID, api, steps, templateName)
		}
	}
	if steps != nil {
		subflows = append(subflows, fapp.SubFlow{
			FlowType: flowType,
			Triggers: []fapp.SubFlowTrigger{fapp.WorkloadTrigger},
			Steps:    [][]fapp.DataFlowStep{steps},
		})
	}
	// If everything finished without errors build the flow and add it to the plotter spec
	// Also add new assets as well as templates
	flowName := item.Context.DataSetID + "-" + string(flowType)
	flow := fapp.Flow{
		Name:     flowName,
		FlowType: flowType,
		AssetID:  item.Context.DataSetID,
		SubFlows: subflows,
	}
	plotterSpec.Flows = append(plotterSpec.Flows, flow)
	return nil
}

func moduleAPIToService(api *datacatalog.ResourceDetails, scope fapp.CapabilityScope, appContext *fapp.FybrikApplication,
	moduleName, assetID string) (*datacatalog.ResourceDetails, error) {
	instanceName := moduleName
	if scope == fapp.Asset {
		// if the scope of the module is asset then concat its id to the module name
		// to create the instance name.
		instanceName = managerUtils.CreateStepName(moduleName, assetID)
	}
	releaseName := managerUtils.GetReleaseName(appContext.Name, appContext.Namespace, instanceName)
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

func generateBucketName(owner types.NamespacedName, id string) string {
	name := owner.Name + "-" + owner.Namespace + utils.Hash(id, bucketNameHashLength)
	name = strings.ReplaceAll(name, ".", "-")
	return utils.K8sConformName(name)
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
