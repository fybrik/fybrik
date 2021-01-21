// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"emperror.dev/errors"
	"github.com/go-logr/logr"
	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	modules "github.com/ibm/the-mesh-for-data/manager/controllers/app/modules"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ModuleManager builds a set of modules based on the requirements (governance actions, data location) and the existing set of M4DModules
type ModuleManager struct {
	Client   client.Client
	Log      logr.Logger
	Modules  map[string]*app.M4DModule
	Clusters []multicluster.Cluster
	Owner    types.NamespacedName
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

// GetCopyDestination chooses one of the buckets pre-allocated for use by implicit copies or ingest.
// These buckets are allocated during deployment of the control plane.
// If there are no free buckets the creation of the runtime environment for the application will fail.
// TODO - In the future need to implement dynamic provisioning of buckets for implicit copy.
func (m *ModuleManager) GetCopyDestination(item modules.DataInfo, destinationInterface *app.InterfaceDetails, geo string) (*app.DataStore, error) {
	// provisioned storage for COPY
	originalAssetName := item.DataDetails.Name
	bucket := FindAvailableBucket(m.Client, m.Log, m.Owner, item.Context.DataSetID, originalAssetName, false, geo)
	if bucket == nil {
		return nil, errors.New(app.InsufficientStorage)
	}
	return &app.DataStore{
		CredentialLocation: utils.GetFullCredentialsPath(bucket.Spec.VaultPath),
		Connection: &pb.DataStore{
			Type: pb.DataStore_S3,
			Name: "S3",
			S3: &pb.S3DataStore{
				Bucket:    bucket.Spec.Name,
				Endpoint:  bucket.Spec.Endpoint,
				ObjectKey: bucket.Status.AssetPrefixPerDataset[item.Context.DataSetID],
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
	// select a read module that supports user interface requirements
	// actions are not checked since they are not necessarily done by the read module
	readSelector := &modules.Selector{Flow: app.Read,
		Destination:  &item.Context.Requirements.Interface,
		Actions:      make([]pb.EnforcementAction, 0),
		Source:       nil,
		Dependencies: make([]*app.M4DModule, 0),
		Module:       nil,
		Message:      ""}
	if !readSelector.SelectModule(m.Modules) {
		m.Log.Info(readSelector.GetError())
		return nil, errors.New(readSelector.GetError())
	}
	return readSelector, nil
}

func (m *ModuleManager) selectCopyModule(item modules.DataInfo, appContext *app.M4DApplication, readSelector *modules.Selector) (*modules.Selector, error) {
	// logic for deciding whether copy module is required
	var interfaces []*app.InterfaceDetails
	var copyRequired bool
	var actionsOnCopy []pb.EnforcementAction

	if readSelector != nil {
		copyRequired, interfaces, actionsOnCopy = m.getCopyRequirements(item, readSelector)
	} else if item.Context.Requirements.Copy.Required {
		copyRequired = true
		actionsOnCopy = item.Actions[pb.AccessOperation_WRITE].EnforcementActions
		interfaces = []*app.InterfaceDetails{&item.Context.Requirements.Interface}
	}
	if !copyRequired {
		return nil, nil
	}
	if !item.Actions[pb.AccessOperation_WRITE].Allowed {
		return nil, errors.New(app.CopyNotAllowed)
	}
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
	instances := make([]modules.ModuleInstanceSpec, 0)
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
	var err error

	var readSelector, copySelector *modules.Selector
	if readSelector, err = m.selectReadModule(item, appContext); err != nil {
		return instances, err
	}
	if copySelector, err = m.selectCopyModule(item, appContext, readSelector); err != nil {
		return instances, err
	}

	if copySelector != nil {
		m.Log.Info("Found copy module " + copySelector.GetModule().Name)
		// copy should be applied - allocate storage
		if sinkDataStore, err = m.GetCopyDestination(item, copySelector.Destination, item.Actions[pb.AccessOperation_WRITE].Geo); err != nil {
			return instances, nil
		}
		// append moduleinstances to the list
		copyArgs := &app.ModuleArguments{
			Copy: &app.CopyModuleArgs{
				Source:          *sourceDataStore,
				Destination:     *sinkDataStore,
				Transformations: copySelector.Actions},
		}
		copyCluster, err := copySelector.SelectCluster(item, m.Clusters)
		if err != nil {
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

		readInstructions := make([]app.ReadModuleArgs, 0)
		readInstructions = append(readInstructions, app.ReadModuleArgs{
			Source:          readSource,
			AssetID:         utils.CreateDataSetIdentifier(item.Context.DataSetID),
			Transformations: readSelector.Actions})
		readArgs := &app.ModuleArguments{
			Read: readInstructions,
		}
		readCluster, err := readSelector.SelectCluster(item, m.Clusters)
		if err != nil {
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
// - actions that copy has to support
func (m *ModuleManager) getCopyRequirements(item modules.DataInfo, readSelector *modules.Selector) (bool, []*app.InterfaceDetails, []pb.EnforcementAction) {
	m.Log.Info("Checking supported read sources")
	sources := GetSupportedReadSources(readSelector.GetModule())
	utils.PrintStructure(sources, m.Log, "Read sources")
	// check if read sources include the data source
	source, _ := StructToInterfaceDetails(item)
	supportsDataSource := utils.SupportsInterface(sources, source)
	// check if read supports all governance actions
	actionsOnRead := item.Actions[pb.AccessOperation_READ]
	supportsAllActions := readSelector.SupportsGovernanceActions(readSelector.GetModule(), actionsOnRead.EnforcementActions)
	// Copy is required when data has to be transformed and read is done at another location
	transformAtSource := len(actionsOnRead.EnforcementActions) > 0 && item.DataDetails.Geo != actionsOnRead.Geo
	actionsOnCopy := item.Actions[pb.AccessOperation_WRITE].EnforcementActions
	if transformAtSource {
		actionsOnCopy = append(actionsOnCopy, actionsOnRead.EnforcementActions...)
	} else {
		// ensure that copy + read support all needed actions
		// actions that the read module can not perform are required to be done during copy
		for _, action := range actionsOnRead.EnforcementActions {
			if !readSelector.SupportsGovernanceAction(readSelector.GetModule(), action) {
				actionsOnCopy = append(actionsOnCopy, action)
			} else {
				readSelector.Actions = append(readSelector.Actions, action)
			}
		}
	}
	copyRequired := !supportsDataSource || !supportsAllActions || transformAtSource || item.Context.Requirements.Copy.Required
	return copyRequired, sources, actionsOnCopy
}
