// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"strings"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	modules "github.com/mesh-for-data/mesh-for-data/manager/controllers/app/modules"
	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	connectors "github.com/mesh-for-data/mesh-for-data/pkg/connectors/clients"
	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
	"github.com/mesh-for-data/mesh-for-data/pkg/multicluster"
	local "github.com/mesh-for-data/mesh-for-data/pkg/multicluster/local"
	"github.com/mesh-for-data/mesh-for-data/pkg/serde"
	"github.com/mesh-for-data/mesh-for-data/pkg/storage"
	vault "github.com/mesh-for-data/mesh-for-data/pkg/vault"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewAssetInfo points to the provisoned storage and hold information about the new asset
type NewAssetInfo struct {
	Storage *storage.ProvisionedBucket
	Details *pb.DatasetDetails
}

// ModuleManager builds a set of modules based on the requirements (governance actions, data location) and the existing set of M4DModules
type ModuleManager struct {
	Client             client.Client
	Log                logr.Logger
	Modules            map[string]*app.M4DModule
	Clusters           []multicluster.Cluster
	Owner              types.NamespacedName
	PolicyManager      connectors.PolicyManager
	WorkloadGeography  string
	Provision          storage.ProvisionInterface
	VaultConnection    vault.Interface
	ProvisionedStorage map[string]NewAssetInfo
}

// SelectModuleInstances builds a list of required modules with the relevant arguments
/*

Future (ingest & write) order of the lookup to support ingest is:
- If no label selector assume ingest of external data (what about archive in future?)
	- run Copy module close to destination (determined based on governance decisions)
	- and register new data set in data catalog
- If Data Context Flow=Write
   - Write is always required, and always close to compute
   - Implicit Copy is used on demand, e.g. if a write module does not support the existing source of data or governance actions
   - Transformations are always done at workload location
   - If not external data, then register in data catalog

Updates to add ingest:
- If no label selector assume ingest of external data
	- run Copy module close to destination (determined based on governance decisions)
	- and register new data set in data catalog
- Otherwise assume workload wants to read from cataloged data
   - Read is always required.
   - Copy is used on demand, e.g. if a read module does not support the existing source of data or actions
   - Transformations are always done at data source location
   - Read module runs close to compute (in processing geography)
   - Write module has not yet been implemented - will be implemented in future release
   - All data sets are processed, even if an error is encountered in one or more, to provide a complete status at the end of the reconcile
   - Dependencies are checked but not added yet to the blueprint
*/

// GetCopyDestination creates a Dataset for bucket allocation by implicit copies or ingest.
func (m *ModuleManager) GetCopyDestination(item modules.DataInfo, destinationInterface *app.InterfaceDetails, geo string) (*app.DataStore, error) {
	// provisioned storage for COPY
	originalAssetName := item.DataDetails.Name
	var bucket *storage.ProvisionedBucket
	var err error
	if bucket, err = AllocateBucket(m.Client, m.Log, m.Owner, originalAssetName, geo); err != nil {
		m.Log.Info("Bucket allocation failed: " + err.Error())
		return nil, err
	}
	bucketRef := &types.NamespacedName{Name: bucket.Name, Namespace: utils.GetSystemNamespace()}
	if err = m.Provision.CreateDataset(bucketRef, bucket, &m.Owner); err != nil {
		m.Log.Info("Dataset creation failed: " + err.Error())
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
			ObjectKey: originalAssetName + utils.Hash(m.Owner.Name+m.Owner.Namespace, 10),
		},
	}
	connection := serde.NewArbitrary(datastore)
	assetInfo := NewAssetInfo{
		Storage: bucket,
		Details: &pb.DatasetDetails{
			Name:       originalAssetName,
			Geo:        item.DataDetails.Geography,
			DataFormat: string(destinationInterface.DataFormat),
			DataStore:  datastore,
			Metadata:   item.DataDetails.Metadata,
		}}
	m.ProvisionedStorage[item.Context.DataSetID] = assetInfo
	utils.PrintStructure(&assetInfo, m.Log, "ProvisionedStorage element")

	vaultSecretPath := vault.PathForReadingKubeSecret(bucket.SecretRef.Namespace, bucket.SecretRef.Name)
	return &app.DataStore{
		Vault: app.Vault{
			SecretPath: vaultSecretPath,
			Role:       utils.GetModulesRole(),
			Address:    utils.GetVaultAddress(),
		},
		Connection: *connection,
		Format:     string(destinationInterface.DataFormat),
	}, nil
}

