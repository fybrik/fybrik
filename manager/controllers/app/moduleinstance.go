// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	modules "github.com/ibm/the-mesh-for-data/manager/controllers/app/modules"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SelectModuleInstances builds a list of required modules with the relevant arguments
/* The order of the lookup is Read, Copy, Write.
   Assumptions:
   - Read is always required.
   - Copy is used on demand, if a read module does not support the existing source of data
   - Write module has not yet been implemented - will be implemented in future release
   - Each module is responsible for all transformations required for its flow: read module performs actions on read, copy module - actions on copy, etc.
   - All data sets are processed, even if an error is encountered in one or more, to provide a complete status at the end of the reconcile
   - Dependencies are checked but not added yet to the blueprint
*/
func (r *M4DApplicationReconciler) SelectModuleInstances(requirements []modules.DataInfo, appContext *app.M4DApplication) ([]modules.ModuleInstanceSpec, error) {
	instances := make([]modules.ModuleInstanceSpec, 0)
	moduleMap, err := r.GetAllModules()
	if err != nil {
		return instances, err
	}
	for _, item := range requirements {
		instances = append(instances, r.SelectModuleInstancesPerDataset(item, appContext, moduleMap)...)
	}
	return instances, nil
}

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

// GetCopyDestination chooses on of the buckets pre-allocated for use by implicit copies.
// These buckets are allocated during deployment of the control plane.
// If there are no free buckets the creation of the runtime environment for the application will fail.
// TODO - In the future need to implement dynamic provisioning of buckets for implicit copy.
func (r *M4DApplicationReconciler) GetCopyDestination(item modules.DataInfo, appContext *app.M4DApplication, destinationInterface *app.InterfaceDetails) *app.DataStore {
	// provisioned storage for COPY
	objectKey, _ := client.ObjectKeyFromObject(appContext)
	originalAssetName := item.DataDetails.Name
	bucket := r.FindAvailableBucket(objectKey, item.AssetID, originalAssetName, false)
	if bucket == nil {
		setCondition(appContext, item.AssetID, app.InsufficientStorage, "Storage Provisioner", true)
		return nil
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
	}
}

// SelectModuleInstancesPerDataset selects the necessary read/copy/write modules for the blueprint
func (r *M4DApplicationReconciler) SelectModuleInstancesPerDataset(item modules.DataInfo, appContext *app.M4DApplication, moduleMap map[string]*app.M4DModule) []modules.ModuleInstanceSpec {
	instances := make([]modules.ModuleInstanceSpec, 0)

	// Write path is not yet implemented
	var readSelector, copySelector *modules.Selector
	r.Log.V(0).Info("Select read path for " + item.AssetID)
	// Select a module that supports READ flow, supports actions-on-read, has the required dependency modules (recursively), with API = sink.
	actionsOnRead := item.Actions[app.Read]
	r.Log.V(0).Info("Finding modules for " + item.AssetID)
	// Each selector receives source/sink interface and relevant actions
	// Starting with the existing location for source and user request for sink
	source, err := StructToInterfaceDetails(item)
	if err != nil {
		setCondition(appContext, item.AssetID, err.Error(), "", true)
		return instances
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

	readSelector = &modules.Selector{Flow: app.Read,
		Destination:  sink,
		Actions:      actionsOnRead.EnforcementActions,
		Source:       nil,
		Dependencies: make([]*app.M4DModule, 0),
		Module:       nil,
		Message:      ""}
	if !readSelector.SelectModule(moduleMap) {
		r.Log.V(0).Info(item.AssetID + " : " + readSelector.GetError())
		setCondition(appContext, item.AssetID, readSelector.GetError(), "", true)
		return instances
	}

	// If sources of this module include source, copy is not required
	r.Log.V(0).Info("Checking supported read sources")
	sources := GetSupportedReadSources(readSelector.GetModule())
	utils.PrintStructure(sources, r.Log, "Read sources")
	// logic for deciding whether copy module is required
	// In some cases copy is required to perform transformations at source
	// Temporary solution: in these cases mark copy actions as required until rules for transformations at data source are implemented in policy manager
	if !utils.SupportsInterface(sources, source) || item.Actions[app.Copy].Required {
		r.Log.V(0).Info("Copy is required for " + item.AssetID)
		// is copy allowed?
		actionsOnCopy := item.Actions[app.Copy]
		if !actionsOnCopy.Allowed {
			setCondition(appContext, item.AssetID, actionsOnCopy.Message, "", true)
			return instances
		}
		// select a module that supports COPY, supports actions-on-copy, has the required dependencies, with source in module sources and a non-empty intersection between READ_SOURCES and module destinations.
		for _, copyDest := range sources {
			copySelector = &modules.Selector{
				Flow:         app.Copy,
				Source:       source,
				Actions:      actionsOnCopy.EnforcementActions,
				Destination:  copyDest,
				Dependencies: make([]*app.M4DModule, 0),
				Module:       nil,
				Message:      ""}

			if copySelector.SelectModule(moduleMap) {
				break
			}
		}
		// no copy module - report an error
		if copySelector.GetModule() == nil {
			r.Log.V(0).Info("Could not find copy module for " + item.AssetID)
			setCondition(appContext, item.AssetID, copySelector.GetError(), "", true)
			return instances
		}
		r.Log.V(0).Info("Found copy module " + copySelector.GetModule().Name)
		// copy should be applied - allocate storage
		sinkDataStore = r.GetCopyDestination(item, appContext, copySelector.Destination)
		if sinkDataStore == nil {
			return instances
		}
		// append moduleinstances to the list
		copyArgs := &app.ModuleArguments{
			Copy: &app.CopyModuleArgs{
				Source:          *sourceDataStore,
				Destination:     *sinkDataStore,
				Transformations: actionsOnCopy.EnforcementActions},
		}
		r.Log.V(0).Info("Adding copy module")
		instances = copySelector.AddModuleInstances(copyArgs, item)
	}
	r.Log.V(0).Info("Adding read path")
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
		Transformations: actionsOnRead.EnforcementActions})
	readArgs := &app.ModuleArguments{
		Read: readInstructions,
	}
	instances = append(instances, readSelector.AddModuleInstances(readArgs, item)...)
	return instances
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
