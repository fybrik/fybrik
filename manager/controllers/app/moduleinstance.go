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
/* The order of the lookup is Read, Copy, Write.
   Assumptions:
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

// GetCopyDestination chooses one of the buckets pre-allocated for use by implicit copies.
// These buckets are allocated during deployment of the control plane.
// If there are no free buckets the creation of the runtime environment for the application will fail.
// TODO - In the future need to implement dynamic provisioning of buckets for implicit copy.
func (m *ModuleManager) GetCopyDestination(item modules.DataInfo, destinationInterface *app.InterfaceDetails) (*app.DataStore, error) {
	// provisioned storage for COPY
	originalAssetName := item.DataDetails.Name
	bucket := FindAvailableBucket(m.Client, m.Log, m.Owner, item.AssetID, originalAssetName, false)
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
				ObjectKey: bucket.Status.AssetPrefixPerDataset[item.AssetID],
			},
		},
		Format: string(destinationInterface.DataFormat),
	}, nil
}

// SelectModuleInstances selects the necessary read/copy/write modules for the blueprint
func (m *ModuleManager) SelectModuleInstances(item modules.DataInfo) ([]modules.ModuleInstanceSpec, error) {
	instances := make([]modules.ModuleInstanceSpec, 0)

	// Write path is not yet implemented
	var readSelector, copySelector *modules.Selector
	m.Log.Info("Select read path for " + item.AssetID)
	// Select a module that supports READ flow, supports actions-on-read, has the required dependency modules (recursively), with API = sink.
	actionsOnRead := item.Actions[app.Read]
	if !actionsOnRead.Allowed {
		return instances, errors.New(actionsOnRead.Message)
	}
	m.Log.Info("Finding modules for " + item.AssetID)
	// Each selector receives source/sink interface and relevant actions
	// Starting with the existing location for source and user request for sink
	source, err := StructToInterfaceDetails(item)
	if err != nil {
		return instances, err
	}
	sink := item.AppInterface
	var sourceDataStore, sinkDataStore *app.DataStore
	sourceDataStore = &app.DataStore{
		Connection:         item.DataDetails.GetDataStore(),
		CredentialLocation: utils.GetDatasetVaultPath(item.AssetID),
		Format:             item.DataDetails.DataFormat,
	}
	// DataStore for destination will be determined if an implicit copy is required
	sinkDataStore = nil

	// select a read module that supports user interface requirements
	// actions are not checked since they are not necessarily done by the read module
	readSelector = &modules.Selector{Flow: app.Read,
		Destination:  sink,
		Actions:      make([]pb.EnforcementAction, 0),
		Source:       nil,
		Dependencies: make([]*app.M4DModule, 0),
		Module:       nil,
		Message:      ""}
	if !readSelector.SelectModule(m.Modules) {
		m.Log.Info(item.AssetID + " : " + readSelector.GetError())
		return instances, errors.New(readSelector.GetError())
	}
	// logic for deciding whether copy module is required
	copyRequired, interfaces, actions := m.getCopyRequirements(item, readSelector)

	if copyRequired {
		m.Log.Info("Copy is required for " + item.AssetID)
		// is copy allowed?
		if !item.Actions[app.Copy].Allowed {
			return instances, errors.New(item.Actions[app.Copy].Message)
		}
		// select a module that supports COPY, supports required governance actions, has the required dependencies, with source in module sources and a non-empty intersection between READ_SOURCES and module destinations.
		for _, copyDest := range interfaces {
			copySelector = &modules.Selector{
				Flow:         app.Copy,
				Source:       source,
				Actions:      actions,
				Destination:  copyDest,
				Dependencies: make([]*app.M4DModule, 0),
				Module:       nil,
				Message:      ""}

			if copySelector.SelectModule(m.Modules) {
				break
			}
		}
		// no copy module - report an error
		if copySelector.GetModule() == nil {
			m.Log.Info("Could not find copy module for " + item.AssetID)
			return instances, errors.New(copySelector.GetError())
		}
		m.Log.Info("Found copy module " + copySelector.GetModule().Name)
		// copy should be applied - allocate storage
		if sinkDataStore, err = m.GetCopyDestination(item, copySelector.Destination); err != nil {
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
		AssetID:         item.AssetID,
		Transformations: readSelector.Actions})
	readArgs := &app.ModuleArguments{
		Read: readInstructions,
	}
	readCluster, err := readSelector.SelectCluster(item, m.Clusters)
	if err != nil {
		return instances, err
	}
	instances = append(instances, readSelector.AddModuleInstances(readArgs, item, readCluster)...)
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

// check whether copy is required
// decide on actions performed on read (update readSelector)
// copy is required in the following cases:
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
	supportsAllActions := readSelector.SupportsGovernanceActions(readSelector.GetModule(), item.Actions[app.Read].EnforcementActions)
	// Copy is required when data has to be transformed and read is done at another location
	transformAtSource := len(item.Actions[app.Read].EnforcementActions) > 0 && item.DataDetails.Geo != item.Geo
	actions := item.Actions[app.Copy].EnforcementActions
	if transformAtSource {
		actions = append(actions, item.Actions[app.Read].EnforcementActions...)
	} else {
		// ensure that copy + read support all needed actions
		// actions that the read module can not perform are required to be done during copy
		for _, action := range item.Actions[app.Read].EnforcementActions {
			if !readSelector.SupportsGovernanceAction(readSelector.GetModule(), action) {
				actions = append(actions, action)
			} else {
				readSelector.Actions = append(readSelector.Actions, action)
			}
		}
	}
	copyRequired := !supportsDataSource || !supportsAllActions || transformAtSource
	return copyRequired, sources, actions
}