func (m *ModuleManager) selectReadModule(item modules.DataInfo, appContext *app.M4DApplication) (*modules.Selector, error) {
	// read module is required if the workload exists
	if appContext.Spec.Selector.WorkloadSelector.Size() == 0 {
		return nil, nil
	}
	m.Log.Info("Select read path for " + item.Context.DataSetID)

	// Read policies for data that is processed in the workload geography
	var readActions []*pb.EnforcementAction
	var err error
	readActions, err = LookupPolicyDecisions(item.Context.DataSetID, m.PolicyManager, appContext,
		&pb.AccessOperation{Type: pb.AccessOperation_READ, Destination: m.WorkloadGeography})
	if err != nil {
		return nil, err
	}
	// select a read module that supports user interface requirements
	// actions are not checked since they are not necessarily done by the read module
	readSelector := &modules.Selector{Flow: app.Read,
		Destination:  &item.Context.Requirements.Interface,
		Actions:      []*pb.EnforcementAction{},
		Source:       nil,
		Dependencies: []*app.M4DModule{},
		Module:       nil,
		Message:      "",
		Geo:          m.WorkloadGeography,
	}
	if !readSelector.SelectModule(m.Modules) {
		m.Log.Info(readSelector.GetError())
		return nil, errors.New(readSelector.GetError())
	}
	readSelector.Actions = readActions
	return readSelector, nil
}

func (m *ModuleManager) selectCopyModule(item modules.DataInfo, appContext *app.M4DApplication, readSelector *modules.Selector) (*modules.Selector, error) {
	// logic for deciding whether copy module is required
	var interfaces []*app.InterfaceDetails
	var copyRequired bool
	additionalActions := []*pb.EnforcementAction{}
	if readSelector != nil {
		copyRequired, interfaces, additionalActions = m.getCopyRequirements(item, readSelector)
	} else if item.Context.Requirements.Copy.Required {
		copyRequired = true
		interfaces = []*app.InterfaceDetails{&item.Context.Requirements.Interface}
	}
	if !copyRequired {
		return nil, nil
	}
	actionsOnCopy := []*pb.EnforcementAction{}
	geo := m.WorkloadGeography
	// WRITE actions
	if readSelector == nil {
		var err error
		actionsOnCopy, geo, err = m.enforceWritePolicies(appContext, item.Context.DataSetID)
		if err != nil {
			return nil, err
		}
	}
	actionsOnCopy = append(actionsOnCopy, additionalActions...)
	m.Log.Info("Copy is required for " + item.Context.DataSetID)
	var copySelector *modules.Selector
	// select a module that supports COPY, supports required governance actions, has the required dependencies, with source in module sources and a non-empty intersection between requested and supported interfaces.
	for _, copyDest := range interfaces {
		copySelector = &modules.Selector{
			Flow:         app.Copy,
			Source:       &item.DataDetails.Interface,
			Actions:      actionsOnCopy,
			Destination:  copyDest,
			Dependencies: make([]*app.M4DModule, 0),
			Module:       nil,
			Geo:          geo,
			Message:      ""}

		if copySelector.SelectModule(m.Modules) {
			break
		}
	}
	if copySelector == nil {
		return nil, errors.New("no copy module has been found supporting required source interface")
	}
	if copySelector.GetModule() == nil {
		m.Log.Info("Could not find copy module for " + item.Context.DataSetID)
		return nil, errors.New(copySelector.GetError())
	}
	return copySelector, nil
}

