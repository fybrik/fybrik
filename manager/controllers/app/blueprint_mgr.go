// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	// Temporary - shouldn't have something specific to implicit copies
)

// ModuleInstanceSpec consists of the module spec and arguments
type ModuleInstanceSpec struct {
	Chart       *app.ChartSpec
	Args        *app.ModuleArguments
	AssetIDs    []string
	ClusterName string
	ModuleName  string
	Scope       app.CapabilityScope
}

// RefineInstances collects all instances of the same read/write module with non "Asset" scope
// and creates a new instance instead, with accumulated arguments.
func (r *PlotterReconciler) RefineInstances(instances []ModuleInstanceSpec) []ModuleInstanceSpec {
	newInstances := make([]ModuleInstanceSpec, 0)
	// map instances to be unified, according to the cluster and module
	instanceMap := make(map[string]ModuleInstanceSpec)
	for _, moduleInstance := range instances {
		// If the module scope is of type "asset" then avoid trying to unify it with another module.
		// Copy module is assumed to be of "asset" scope
		if moduleInstance.Scope == app.Asset {
			newInstances = append(newInstances, moduleInstance)
			continue
		}
		key := moduleInstance.ModuleName + "," + moduleInstance.ClusterName
		if instance, ok := instanceMap[key]; !ok {
			instanceMap[key] = moduleInstance
		} else {
			instance.Args.Read = append(instance.Args.Read, moduleInstance.Args.Read...)
			instance.Args.Write = append(instance.Args.Write, moduleInstance.Args.Write...)
			// AssetID is used for step name generation
			instance.AssetIDs = append(instance.AssetIDs, moduleInstance.AssetIDs...)
			instanceMap[key] = instance
		}
	}
	for _, moduleInstance := range instanceMap {
		newInstances = append(newInstances, moduleInstance)
	}
	return newInstances
}

// GenerateBlueprints creates Blueprint specs (one per cluster)
func (r *PlotterReconciler) GenerateBlueprints(instances []ModuleInstanceSpec) map[string]app.BlueprintSpec {
	blueprintMap := make(map[string]app.BlueprintSpec)
	instanceMap := make(map[string][]ModuleInstanceSpec)
	for _, moduleInstance := range instances {
		instanceMap[moduleInstance.ClusterName] = append(instanceMap[moduleInstance.ClusterName], moduleInstance)
	}
	for key, instanceList := range instanceMap {
		// unite several instances of a read/write module
		instances := r.RefineInstances(instanceList)
		blueprintMap[key] = r.GenerateBlueprint(instances, key)
	}
	utils.PrintStructure(blueprintMap, r.Log, "BlueprintMap")
	return blueprintMap
}

// GenerateBlueprint creates the Blueprint spec based on the datasets and the governance actions required, which dictate the modules that must run in the fybrik
// Credentials for accessing data set are stored in a credential management system (such as vault) and the paths for accessing them are included in the blueprint.
// The credentials themselves are not included in the blueprint.
func (r *PlotterReconciler) GenerateBlueprint(instances []ModuleInstanceSpec, clusterName string) app.BlueprintSpec {
	var spec app.BlueprintSpec

	spec.Cluster = clusterName

	// Create the map that contains BlueprintModules

	var blueprintModules = make(map[string]app.BlueprintModule)
	for _, moduleInstance := range instances {
		modulename := moduleInstance.ModuleName

		var blueprintModule app.BlueprintModule
		instanceName := modulename
		if moduleInstance.Scope == app.Asset {
			// Need unique name for each module
			// if the module scope is one per asset then concat the id of the asset to it
			instanceName = utils.CreateStepName(modulename, moduleInstance.AssetIDs[0])
		}
		blueprintModule.Name = modulename
		blueprintModule.Arguments = *moduleInstance.Args
		blueprintModule.Chart = *moduleInstance.Chart
		blueprintModule.AssetIDs = make([]string, len(moduleInstance.AssetIDs))
		copy(blueprintModule.AssetIDs, moduleInstance.AssetIDs)
		blueprintModules[instanceName] = blueprintModule
	}
	spec.Modules = blueprintModules

	return spec
}
