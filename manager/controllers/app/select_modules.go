// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"strings"

	"emperror.dev/errors"
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/app/assetmetadata"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/multicluster"
	taxonomymodels "fybrik.io/fybrik/pkg/taxonomy/model/policymanager/base"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// DataInfo defines all the information about the given data set that comes from the fybrikapplication spec and from the connectors.
type DataInfo struct {
	// Source connection details
	DataDetails *assetmetadata.DataDetails
	// The path to Vault secret which holds the dataset credentials
	VaultSecretPath string
	// Pointer to the relevant data context in the Fybrik application spec
	Context *app.DataContext
	// Evaluated config policies
	Configuration adminconfig.EvaluatorOutput
	// Workload cluster
	WorkloadCluster multicluster.Cluster
	// Governance actions to perform on this asset
	Actions []taxonomymodels.Action
}

// Temporary hard-coded capability representing Actions of read/copy capability.
// A change to modules is requested to add "transform" as an additional capability to the existing read and copy modules.
const Transform = "transform"

// Node represents an access point to data (as a physical source/sink, or a virtual endpoint)
// A virtual endpoint is activated by the workload for read/write actions.
type Node struct {
	Connection *app.InterfaceDetails
	Virtual    bool
}

// Edge represents a module capability that gets data via source and returns data via sink interface
type Edge struct {
	Source          *Node
	Sink            *Node
	Module          *app.FybrikModule
	CapabilityIndex int
}

// ResolvedEdge extends an Edge by adding actions that a module should perform, and the cluster where the module will be deployed
// TODO(shlomitk1): add plugins/transformation capabilities to this structure
type ResolvedEdge struct {
	Edge
	Actions              []taxonomymodels.Action
	Cluster              string
	StorageAccountRegion string
}

// Solution is a final solution enabling a plotter construction.
// It represents a full data flow between the data source and the workload.
type Solution struct {
	DataPath []ResolvedEdge
}

// FindPaths finds all valid data paths between the data source and the workload
// First, data paths are constructed using interface connections, starting from data source.
// Then, transformations are added to the found paths, and clusters are matched to satisfy restrictions from admin config policies.
// Optimization is done by the shortest path (the paths are sorted by the length). To be changed in future versions.
func (p *PlotterGenerator) FindPaths(item *DataInfo, appContext *app.FybrikApplication) []Solution {
	// data source as appears in the asset metadata
	source := Node{Connection: &item.DataDetails.Interface, Virtual: false}
	// data sink, either a virtual endpoint in read scenarios, or a datastore as in ingest scenario
	destination := Node{Connection: &item.Context.Requirements.Interface, Virtual: (appContext.Spec.Selector.WorkloadSelector.Size() > 0)}
	// find data paths of length up to DATAPATH_LIMIT from data source to the workload, not including transformations or branches
	bound, err := utils.GetDataPathMaxSize()
	if err != nil {
		fmt.Println("Warning: a default value for DATAPATH_LIMIT will be used")
	}
	solutions := p.findPathsWithinLimit(item, &source, &destination, bound)
	// get valid solutions by extending data paths with transformations and selecting an appropriate cluster for each capability
	solutions = p.validSolutions(item, solutions, appContext)
	return solutions
}

// extend the received data paths with transformations and select an appropriate cluster for each capability in a data path
func (p *PlotterGenerator) validSolutions(item *DataInfo, solutions []Solution, appContext *app.FybrikApplication) []Solution {
	validPaths := []Solution{}
	for ind := range solutions {
		if p.validate(item, solutions[ind], appContext) {
			validPaths = append(validPaths, solutions[ind])
		}
	}
	return validPaths
}