// SelectModuleInstances selects the necessary read/copy/write modules for the blueprint for a given data set
// Write path is not yet implemented
func (m *ModuleManager) SelectModuleInstances(item modules.DataInfo, appContext *app.M4DApplication) ([]modules.ModuleInstanceSpec, error) {
	datasetID := item.Context.DataSetID
	m.Log.Info("Select modules for " + datasetID)
	instances := make([]modules.ModuleInstanceSpec, 0)
	var err error
	if m.WorkloadGeography, err = m.GetProcessingGeography(appContext); err != nil {
		m.Log.Info("Could not determine the workload geography")
		return nil, err
	}

	// Set the value received from the catalog connector.
	vaultSecretPath := item.VaultSecretPath

	// Each selector receives source/sink interface and relevant actions
	// Starting with the data location interface for source and the required interface for sink
	sourceDataStore := &app.DataStore{
		Connection: item.DataDetails.Connection,
		Vault: app.Vault{
			SecretPath: vaultSecretPath,
			Role:       utils.GetModulesRole(),
			Address:    utils.GetVaultAddress(),
		},
		Format: string(item.DataDetails.Interface.DataFormat),
	}
	// DataStore for destination will be determined if an implicit copy is required
	var sinkDataStore *app.DataStore

	var readSelector, copySelector *modules.Selector
	if readSelector, err = m.selectReadModule(item, appContext); err != nil {
		m.Log.Info("Could not select a read module for " + datasetID + " : " + err.Error())
		return instances, err
	}
	if copySelector, err = m.selectCopyModule(item, appContext, readSelector); err != nil {
		m.Log.Info("Could not select a copy module for " + datasetID + " : " + err.Error())
		return instances, err
	}

	if copySelector != nil {
		m.Log.Info("Found copy module " + copySelector.GetModule().Name + " for " + datasetID)
		// copy should be applied - allocate storage
		if sinkDataStore, err = m.GetCopyDestination(item, copySelector.Destination, copySelector.Geo); err != nil {
			m.Log.Info("Allocation failed: " + err.Error())
			return instances, err
		}
		// append moduleinstances to the list
		actions := actionsToArbitrary(copySelector.Actions)
		copyArgs := &app.ModuleArguments{
			Copy: &app.CopyModuleArgs{
				Source:          *sourceDataStore,
				Destination:     *sinkDataStore,
				Transformations: actions,
			},
		}
		copyCluster, err := copySelector.SelectCluster(item, m.Clusters)
		if err != nil {
			m.Log.Info("Could not determine the cluster for copy: " + err.Error())
			return instances, err
		}
		for _, cluster := range m.Clusters {
			if copyCluster == cluster.Name {
				copyArgs.Copy.Destination.Vault.AuthPath = utils.GetAuthPath(cluster.Metadata.VaultAuthPath)
				copyArgs.Copy.Source.Vault.AuthPath = utils.GetAuthPath(cluster.Metadata.VaultAuthPath)
				break
			}
		}

		m.Log.Info("Adding copy module")
		instances = copySelector.AddModuleInstances(copyArgs, item, copyCluster)
	}

	if readSelector != nil {
		m.Log.Info("Adding read path")
		var readSource app.DataStore
		if sinkDataStore == nil {
			readSource = *sourceDataStore
		} else {
			readSource = *sinkDataStore
		}

		actions := actionsToArbitrary(readSelector.Actions)
		readCluster, err := readSelector.SelectCluster(item, m.Clusters)
		if err != nil {
			m.Log.Info("Could not determine the cluster for read: " + err.Error())
			return instances, err
		}
		for _, cluster := range m.Clusters {
			if readCluster == cluster.Name {
				readSource.Vault.AuthPath = utils.GetAuthPath(cluster.Metadata.VaultAuthPath)
				break
			}
		}
		readInstructions := []app.ReadModuleArgs{
			{
				Source:          readSource,
				AssetID:         utils.CreateDataSetIdentifier(item.Context.DataSetID),
				Transformations: actions,
			},
		}

		readArgs := &app.ModuleArguments{
			Read: readInstructions,
		}

		instances = append(instances, readSelector.AddModuleInstances(readArgs, item, readCluster)...)
	}
	return instances, nil
}

// GetSupportedReadSources returns a list of supported READ interfaces of a module
func GetSupportedReadSources(module *app.M4DModule) []*app.InterfaceDetails {
	var list []*app.InterfaceDetails
	for _, inter := range module.Spec.Capabilities.SupportedInterfaces {
		if inter.Flow != app.Read {
			continue
		}
		list = append(list, inter.Source)
	}
	return list
}

