// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"strconv"
	"strings"

	"emperror.dev/errors"
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/adminrules"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/policymanager"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// DataInfo defines all the information about the given data set that comes from the fybrikapplication spec and from the connectors.
type DataInfo struct {
	// Source connection details
	DataDetails *datacatalog.GetAssetResponse
	// Pointer to the relevant data context in the Fybrik application spec
	Context *app.DataContext
	// Evaluated config policies
	Configuration adminconfig.EvaluatorOutput
	// Workload cluster
	WorkloadCluster multicluster.Cluster
	// Governance actions to perform on this asset
	Actions []taxonomy.Action
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
	Actions        []taxonomy.Action
	Cluster        string
	StorageAccount app.FybrikStorageAccountSpec
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
	source := Node{
		Connection: &app.InterfaceDetails{
			Protocol:   item.DataDetails.Details.Connection.Name,
			DataFormat: item.DataDetails.Details.DataFormat,
		},
		Virtual: false,
	}
	// data sink, either a virtual endpoint in read scenarios, or a datastore as in ingest scenario
	destination := Node{Connection: &item.Context.Requirements.Interface, Virtual: (appContext.Spec.Selector.WorkloadSelector.Size() > 0)}
	// find data paths of length up to DATAPATH_LIMIT from data source to the workload, not including transformations or branches
	bound, err := utils.GetDataPathMaxSize()
	if err != nil {
		p.Log.Warn().Str(logging.DATASETID, item.Context.DataSetID).Msg("a default value for DATAPATH_LIMIT will be used")
	}
	solutions := p.findPathsWithinLimit(item, &source, &destination, bound)
	// get valid solutions by extending data paths with transformations and selecting an appropriate cluster for each capability
	solutions = p.validSolutions(item, solutions, appContext)
	return solutions
}

// extend the received data paths with transformations and select an appropriate cluster for each capability in a data path
func (p *PlotterGenerator) validSolutions(item *DataInfo, solutions []Solution, application *app.FybrikApplication) []Solution {
	validPaths := []Solution{}
	for ind := range solutions {
		if p.validate(item, solutions[ind], application) {
			validPaths = append(validPaths, solutions[ind])
		}
	}
	return validPaths
}

func (p *PlotterGenerator) validate(item *DataInfo, solution Solution, application *app.FybrikApplication) bool {
	// start from data source, check supported actions and cluster restrictions
	appContext := ApplicationContext{Application: application, Log: p.Log}
	requiredActions := item.Actions
	for ind := range solution.DataPath {
		element := &solution.DataPath[ind]
		element.Actions = []taxonomy.Action{}
		moduleCapability := element.Module.Spec.Capabilities[element.CapabilityIndex]
		if !element.Edge.Sink.Virtual {
			// storage is required, plus more actions on copy may be needed
			isAccountFound := false
			for _, account := range p.StorageAccounts {
				// validate restrictions
				if !p.validateRestrictions(item.Configuration.ConfigDecisions[moduleCapability.Capability].DeploymentRestrictions.StorageAccounts, &account.Spec, account.Name) {
					p.Log.Debug().Str(logging.DATASETID, item.Context.DataSetID).Msgf("storage account %s does not match the requirements", account.Name)
					continue
				}
				// query the policy manager whether WRITE operation is allowed
				operation := new(policymanager.RequestAction)
				operation.ActionType = policymanager.WRITE
				operation.Destination = string(account.Spec.Region)
				operation.ProcessingLocation = account.Spec.Region
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
				element.StorageAccount = account.Spec
				isAccountFound = true
			}
			if !isAccountFound {
				p.Log.Debug().Str(logging.DATASETID, item.Context.DataSetID).Msg("Could not find a storage account, aborting data path construction")
				return false
			}
		}
		// read actions need to be handled somewhere on the path
		unsupported := []taxonomy.Action{}
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
		if !p.findCluster(item, element) {
			p.Log.Debug().Str(logging.DATASETID, item.Context.DataSetID).Msg("Could not find an available cluster for " + string(moduleCapability.Capability))
			return false
		}
	}
	// Are all actions supported by the capabilities in this data path?
	if len(requiredActions) > 0 {
		p.Log.Debug().Str(logging.DATASETID, item.Context.DataSetID).Msg("Not all governance actions are supported, aborting data path construction")
		return false
	}
	// Are all capabilities that need to be deployed supported in this data path?
	supportedCapabilities := map[taxonomy.Capability]bool{}
	for _, element := range solution.DataPath {
		supportedCapabilities[element.Module.Spec.Capabilities[element.CapabilityIndex].Capability] = true
	}
	for capability, decision := range item.Configuration.ConfigDecisions {
		if decision.Deploy == adminrules.StatusTrue {
			// check that it is supported
			if !supportedCapabilities[capability] {
				return false
			}
		}
	}
	return true
}