func (p *PlotterGenerator) validate(item *DataInfo, solution Solution, appContext *app.FybrikApplication) bool {
	// start from data source, check supported actions and cluster restrictions
	requiredActions := item.Actions
	for ind := range solution.DataPath {
		element := &solution.DataPath[ind]
		element.Actions = []taxonomymodels.Action{}
		moduleCapability := element.Module.Spec.Capabilities[element.CapabilityIndex]
		if !element.Edge.Sink.Virtual {
			// storage is required, plus more actions on copy may be needed
			for _, region := range p.StorageAccountRegions {
				// query the policy manager whether WRITE operation is allowed
				operation := new(taxonomymodels.PolicyManagerRequestAction)
				operation.SetActionType(taxonomymodels.WRITE)
				operation.SetDestination(region)
				actions, err := LookupPolicyDecisions(item.Context.DataSetID, p.PolicyManager, appContext, operation)
				if err != nil && err.Error() == app.WriteNotAllowed {
					continue
				}
				// check whether WRITE actions are supported by the capability that writes the data
				if !supportsGovernanceActions(&element.Edge, actions) {
					continue
				}
				// add WRITE actions and the selected storage account region
				element.Actions = actions
				element.StorageAccountRegion = region
			}
			if element.StorageAccountRegion == "" {
				p.Log.Fatal().Str(logging.DATASETID, item.Context.DataSetID).Msg("Could not find a storage account, aborting data path construction")
				return false
			}
		}
		// read actions need to be handled somewhere on the path
		unsupported := []taxonomymodels.Action{}
		for _, action := range requiredActions {
			if supportsGovernanceAction(&element.Edge, action) {
				element.Actions = append(element.Actions, action)
			} else {
				// forward actions to the next capability in the data path
				unsupported = append(unsupported, action)
			}
		}
		requiredActions = unsupported
		// select a cluster for the capability that satisfy cluster restrictions specified in admin config policies
		if !p.findCluster(item, element, appContext) {
			p.Log.Error().Str(logging.DATASETID, item.Context.DataSetID).Msg("Could not find an available cluster for " + moduleCapability.Capability)
			return false
		}
	}
	// Are all actions supported by the capabilities in this data path?
	if len(requiredActions) > 0 {
		p.Log.Fatal().Str(logging.DATASETID, item.Context.DataSetID).Msg("Not all governance actions are supported, aborting data path construction")
		return false
	}
	// Are all capabilities that need to be deployed supported in this data path?
	supportedCapabilities := []string{}
	for _, element := range solution.DataPath {
		supportedCapabilities = append(supportedCapabilities, element.Module.Spec.Capabilities[element.CapabilityIndex].Capability)
	}
	for capability, decision := range item.Configuration.ConfigDecisions {
		if decision.Deploy == v1.ConditionTrue {
			// check that it is supported
			if !utils.HasString(capability, supportedCapabilities) {
				return false
			}
		}
	}
	return true
}

// find a cluster that satisfies the requirements
func (p *PlotterGenerator) findCluster(item *DataInfo, element *ResolvedEdge, appContext *app.FybrikApplication) bool {
	for _, cluster := range p.Clusters {
		if validateClusterRestrictions(item, element, cluster) {
			element.Cluster = cluster.Name
			return true
		}
	}
	return false
}

// find all data paths up to length = n
// Only data movements between data stores/endpoints are considered.
// Transformations are added in the later stage.
// Capabilities outside the data path are not handled yet.
func (p *PlotterGenerator) findPathsWithinLimit(item *DataInfo, source *Node, sink *Node, n int) []Solution {
	solutions := []Solution{}
	for _, module := range p.Modules {
		for capabilityInd, capability := range module.Spec.Capabilities {
			// check if capability is allowed
			if !allowCapability(item, capability.Capability) {
				continue
			}
			edge := Edge{Module: module, CapabilityIndex: capabilityInd, Source: nil, Sink: nil}
			// check that the module + module capability satisfy the requirements from the admin config policies
			if !validateModuleRestrictions(item, &edge) {
				continue
			}
			// check whether the module supports the final destination
			if !supportsSinkInterface(&edge, sink) {
				continue
			}
			edge.Sink = sink
			// if a module supports both source and sink interfaces, it's an end of the recursion
			if supportsSourceInterface(&edge, source) {
				edge.Source = source
				// found a path
				var path []ResolvedEdge
				path = append(path, ResolvedEdge{Edge: edge})
				solutions = append(solutions, Solution{DataPath: path})
			}
			// try to build data paths using the selected module capability
			if n > 1 {
				for _, inter := range capability.SupportedInterfaces {
					// recursive call to find paths of length = n-1 using the supported source of the selected module capability
					paths := p.findPathsWithinLimit(item, source, &Node{Connection: inter.Source}, n-1)
					// add the selected module to the found paths
					for i := range paths {
						auxEdge := Edge{Module: module, CapabilityIndex: capabilityInd, Source: &Node{Connection: inter.Source}, Sink: sink}
						paths[i].DataPath = append(paths[i].DataPath, ResolvedEdge{Edge: auxEdge})
					}
					if len(paths) > 0 {
						solutions = append(solutions, paths...)
					}
				}
			}
		}
	}
	return solutions
}

// helper functions

// CheckDependencies returns dependent modules
func CheckDependencies(module *app.FybrikModule, moduleMap map[string]*app.FybrikModule) ([]*app.FybrikModule, []string) {
	var found []*app.FybrikModule
	var missing []string
	for _, dependency := range module.Spec.Dependencies {
		if dependency.Type != app.Module {
			continue
		}
		if moduleMap[dependency.Name] == nil {
			missing = append(missing, dependency.Name)
		} else {
			found = append(found, moduleMap[dependency.Name])
			additionalDependencies, notFound := CheckDependencies(moduleMap[dependency.Name], moduleMap)
			found = append(found, additionalDependencies...)
			missing = append(missing, notFound...)
		}
	}
	return found, missing
}