// check whether IMPLICIT copy is required
// decide on actions performed on read (update readSelector)
// copy is required in the following cases:
// - specifically requested by the user
// - the read module does not support data interface
// - the read module does not support all governance actions
// - transformations are required while the read module does not run at source location
// output:
// - true if copy is required, false - otherwise
// - interface capabilities to match copy destination, based on read sources
// - read actions that copy has to support
func (m *ModuleManager) getCopyRequirements(item modules.DataInfo, readSelector *modules.Selector) (bool, []*app.InterfaceDetails, []*pb.EnforcementAction) {
	m.Log.Info("Checking supported read sources")
	sources := GetSupportedReadSources(readSelector.GetModule())
	// check if read sources include the data source
	supportsDataSource := utils.SupportsInterface(sources, &item.DataDetails.Interface)
	// check if read supports all governance actions
	supportsAllActions := readSelector.SupportsGovernanceActions(readSelector.GetModule(), readSelector.Actions)
	// Copy is required when data has to be transformed and read is done at another location
	transformAtSource := len(readSelector.Actions) > 0 && item.DataDetails.Geography != readSelector.Geo
	readActionsOnCopy := []*pb.EnforcementAction{}
	if transformAtSource {
		readActionsOnCopy = append(readActionsOnCopy, readSelector.Actions...)
		readSelector.Actions = []*pb.EnforcementAction{}
	} else {
		// ensure that copy + read support all needed actions
		// actions that the read module can not perform are required to be done during copy
		readActionsOnRead := []*pb.EnforcementAction{}
		for _, action := range readSelector.Actions {
			if !readSelector.SupportsGovernanceAction(readSelector.GetModule(), action) {
				readActionsOnCopy = append(readActionsOnCopy, action)
			} else {
				readActionsOnRead = append(readActionsOnRead, action)
			}
		}
		readSelector.Actions = readActionsOnRead
	}
	// debug info
	if !supportsDataSource {
		m.Log.Info("Copy is required to support read-path module interface")
		utils.PrintStructure(sources, m.Log, "Read sources")
	}
	if !supportsAllActions {
		m.Log.Info("Copy is required because the read-path does not support all actions")
		utils.PrintStructure(readActionsOnCopy, m.Log, "Unsupported actions")
	}
	if transformAtSource {
		m.Log.Info("Copy is required because " + readSelector.Geo + " does not match " + item.DataDetails.Geography)
	}
	if item.Context.Requirements.Copy.Required {
		m.Log.Info("Copy has been explicitly requested")
	}
	copyRequired := !supportsDataSource || !supportsAllActions || transformAtSource || item.Context.Requirements.Copy.Required
	return copyRequired, sources, readActionsOnCopy
}

func (m *ModuleManager) enforceWritePolicies(appContext *app.M4DApplication, datasetID string) ([]*pb.EnforcementAction, string, error) {
	var err error
	actions := []*pb.EnforcementAction{}
	//	if the cluster selector is non-empty, the write will be done to the specified geography if possible
	if m.WorkloadGeography != "" {
		if actions, err = LookupPolicyDecisions(datasetID, m.PolicyManager, appContext,
			&pb.AccessOperation{Type: pb.AccessOperation_WRITE, Destination: m.WorkloadGeography}); err == nil {
			return actions, m.WorkloadGeography, nil
		}
	}
	var excludedGeos string
	for _, cluster := range m.Clusters {
		operation := &pb.AccessOperation{Type: pb.AccessOperation_WRITE, Destination: cluster.Metadata.Region}
		if actions, err = LookupPolicyDecisions(datasetID, m.PolicyManager, appContext, operation); err == nil {
			return actions, cluster.Metadata.Region, nil
		}
		if err.Error() != app.WriteNotAllowed {
			return actions, "", err
		}
		if excludedGeos != "" {
			excludedGeos += ", "
		}
		excludedGeos += cluster.Metadata.Region
	}
	return actions, "", errors.New("writing to all geographies is denied: " + excludedGeos)
}

// GetProcessingGeography determines the geography of the workload cluster.
// If no cluster has been specified for a workload, a local cluster is assumed.
func (m *ModuleManager) GetProcessingGeography(applicationContext *app.M4DApplication) (string, error) {
	clusterName := applicationContext.Spec.Selector.ClusterName
	if clusterName == "" {
		if applicationContext.Spec.Selector.WorkloadSelector.Size() == 0 {
			// no workload
			return "", nil
		}
		// the workload runs in a local cluster
		localClusterManager, err := local.NewManager(m.Client, utils.GetSystemNamespace())
		if err != nil {
			return "", err
		}
		clusters, err := localClusterManager.GetClusters()
		if err != nil || len(clusters) != 1 {
			return "", err
		}
		return clusters[0].Metadata.Region, nil
	}
	for _, cluster := range m.Clusters {
		if cluster.Name == clusterName {
			return cluster.Metadata.Region, nil
		}
	}
	return "", errors.New("Unknown cluster: " + clusterName)
}

func actionsToArbitrary(actions []*pb.EnforcementAction) []serde.Arbitrary {
	result := []serde.Arbitrary{}
	for _, action := range actions {
		raw := serde.NewArbitrary(action)
		result = append(result, *raw)
	}
	return result
}
