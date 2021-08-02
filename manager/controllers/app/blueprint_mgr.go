// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/app/modules"
	"fybrik.io/fybrik/manager/controllers/utils"
	// Temporary - shouldn't have something specific to implicit copies
)

func containsTemplate(templateList []app.ComponentTemplate, moduleName string) bool {
	for _, template := range templateList {
		if template.Name == moduleName {
			return true
		}
	}

	return false
}

// RefineInstances collects all instances of the same read/write module and creates a new instance instead, with accumulated arguments.
// Copy modules are left unchanged.
func (r *FybrikApplicationReconciler) RefineInstances(instances []modules.ModuleInstanceSpec) []modules.ModuleInstanceSpec {
	newInstances := make([]modules.ModuleInstanceSpec, 0)
	// map instances to be unified, according to the cluster and module
	instanceMap := make(map[string]modules.ModuleInstanceSpec)
	for _, moduleInstance := range instances {
		if moduleInstance.Args.Copy != nil {
			newInstances = append(newInstances, moduleInstance)
			continue
		}
		key := moduleInstance.Module.GetName() + "," + moduleInstance.ClusterName
		if instance, ok := instanceMap[key]; !ok {
			instanceMap[key] = moduleInstance
		} else {
			instance.Args.Read = append(instance.Args.Read, moduleInstance.Args.Read...)
			instance.Args.Write = append(instance.Args.Write, moduleInstance.Args.Write...)
			// AssetID is used for step name generation
			instance.AssetID += "," + moduleInstance.AssetID
			instanceMap[key] = instance
		}
	}
	for _, moduleInstance := range instanceMap {
		newInstances = append(newInstances, moduleInstance)
	}
	return newInstances
}

// GenerateBlueprints creates Blueprint specs (one per cluster)
func (r *FybrikApplicationReconciler) GenerateBlueprints(instances []modules.ModuleInstanceSpec, appContext *app.FybrikApplication) map[string]app.BlueprintSpec {
	blueprintMap := make(map[string]app.BlueprintSpec)
	instanceMap := make(map[string][]modules.ModuleInstanceSpec)
	for _, moduleInstance := range instances {
		instanceMap[moduleInstance.ClusterName] = append(instanceMap[moduleInstance.ClusterName], moduleInstance)
	}
	for key, instanceList := range instanceMap {
		// unite several instances of a read/write module
		instances := r.RefineInstances(instanceList)
		blueprintMap[key] = r.GenerateBlueprint(instances, appContext)
	}
	utils.PrintStructure(blueprintMap, r.Log, "BlueprintMap")
	return blueprintMap
}

// GenerateBlueprint creates the Blueprint spec based on the datasets and the governance actions required, which dictate the modules that must run in the fybrik
// Credentials for accessing data set are stored in a credential management system (such as vault) and the paths for accessing them are included in the blueprint.
// The credentials themselves are not included in the blueprint.
func (r *FybrikApplicationReconciler) GenerateBlueprint(instances []modules.ModuleInstanceSpec, appContext *app.FybrikApplication) app.BlueprintSpec {
	var spec app.BlueprintSpec

	// Entrypoint is always the name of the application
	appName := appContext.GetName()
	spec.Entrypoint = appName
	r.Log.V(0).Info("\tappName: " + appName)

	// Define the flow structure, which indicates the flow of data between the components in the fybrik
	// Loop over the list of modules and create a step for each
	// Also create a template for each module specification - i.e. there could be multiple instances of a module, each with different arguments
	var flow app.DataFlow
	flow.Name = appName
	var steps []app.FlowStep
	var templates []app.ComponentTemplate
	for _, moduleInstance := range instances {
		modulename := moduleInstance.Module.GetName()

		// Create a flow step
		var step app.FlowStep
		step.Name = utils.CreateStepName(modulename, moduleInstance.AssetID) // Need unique name for each step so include ids for dataset
		step.Template = modulename

		step.Arguments = *moduleInstance.Args

		steps = append(steps, step)

		// If one doesn't exist already, create a template
		if !containsTemplate(templates, modulename) {
			var template app.ComponentTemplate
			template.Name = modulename
			template.Kind = moduleInstance.Module.TypeMeta.Kind
			template.Chart = moduleInstance.Module.Spec.Chart

			templates = append(templates, template)
		}
	}
	flow.Steps = steps

	spec.Flow = flow
	spec.Templates = templates

	return spec
}