// SupportsDependencies checks whether the module supports the dependency requirements
func SupportsDependencies(module *app.FybrikModule, moduleMap map[string]*app.FybrikModule) bool {
	// check dependencies
	_, missingModules := CheckDependencies(module, moduleMap)
	return len(missingModules) == 0
}

// GetDependencies returns dependencies of a selected module
func GetDependencies(module *app.FybrikModule, moduleMap map[string]*app.FybrikModule) ([]*app.FybrikModule, error) {
	dependencies, missingModules := CheckDependencies(module, moduleMap)
	if len(missingModules) > 0 {
		return dependencies, errors.New("Module " + module.Name + " has missing dependencies")
	}
	return dependencies, nil
}

// SupportsGovernanceActions checks whether the module supports the required governance actions
func supportsGovernanceActions(edge *Edge, actions []taxonomymodels.Action) bool {
	// Loop over the actions requested for the declared capability
	for _, action := range actions {
		// If any one of the actions is not supported, return false
		if !supportsGovernanceAction(edge, action) {
			return false
		}
	}
	return true // All actions supported
}

// supportsGovernanceAction checks whether the module supports the required governance action
func supportsGovernanceAction(edge *Edge, action taxonomymodels.Action) bool {
	// Loop over the data transforms (actions) performed by the module for this capability
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	for _, act := range capability.Actions {
		// TODO(shlomitk1): check for matching of additional fields declared by the module
		if act.Name == action.Name {
			return true
		}
	}
	return false // Action not supported by module
}

func match(source *app.InterfaceDetails, sink *app.InterfaceDetails) bool {
	if source == nil || sink == nil {
		return false
	}
	return source.DataFormat == sink.DataFormat && source.Protocol == sink.Protocol
}

// supportsSourceInterface indicates whether the source interface requirements are met.
func supportsSourceInterface(edge *Edge, source *Node) bool {
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	if capability.API != nil && source.Virtual {
		if match(&capability.API.InterfaceDetails, source.Connection) {
			return true
		}
	}
	for _, inter := range capability.SupportedInterfaces {
		// connection via Source
		if inter.Source != nil && match(inter.Source, source.Connection) {
			return true
		}
	}
	return false
}

// supportsSinkInterface indicates whether the sink interface requirements are met.
func supportsSinkInterface(edge *Edge, sink *Node) bool {
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	if capability.API != nil && sink.Virtual {
		if match(&capability.API.InterfaceDetails, sink.Connection) {
			return true
		}
	}
	for _, inter := range capability.SupportedInterfaces {
		if inter.Sink != nil && match(inter.Sink, sink.Connection) {
			return true
		}
	}
	return false
}

func allowCapability(item *DataInfo, capability string) bool {
	return item.Configuration.ConfigDecisions[capability].Deploy != v1.ConditionFalse
}

func validateModuleRestrictions(item *DataInfo, edge *Edge) bool {
	// TODO(shlomitk1): validate module restrictions
	return true
}

func validateClusterRestrictions(item *DataInfo, edge *ResolvedEdge, cluster multicluster.Cluster) bool {
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	if !validateClusterRestrictionsPerCapability(item, capability.Capability, cluster) {
		return false
	}
	if len(edge.Actions) > 0 {
		if !validateClusterRestrictionsPerCapability(item, Transform, cluster) {
			return false
		}
	}
	return true
}

func validateClusterRestrictionsPerCapability(item *DataInfo, capability string, cluster multicluster.Cluster) bool {
	restrictions := item.Configuration.ConfigDecisions[capability].DeploymentRestrictions[adminconfig.Clusters]
	if len(restrictions) == 0 {
		return true
	}
	clusterDetails, err := utils.StructToMap(&cluster)
	if err != nil {
		return false
	}
	for key, values := range restrictions {
		fields := strings.Split(key, ".")
		value, found, err := unstructured.NestedString(clusterDetails, fields...)
		if err != nil || !found {
			return false
		}
		if !utils.HasString(value, values) {
			return false
		}
	}
	return true
}

func createActionStructure(actions []taxonomymodels.Action) []app.SupportedAction {
	result := []app.SupportedAction{}
	for _, action := range actions {
		supportedAction := app.SupportedAction{Action: action}
		result = append(result, supportedAction)
	}
	return result
}
