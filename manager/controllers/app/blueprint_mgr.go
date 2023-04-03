// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/api/equality"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	mngrUtils "fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/utils"
)

const sep string = ","

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
		key := instances[ind].Module.Name + sep + instances[ind].ClusterName
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
	uuid := mngrUtils.GetFybrikApplicationUUIDfromAnnotations(plotter.GetAnnotations())
	ingressMap := createIngressMap(plotter, instances, services)
	for ind := range instances {
		instanceMap[instances[ind].ClusterName] = append(instanceMap[instances[ind].ClusterName], instances[ind])
	}
	for key, instanceList := range instanceMap {
		// unite several instances of a read/write module
		instances := r.RefineInstances(instanceList)
		blueprintMap[key] = r.GenerateBlueprint(instances, key, plotter, services, ingressMap)
	}

	log := r.Log.With().Str(mngrUtils.FybrikAppUUID, uuid).Logger()
	logging.LogStructure("BlueprintMap", blueprintMap, &log, zerolog.DebugLevel, false, false)
	return blueprintMap
}

// GenerateBlueprint creates the Blueprint spec based on the datasets and the governance actions required,
// which dictate the modules that must run in the fybrik
// Credentials for accessing data set are stored in a credential management system (such as vault) and
// the paths for accessing them are included in the blueprint.
// The credentials themselves are not included in the blueprint.
func (r *PlotterReconciler) GenerateBlueprint(instances []ModuleInstanceSpec,
	clusterName string, plotter *fapp.Plotter, services Services, ingressMap map[string][]string) fapp.BlueprintSpec {
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
		instanceName := mngrUtils.CreateStepName(instances[ind].Module.Name, assetID, instances[ind].Scope)
		releaseName := mngrUtils.GetReleaseName(mngrUtils.GetApplicationNameFromLabels(plotter.Labels),
			mngrUtils.GetFybrikApplicationUUIDfromAnnotations(plotter.Annotations), instanceName)
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
							URLs:    getURLsFromConnection(arg.Connection)})
					} else {
						urls = append(urls, getURLsFromConnection(arg.Connection)...)
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
				URLs:    getURLsFromConnection(argService.API.Connection)})
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
		if services[key].API != nil && equality.Semantic.DeepEqual(conn, services[key].API.Connection) {
			return key, true
		}
	}
	return "", false
}

// get module service URL or dataset location from the connection structure
func getURLsFromConnection(conn taxonomy.Connection) []string {
	urls := []string{}
	exprList := []string{"endpoint", "url", "host"}
	if host, found := matchProperty(conn.AdditionalProperties.Items, conn.Name, exprList); found {
		if port, found := matchProperty(conn.AdditionalProperties.Items, conn.Name, []string{"port"}); found {
			urls = append(urls, fmt.Sprintf("%s:%s", host, port))
		} else {
			urls = append(urls, host)
		}
	}
	if hosts, found := matchProperty(conn.AdditionalProperties.Items, conn.Name, []string{"servers"}); found {
		urls = append(urls, strings.Split(hosts, sep)...)
	}
	return urls
}

// get property containing a substring from a given list
func matchProperty(props map[string]interface{}, t taxonomy.ConnectionType, exprList []string) (string, bool) {
	propertyMap := props[string(t)]
	if propertyMap == nil {
		return "", false
	}
	switch propertyMap := propertyMap.(type) {
	case map[string]interface{}:
		for key := range propertyMap {
			for i := range exprList {
				expr := strings.ToLower(exprList[i])
				if strings.Contains(strings.ToLower(key), expr) {
					return fmt.Sprintf("%v", propertyMap[key]), true
				}
			}
		}
	default:
		break
	}
	return "", false
}

// ingress map that for each service contains a list of services (represented by release and cluster) that connect to this service
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
		instanceName := mngrUtils.CreateStepName(inst.Module.Name, assetID, inst.Scope)
		releaseName := mngrUtils.GetReleaseName(mngrUtils.GetApplicationNameFromLabels(plotter.Labels),
			mngrUtils.GetFybrikApplicationUUIDfromAnnotations(plotter.Annotations), instanceName)
		moduleKey := UniqueReleaseName(inst.ClusterName, releaseName)
		for _, asset := range inst.Module.Arguments.Assets {
			for _, arg := range asset.Arguments {
				if arg != nil {
					if argKey, found := getServiceUniqueKey(arg.Connection, services); found {
						if ingressMap[argKey] == nil {
							ingressMap[argKey] = []string{}
						}
						moduleKeys := ingressMap[argKey]
						if !utils.HasString(moduleKey, moduleKeys) {
							ingressMap[argKey] = append(moduleKeys, moduleKey)
						}
					}
				}
			}
		}
	}
	return ingressMap
}