// find a cluster that satisfies the requirements
func (p *PlotterGenerator) findCluster(item *DataInfo, element *ResolvedEdge) bool {
	for _, cluster := range p.Clusters {
		if p.validateClusterRestrictions(item, element, cluster) {
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
func supportsGovernanceActions(edge *Edge, actions []taxonomy.Action) bool {
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
func supportsGovernanceAction(edge *Edge, action taxonomy.Action) bool {
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
		apiInterface := &app.InterfaceDetails{Protocol: capability.API.Connection.Name, DataFormat: capability.API.DataFormat}
		if match(apiInterface, source.Connection) {
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
		apiInterface := &app.InterfaceDetails{Protocol: capability.API.Connection.Name, DataFormat: capability.API.DataFormat}
		if match(apiInterface, sink.Connection) {
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

func allowCapability(item *DataInfo, capability taxonomy.Capability) bool {
	return item.Configuration.ConfigDecisions[capability].Deploy != adminrules.StatusFalse
}

func validateModuleRestrictions(item *DataInfo, edge *Edge) bool {
	// TODO(shlomitk1): validate module restrictions
	return true
}

func (p *PlotterGenerator) validateClusterRestrictions(item *DataInfo, edge *ResolvedEdge, cluster multicluster.Cluster) bool {
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	if !p.validateClusterRestrictionsPerCapability(item, capability.Capability, cluster) {
		return false
	}
	if len(edge.Actions) > 0 {
		if !p.validateClusterRestrictionsPerCapability(item, Transform, cluster) {
			return false
		}
	}
	return true
}

func (p *PlotterGenerator) validateClusterRestrictionsPerCapability(item *DataInfo, capability taxonomy.Capability, cluster multicluster.Cluster) bool {
	restrictions := item.Configuration.ConfigDecisions[capability].DeploymentRestrictions.Clusters
	return p.validateRestrictions(restrictions, &cluster, "")
}

// Validation of an object with respect to the admin config restrictions
func (p *PlotterGenerator) validateRestrictions(restrictions []adminrules.Restriction, spec interface{}, instanceName string) bool {
	if len(restrictions) == 0 {
		return true
	}
	details, err := utils.StructToMap(spec)
	if err != nil {
		return false
	}
	for _, restrict := range restrictions {
		var value interface{}
		var err error
		var found bool
		// infrastructure attribute or a property in the spec?
		attributeObj := p.AttributeManager.GetAttribute(taxonomy.Attribute(restrict.Property), instanceName)
		if attributeObj != nil {
			value = attributeObj.Value
			found = true
		} else {
			fields := strings.Split(restrict.Property, ".")
			value, found, err = unstructured.NestedFieldNoCopy(details, fields...)
		}
		if err != nil || !found {
			return false
		}
		if restrict.Range != nil {
			var numericVal int
			switch value := value.(type) {
			case int64:
				numericVal = int(value)
			case float64:
				numericVal = int(value)
			case int:
				numericVal = value
			case string:
				if numericVal, err = strconv.Atoi(value); err != nil {
					return false
				}
			}
			if restrict.Range.Max > 0 && numericVal > restrict.Range.Max {
				return false
			}
			if restrict.Range.Min > 0 && numericVal < restrict.Range.Min {
				return false
			}
		} else if len(restrict.Values) != 0 {
			if !utils.HasString(value.(string), restrict.Values) {
				return false
			}
		}
	}
	return true
}
