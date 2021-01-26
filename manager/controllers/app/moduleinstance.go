// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"emperror.dev/errors"
	comv1alpha1 "github.com/IBM/dataset-lifecycle-framework/src/dataset-operator/pkg/apis/com/v1alpha1"
	"github.com/go-logr/logr"
	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	modules "github.com/ibm/the-mesh-for-data/manager/controllers/app/modules"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
	local "github.com/ibm/the-mesh-for-data/pkg/multicluster/local"
	pc "github.com/ibm/the-mesh-for-data/pkg/policy-compiler/policy-compiler"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ModuleManager builds a set of modules based on the requirements (governance actions, data location) and the existing set of M4DModules
type ModuleManager struct {
	Client            client.Client
	Log               logr.Logger
	Modules           map[string]*app.M4DModule
	Clusters          []multicluster.Cluster
	Owner             types.NamespacedName
	PolicyCompiler    pc.IPolicyCompiler
	WorkloadGeography string
	Datasets          []*comv1alpha1.Dataset
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

// StructToInterfaceDetails constructs a valid InterfaceDetails object
func StructToInterfaceDetails(item modules.DataInfo) (*app.InterfaceDetails, error) {
	source := &app.InterfaceDetails{}
	var err error
	if source.Protocol, err = utils.GetProtocol(item.DataDetails); err != nil {
		return nil, err
	}
	if source.DataFormat, err = utils.GetDataFormat(item.DataDetails); err != nil {
		return nil, err
	}
	return source, nil
}

// GetCopyDestination creates a Dataset for bucket allocation by implicit copies or ingest.
func (m *ModuleManager) GetCopyDestination(item modules.DataInfo, destinationInterface *app.InterfaceDetails, geo string) (*app.DataStore, error) {
	// provisioned storage for COPY
	originalAssetName := item.DataDetails.Name
	var dataset *comv1alpha1.Dataset = nil
	var err error
	if dataset, err = AllocateBucket(m.Client, m.Log, m.Owner, item.Context.DataSetID, geo); err != nil {
		return nil, err
	}
	m.Datasets = append(m.Datasets, dataset)
	vaultPath := "/v1/" + utils.GetVaultDatasetHome() + dataset.Spec.Local["bucket"]
	// TODO(shlomitk1): fetch the secret and register credentials
	return &app.DataStore{
		CredentialLocation: utils.GetFullCredentialsPath(vaultPath),
		Connection: &pb.DataStore{
			Type: pb.DataStore_S3,
			Name: "S3",
			S3: &pb.S3DataStore{
				Bucket:    dataset.Spec.Local["bucket"],
				Endpoint:  dataset.Spec.Local["endpoint"],
				ObjectKey: originalAssetName + utils.Hash(originalAssetName, 10),
			},
		},
		Format: string(destinationInterface.DataFormat),
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
	readActions, err = LookupPolicyDecisions(item.Context.DataSetID, m.PolicyCompiler, appContext,
		pb.AccessOperation{Type: pb.AccessOperation_READ, Destination: m.WorkloadGeography})
	if err != nil {
		return nil, err
	}
	// select a read module that supports user interface requirements
	// actions are not checked since they are not necessarily done by the read module
	readSelector := &modules.Selector{Flow: app.Read,
		Destination:  &item.Context.Requirements.Interface,
		Actions:      make([]*pb.EnforcementAction, 0),
		Source:       nil,
		Dependencies: make([]*app.M4DModule, 0),
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
	// WRITE actions
	actionsOnCopy, geo, err := m.enforceWritePolicies(appContext, item.Context.DataSetID)
	if err != nil {
		if readSelector != nil && err.Error() == app.WriteNotAllowed {
			return nil, errors.New(app.CopyNotAllowed)
		}
		return nil, err
	}
	actionsOnCopy = append(actionsOnCopy, additionalActions...)
	source, err := StructToInterfaceDetails(item)
	if err != nil {
		return nil, err
	}

	m.Log.Info("Copy is required for " + item.Context.DataSetID)
	var copySelector *modules.Selector
	// select a module that supports COPY, supports required governance actions, has the required dependencies, with source in module sources and a non-empty intersection between requested and supported interfaces.
	for _, copyDest := range interfaces {
		copySelector = &modules.Selector{
			Flow:         app.Copy,
			Source:       source,
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
		return nil, errors.New("No copy module has been found supporting required source interface")
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

	// Each selector receives source/sink interface and relevant actions
	// Starting with the data location interface for source and the required interface for sink
	var sourceDataStore, sinkDataStore *app.DataStore
	sourceDataStore = &app.DataStore{
		Connection:         item.DataDetails.GetDataStore(),
		CredentialLocation: utils.GetDatasetVaultPath(datasetID),
		Format:             item.DataDetails.DataFormat,
	}
	// DataStore for destination will be determined if an implicit copy is required
	sinkDataStore = nil
	var readSelector, copySelector *modules.Selector
	if readSelector, err = m.selectReadModule(item, appContext); err != nil {
		m.Log.Info("Could not select a read module for " + datasetID + " : " + err.Error())
		return instances, err
	}
	if copySelector, err = m.selectCopyModule(item, appContext, readSelector); err != nil {
		m.Log.Info("Could not select a read module for " + datasetID + " : " + err.Error())
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
		copyTransformations := []pb.EnforcementAction{}
		for _, action := range copySelector.Actions {
			copyTransformations = append(copyTransformations, *action)
		}
		copyArgs := &app.ModuleArguments{
			Copy: &app.CopyModuleArgs{
				Source:          *sourceDataStore,
				Destination:     *sinkDataStore,
				Transformations: copyTransformations},
		}
		copyCluster, err := copySelector.SelectCluster(item, m.Clusters)
		if err != nil {
			m.Log.Info("Could not determine the cluster for copy: " + err.Error())
			return instances, err
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

		readTransformations := []pb.EnforcementAction{}
		for _, action := range readSelector.Actions {
			readTransformations = append(readTransformations, *action)
		}

		readInstructions := make([]app.ReadModuleArgs, 0)
		readInstructions = append(readInstructions, app.ReadModuleArgs{
			Source:          readSource,
			AssetID:         utils.CreateDataSetIdentifier(item.Context.DataSetID),
			Transformations: readTransformations})
		readArgs := &app.ModuleArguments{
			Read: readInstructions,
		}
		readCluster, err := readSelector.SelectCluster(item, m.Clusters)
		if err != nil {
			m.Log.Info("Could not determine the cluster for read: " + err.Error())
			return instances, err
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
	utils.PrintStructure(sources, m.Log, "Read sources")
	// check if read sources include the data source
	source, _ := StructToInterfaceDetails(item)
	supportsDataSource := utils.SupportsInterface(sources, source)
	// check if read supports all governance actions
	supportsAllActions := readSelector.SupportsGovernanceActions(readSelector.GetModule(), readSelector.Actions)
	// Copy is required when data has to be transformed and read is done at another location
	transformAtSource := len(readSelector.Actions) > 0 && item.DataDetails.Geo != readSelector.Geo
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
	copyRequired := !supportsDataSource || !supportsAllActions || transformAtSource || item.Context.Requirements.Copy.Required
	return copyRequired, sources, readActionsOnCopy
}

func (m *ModuleManager) enforceWritePolicies(appContext *app.M4DApplication, datasetID string) ([]*pb.EnforcementAction, string, error) {
	var geo string
	var err error
	actions := []*pb.EnforcementAction{}
	//	if the cluster selector is non-empty, the write will be done to the specified geography
	if m.WorkloadGeography != "" {
		if actions, err = LookupPolicyDecisions(datasetID, m.PolicyCompiler, appContext,
			pb.AccessOperation{Type: pb.AccessOperation_WRITE, Destination: m.WorkloadGeography}); err != nil {
			return actions, geo, err
		}
		return actions, m.WorkloadGeography, nil
	}
	var excludedGeos string
	for _, cluster := range m.Clusters {
		operation := pb.AccessOperation{Type: pb.AccessOperation_WRITE, Destination: cluster.Metadata.Region}
		if actions, err = LookupPolicyDecisions(datasetID, m.PolicyCompiler, appContext, operation); err == nil {
			return actions, cluster.Metadata.Region, nil
		}
		if excludedGeos != "" {
			excludedGeos += ", "
		}
		excludedGeos += cluster.Metadata.Region
		if err.Error() != app.WriteNotAllowed {
			return actions, "", err
		}
	}
	return actions, "", errors.New("Writing to all geographies is denied: " + excludedGeos)
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
		localClusterManager := local.NewManager(m.Client, utils.GetSystemNamespace())
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
