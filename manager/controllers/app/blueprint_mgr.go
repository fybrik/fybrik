// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/api/equality"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// ModuleInstanceSpec consists of the module spec and arguments
type ModuleInstanceSpec struct {
	Module      fapp.BlueprintModule
	ClusterName string
	Scope       fapp.CapabilityScope
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
		if instances[ind].Scope == fapp.Asset {
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
func (r *PlotterReconciler) GenerateBlueprints(instances []ModuleInstanceSpec,
	plotter *fapp.Plotter, services Services) map[string]fapp.BlueprintSpec {
	blueprintMap := make(map[string]fapp.BlueprintSpec)
	instanceMap := make(map[string][]ModuleInstanceSpec)
	uuid := utils.GetFybrikApplicationUUIDfromAnnotations(plotter.GetAnnotations())
	for ind := range instances {
		instanceMap[instances[ind].ClusterName] = append(instanceMap[instances[ind].ClusterName], instances[ind])
	}
	for key, instanceList := range instanceMap {
		// unite several instances of a read/write module
		instances := r.RefineInstances(instanceList)
		blueprintMap[key] = r.GenerateBlueprint(instances, key, plotter, services)
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
func (r *PlotterReconciler) GenerateBlueprint(instances []ModuleInstanceSpec,
	clusterName string, plotter *fapp.Plotter, services Services) fapp.BlueprintSpec {
	ingressMap := createIngressMap(plotter, instances, services)
	spec := fapp.BlueprintSpec{
		Cluster:          clusterName,
		ModulesNamespace: plotter.Spec.ModulesNamespace,
		Modules:          map[string]fapp.BlueprintModule{},
		Application: &fapp.ApplicationDetails{
			WorkloadSelector: plotter.Spec.Selector.WorkloadSelector,
			Context:          plotter.Spec.AppInfo,
		},
	}
	// Create the map that contains BlueprintModules
	for ind := range instances {
		var assetID string
		if len(instances[ind].Module.AssetIDs) > 0 {
			assetID = instances[ind].Module.AssetIDs[0]
		}
		instanceName := utils.CreateStepName(instances[ind].Module.Name, assetID, instances[ind].Scope)
		releaseName := utils.GetReleaseName(utils.GetApplicationNameFromLabels(plotter.Labels),
			utils.GetFybrikApplicationUUIDfromAnnotations(plotter.Annotations), instanceName)
		moduleKey := UniqueReleaseName(instances[ind].ClusterName, releaseName)
		isEndpoint := services[moduleKey].IsEndpoint
		// connections from the module to its arguments (other modules or data locations)
		urls := []string{}
		egress := []fapp.ModuleDeployment{}
		for _, asset := range instances[ind].Module.Arguments.Assets {
			for _, arg := range asset.Arguments {
				if arg != nil {
					if argKey, found := getServiceUniqueKey(arg.Connection, services); found {
						argService := services[argKey]
						egress = append(egress, fapp.ModuleDeployment{
							Cluster: argService.Cluster,
							Release: argService.Release,
							URL:     getURLFromConnection(arg.Connection)})
					} else {
						if url := getURLFromConnection(arg.Connection); url != "" {
							urls = append(urls, url)
						}
					}
				}
			}
		}
		// ingress
		ingress := []fapp.ModuleDeployment{}
		keys := ingressMap[moduleKey]
		for _, key := range keys {
			argService := services[key]
			ingress = append(ingress, fapp.ModuleDeployment{
				Cluster: argService.Cluster,
				Release: argService.Release,
				URL:     getURLFromConnection(argService.API.Connection)})
		}
		instances[ind].Module.Network = fapp.ModuleNetwork{
			Endpoint: isEndpoint,
			Egress:   egress,
			Ingress:  ingress,
			URLs:     urls,
		}
		spec.Modules[instanceName] = instances[ind].Module
	}
	return spec
}

// get module key (cluster + release) by api connection
func getServiceUniqueKey(conn taxonomy.Connection, services Services) (string, bool) {
	for key := range services {
		if equality.Semantic.DeepEqual(conn, services[key].API.Connection) {
			return key, true
		}
	}
	return "", false
}

// get module service URL or dataset location from the connection structure
func getURLFromConnection(conn taxonomy.Connection) string {
	// TBD
	return ""
}

func createIngressMap(plotter *fapp.Plotter, moduleInstances []ModuleInstanceSpec,
	services Services) map[string][]string {
	ingressMap := map[string][]string{}
	for ind := range moduleInstances {
		inst := &moduleInstances[ind]
		// cluster + release unique key of the module instance
		var assetID string
		if len(inst.Module.AssetIDs) > 0 {
			assetID = inst.Module.AssetIDs[0]
		}
		instanceName := utils.CreateStepName(inst.Module.Name, assetID, inst.Scope)
		releaseName := utils.GetReleaseName(utils.GetApplicationNameFromLabels(plotter.Labels),
			utils.GetFybrikApplicationUUIDfromAnnotations(plotter.Annotations), instanceName)
		moduleKey := UniqueReleaseName(inst.ClusterName, releaseName)
		for _, asset := range inst.Module.Arguments.Assets {
			for _, arg := range asset.Arguments {
				if arg != nil {
					if argKey, found := getServiceUniqueKey(arg.Connection, services); found {
						if ingressMap[argKey] == nil {
							ingressMap[argKey] = []string{}
						}
						ingressMap[argKey] = append(ingressMap[argKey], moduleKey)
					}
				}
			}
		}
	}
	return ingressMap
}
