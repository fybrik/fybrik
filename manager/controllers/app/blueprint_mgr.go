// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/app/modules"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
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

// GetBlueprintSignature returns the signature (name and namespace) of the blueprint owned by the given M4DApplication instance
func (r *M4DApplicationReconciler) GetBlueprintSignature(applicationContext *app.M4DApplication) *app.Blueprint {
	return &app.Blueprint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      applicationContext.Name + "-" + applicationContext.Namespace,
			Namespace: utils.GetSystemNamespace(),
		},
	}
}

// getBlueprint finds the Blueprint instance associated with this M4DApplication if one already exists,
// and creturns it
func (r *M4DApplicationReconciler) getBlueprint(applicationContext *app.M4DApplication) (*app.Blueprint, error) {
	// Get the Blueprints owned by this M4DApplication instance
	m4dBlueprint := r.GetBlueprintSignature(applicationContext)
	namespace, _ := client.ObjectKeyFromObject(m4dBlueprint)
	if err := r.Client.Get(context.Background(), namespace, m4dBlueprint); err != nil {
		return nil, err
	}
	return m4dBlueprint, nil
}

// DeleteOwnedBlueprint deletes the blueprint owned by the given M4DApplication instance,
// checking first that it actually exists
func (r *M4DApplicationReconciler) DeleteOwnedBlueprint(applicationContext *app.M4DApplication) {
	// Check first if the blueprint exists
	blueprint, err := r.getBlueprint(applicationContext)
	if err != nil || blueprint == nil {
		r.Log.V(0).Info("\tDidn't find a blueprint to delete: " + applicationContext.GetName())
		return
	}

	if err = r.Delete(context.Background(), blueprint); err != nil {
		r.Log.V(0).Info("\tFailed to delete blueprint: " + applicationContext.GetName())
	}
}

// GenerateBlueprint creates the Blueprint spec based on the datasets and the governance actions required, which dictate the modules that must run in the m4d
// Credentials for accessing data set are stored in a credential management system (such as vault) and the paths for accessing them are included in the blueprint.
// The credentials themselves are not included in the blueprint.
func (r *M4DApplicationReconciler) GenerateBlueprint(instances []modules.ModuleInstanceSpec, appContext *app.M4DApplication) *app.BlueprintSpec {
	// If no modules received return error.  We should have at least an application
	if len(instances) == 0 {
		return nil
	}
	var spec app.BlueprintSpec

	// clone the selector
	appContext.Spec.Selector.DeepCopyInto(&spec.Selector)
	// Entrypoint is always the name of the application
	appName := appContext.GetName()
	spec.Entrypoint = appName
	r.Log.V(0).Info("\tappName: " + appName)

	// Define the flow structure, which indicates the flow of data between the components in the m4d
	// Loop over the list of modules and create a step for each
	// Also create a template for each module specification - i.e. there could be multiple instances of a module, each with different arguments
	// TODO - currently assumes one data set per module instance.  Read and Write modules can receive multiple data sets
	var flow app.DataFlow
	flow.Name = appName
	var steps []app.FlowStep
	var templates []app.ComponentTemplate
	for _, moduleInstance := range instances {
		modulename := moduleInstance.Module.GetName()

		// Create a flow step
		var step app.FlowStep
		step.Name = appName + "-" + modulename + "-" + moduleInstance.AssetID // Need unique name for each step so include ids for dataset
		step.Template = modulename

		step.Arguments = *moduleInstance.Args

		steps = append(steps, step)

		// If one doesn't exist already, create a template
		if !containsTemplate(templates, modulename) {
			var template app.ComponentTemplate
			template.Name = modulename
			template.Kind = moduleInstance.Module.TypeMeta.Kind
			template.Resources = make([]string, 1)
			template.Resources[0] = moduleInstance.Module.Spec.Chart

			templates = append(templates, template)
		}
	}
	flow.Steps = steps

	spec.Flow = flow
	spec.Templates = templates

	return &spec
}
