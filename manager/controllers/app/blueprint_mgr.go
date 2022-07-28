// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	app "fybrik.io/fybrik/manager/apis/app/v12"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/logging"

	"github.com/rs/zerolog"
)

// ModuleInstanceSpec consists of the module spec and arguments
type ModuleInstanceSpec struct {
	Module      app.BlueprintModule
	ClusterName string
	Scope       app.CapabilityScope
}

// RefineInstances collects all instances of the same read/write module with non "Asset" scope
// and creates a new instance instead, with accumulated arguments.
func (r *PlotterReconciler) RefineInstances(instances []ModuleInstanceSpec) []ModuleInstanceSpec {
	newInstances := make([]ModuleInstanceSpec, 0)
	// map instances to be unified, according to the cluster and module
	instanceMap := make(map[string]ModuleInstanceSpec)
	for ind := range instances {
		// If the module scope is of type "asset" then avoid trying to unify it with another module.
		// Copy module is assumed to be of "asset" scope
		if instances[ind].Scope == app.Asset {
			newInstances = append(newInstances, instances[ind])
			continue
		}
		key := instances[ind].Module.Name + "," + instances[ind].ClusterName
		if instance, ok := instanceMap[key]; !ok {
			instanceMap[key] = instances[ind]
		} else {
			instance.Module.Arguments.Assets = append(instance.Module.Arguments.Assets, instances[ind].Module.Arguments.Assets...)
			// AssetID is used for step name generation
			instance.Module.AssetIDs = append(instance.Module.AssetIDs, instances[ind].Module.AssetIDs...)
			instanceMap[key] = instance
		}
	}
	for moduleName := range instanceMap {
		newInstances = append(newInstances, instanceMap[moduleName])
	}
	return newInstances
}

// GenerateBlueprints creates Blueprint specs (one per cluster)
func (r *PlotterReconciler) GenerateBlueprints(instances []ModuleInstanceSpec, plotter *app.Plotter) map[string]app.BlueprintSpec {
	blueprintMap := make(map[string]app.BlueprintSpec)
	instanceMap := make(map[string][]ModuleInstanceSpec)
	uuid := utils.GetFybrikApplicationUUIDfromAnnotations(plotter.GetAnnotations())
	for ind := range instances {
		instanceMap[instances[ind].ClusterName] = append(instanceMap[instances[ind].ClusterName], instances[ind])
	}
	for key, instanceList := range instanceMap {
		// unite several instances of a read/write module
		instances := r.RefineInstances(instanceList)
		blueprintMap[key] = r.GenerateBlueprint(instances, key, plotter)
	}

	log := r.Log.With().Str(utils.FybrikAppUUID, uuid).Logger()
	logging.LogStructure("BlueprintMap", blueprintMap, &log, zerolog.DebugLevel, false, false)
	return blueprintMap
}

// GenerateBlueprint creates the Blueprint spec based on the datasets and the governance actions required,
// which dictate the modules that must run in the fybrik
// Credentials for accessing data set are stored in a credential management system (such as vault) and
// the paths for accessing them are included in the blueprint.
// The credentials themselves are not included in the blueprint.
func (r *PlotterReconciler) GenerateBlueprint(instances []ModuleInstanceSpec, clusterName string, plotter *app.Plotter) app.BlueprintSpec {
	spec := app.BlueprintSpec{
		Cluster:          clusterName,
		ModulesNamespace: plotter.Spec.ModulesNamespace,
		Modules:          map[string]app.BlueprintModule{},
		Application: &app.ApplicationDetails{
			WorkloadSelector: plotter.Spec.Selector.WorkloadSelector,
			Context:          plotter.Spec.AppInfo,
		},
	}
	// Create the map that contains BlueprintModules
	for ind := range instances {
		modulename := instances[ind].Module.Name
		instanceName := modulename
		if instances[ind].Scope == app.Asset {
			// Need unique name for each module
			// if the module scope is one per asset then concat the id of the asset to it
			instanceName = utils.CreateStepName(modulename, instances[ind].Module.AssetIDs[0])
		}
		spec.Modules[instanceName] = instances[ind].Module
	}
	return spec
}
