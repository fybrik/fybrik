// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"reflect"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/rs/zerolog"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/infrastructure"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
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
	// Required governance actions to perform on this asset
	Actions []taxonomy.Action
	// Potential actions to be taken on storing this asset in a specific location
	StorageRequrements map[taxonomy.ProcessingLocation][]taxonomy.Action
}

// Environment defines the available resources (clusters, modules, storageAccounts)
// It also contains the results of queries to policy manager regarding writing data to storage accounts
type Environment struct {
	Log              zerolog.Logger
	Modules          map[string]*app.FybrikModule
	Clusters         []multicluster.Cluster
	StorageAccounts  []app.FybrikStorageAccount
	AttributeManager *infrastructure.AttributeManager
}

// Temporary hard-coded capability representing Actions of read/copy capability.
// A change to modules is requested to add "transform" as an additional capability to the existing read and copy modules.
const Transform = "transform"

// Node represents an access point to data (as a physical source/sink, or a virtual endpoint)
// A virtual endpoint is activated by the workload for read/write actions.
type Node struct {
	Connection *taxonomy.Interface
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
	DataPath []*ResolvedEdge
}

// find a solution for data plane orchestration
func (env *Environment) solve(item *DataInfo) (Solution, error) {
	env.Log.Trace().Str(logging.DATASETID, item.Context.DataSetID).Msg("Choose modules for dataset")
	solutions := env.FindPaths(item)
	// No data path found for the asset
	if len(solutions) == 0 {
		msg := "Deployed modules do not provide the functionality required to construct a data path"
		env.Log.Error().Str(logging.DATASETID, item.Context.DataSetID).Msg(msg)
		logging.LogStructure("Data Item Context", item, &env.Log, true, true)
		logging.LogStructure("Module Map", env.Modules, &env.Log, true, true)
		return Solution{}, errors.New(msg + " for " + item.Context.DataSetID)
	}
	return solutions[0], nil
}

// FindPaths finds all valid data paths between the data source and the workload
// First, data paths are constructed using interface connections, starting from data source.
// Then, transformations are added to the found paths, and clusters are matched to satisfy restrictions from admin config policies.
// Optimization is done by the shortest path (the paths are sorted by the length). To be changed in future versions.
func (env *Environment) FindPaths(item *DataInfo) []Solution {
	NodeFromAssetMetadata := Node{
		Connection: &taxonomy.Interface{
			Protocol:   item.DataDetails.Details.Connection.Name,
			DataFormat: item.DataDetails.Details.DataFormat,
		},
	}
	NodeFromAppRequirements := Node{Connection: &item.Context.Requirements.Interface}
	// find data paths of length up to DATAPATH_LIMIT from data source to the workload, not including transformations or branches
	bound, err := utils.GetDataPathMaxSize()
	if err != nil {
		env.Log.Warn().Str(logging.DATASETID, item.Context.DataSetID).Msg("a default value for DATAPATH_LIMIT will be used")
	}
	var solutions []Solution
	if item.Context.Flow != taxonomy.WriteFlow {
		solutions = env.findPathsWithinLimit(item, &NodeFromAssetMetadata, &NodeFromAppRequirements, bound)
	} else {
		solutions = env.findPathsWithinLimit(item, &NodeFromAppRequirements, &NodeFromAssetMetadata, bound)
		// reverse each solution to start with the application requirements, e.g. workload
		for ind := range solutions {
			for elementInd := 0; elementInd < len(solutions[ind].DataPath)/2; elementInd++ {
				reversedInd := len(solutions[ind].DataPath) - elementInd - 1
				solutions[ind].DataPath[elementInd], solutions[ind].DataPath[reversedInd] =
					solutions[ind].DataPath[reversedInd], solutions[ind].DataPath[elementInd]
			}
		}
	}
	// get valid solutions by extending data paths with transformations and selecting an appropriate cluster for each capability
	solutions = env.validSolutions(item, solutions)
	return solutions
}

// extend the received data paths with transformations and select an appropriate cluster for each capability in a data path
func (env *Environment) validSolutions(item *DataInfo, solutions []Solution) []Solution {
	validPaths := []Solution{}
	for ind := range solutions {
		if env.validate(item, solutions[ind]) {
			validPaths = append(validPaths, solutions[ind])
		}
	}
	return validPaths
}

// if new storage should be located, check the requirements:
// where storage can be allocated
// what additional actions to perform
func (env *Environment) validateStorageRequirements(item *DataInfo, element *ResolvedEdge) bool {
	var found bool
	var actions []taxonomy.Action

	if item.Context.Flow == taxonomy.WriteFlow && !item.Context.Requirements.FlowParams.IsNewDataSet {
		// no need to allocate storage, write destination is known
		return true
	}
	// select a storage account that
	// 1. satisfies admin config restrictions on storage
	// 2. writing to this storage is not forbidden by governance policies
	for accountInd := range env.StorageAccounts {
		// validate restrictions
		moduleCapability := element.Module.Spec.Capabilities[element.CapabilityIndex]
		account := &env.StorageAccounts[accountInd]
		if !env.validateRestrictions(
			item.Configuration.ConfigDecisions[moduleCapability.Capability].DeploymentRestrictions.StorageAccounts,
			&account.Spec, account.Name) {
			env.Log.Debug().Str(logging.DATASETID, item.Context.DataSetID).Msgf("storage account %s does not match the requirements",
				account.Name)
			continue
		}
		// query the policy manager whether WRITE operation is allowed
		// not relevant for new datasets
		if !item.Context.Requirements.FlowParams.IsNewDataSet {
			actions, found = item.StorageRequrements[account.Spec.Region]
			if !found {
				continue
			}
		}
		// add the selected storage account region
		element.StorageAccount = account.Spec
		break
	}
	if element.StorageAccount.Region == "" {
		env.Log.Debug().Str(logging.DATASETID, item.Context.DataSetID).Msg("Could not find a storage account, aborting data path construction")
		return false
	}
	// add WRITE actions
	element.Actions = actions
	return true
}

func (env *Environment) validate(item *DataInfo, solution Solution) bool {
	// start from data source, check supported actions and cluster restrictions
	requiredActions := item.Actions
	for ind := range solution.DataPath {
		element := solution.DataPath[ind]
		if !element.Edge.Sink.Virtual {
			if !env.validateStorageRequirements(item, element) {
				return false
			}
			// add WRITE actions
			if element.Actions != nil {
				requiredActions = append(requiredActions, element.Actions...)
			}
		}
	}
	// actions need to be handled somewhere on the path
	for ind := range solution.DataPath {
		element := solution.DataPath[ind]
		element.Actions = []taxonomy.Action{}
		moduleCapability := element.Module.Spec.Capabilities[element.CapabilityIndex]
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
		if !env.findCluster(item, element) {
			env.Log.Debug().Str(logging.DATASETID, item.Context.DataSetID).Msg("Could not find an available cluster for " +
				string(moduleCapability.Capability))
			return false
		}
	}
	// Are all actions supported by the capabilities in this data path?
	if len(requiredActions) > 0 {
		env.Log.Debug().Str(logging.DATASETID, item.Context.DataSetID).
			Msg("Not all governance actions are supported, aborting data path construction")
		return false
	}
	// Are all capabilities that need to be deployed supported in this data path?
	supportedCapabilities := map[taxonomy.Capability]bool{}
	for _, element := range solution.DataPath {
		supportedCapabilities[element.Module.Spec.Capabilities[element.CapabilityIndex].Capability] = true
	}
	for capability := range item.Configuration.ConfigDecisions {
		if item.Configuration.ConfigDecisions[capability].Deploy == adminconfig.StatusTrue {
			// check that it is supported
			if !supportedCapabilities[capability] {
				return false
			}
		}
	}
	return true
}

// find a cluster that satisfies the requirements
func (env *Environment) findCluster(item *DataInfo, element *ResolvedEdge) bool {
	for _, cluster := range env.Clusters {
		if env.validateClusterRestrictions(item, element, cluster) {
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
func (env *Environment) findPathsWithinLimit(item *DataInfo, source, sink *Node, n int) []Solution {
	solutions := []Solution{}
	for _, module := range env.Modules {
		for capabilityInd, capability := range module.Spec.Capabilities {
			// check if capability is allowed
			if !allowCapability(item, capability.Capability) {
				continue
			}
			edge := Edge{Module: module, CapabilityIndex: capabilityInd, Source: nil, Sink: nil}
			// check that the module + module capability satisfy the requirements from the admin config policies
			if !env.validateModuleRestrictions(item, &edge) {
				env.Log.Debug().Msgf("module %s does not satisfy requirements for capability %s", module.Name, capability.Capability)
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
				var path []*ResolvedEdge
				path = append(path, &ResolvedEdge{Edge: edge})
				solutions = append(solutions, Solution{DataPath: path})
			}
			// try to build data paths using the selected module capability
			if n > 1 {
				sources := []*taxonomy.Interface{}
				for _, inter := range capability.SupportedInterfaces {
					sources = append(sources, inter.Source)
				}
				if capability.API != nil {
					sources = append(sources, &taxonomy.Interface{
						Protocol:   capability.API.Connection.Name,
						DataFormat: capability.API.DataFormat})
				}
				for _, inter := range sources {
					node := Node{Connection: inter}
					// recursive call to find paths of length = n-1 using the supported source of the selected module capability
					paths := env.findPathsWithinLimit(item, source, &node, n-1)
					// add the selected module to the found paths
					for i := range paths {
						auxEdge := Edge{Module: module, CapabilityIndex: capabilityInd, Source: &node, Sink: sink}
						paths[i].DataPath = append(paths[i].DataPath, &ResolvedEdge{Edge: auxEdge})
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

func match(source, sink *taxonomy.Interface) bool {
	if source == nil || sink == nil {
		return false
	}
	return source.DataFormat == sink.DataFormat && source.Protocol == sink.Protocol
}

// supportsSourceInterface indicates whether the source interface requirements are met.
//nolint:dupl
func supportsSourceInterface(edge *Edge, source *Node) bool {
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	hasSources := false
	for _, inter := range capability.SupportedInterfaces {
		if inter.Source == nil {
			continue
		}
		hasSources = true
		// connection via Source
		if match(inter.Source, source.Connection) {
			return true
		}
	}
	if capability.API != nil && !hasSources {
		apiInterface := &taxonomy.Interface{Protocol: capability.API.Connection.Name, DataFormat: capability.API.DataFormat}
		if match(apiInterface, source.Connection) {
			// consumes data via API
			source.Virtual = true
			return true
		}
	}
	return false
}

// supportsSinkInterface indicates whether the sink interface requirements are met.
//nolint:dupl
func supportsSinkInterface(edge *Edge, sink *Node) bool {
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	hasSinks := false
	for _, inter := range capability.SupportedInterfaces {
		if inter.Sink == nil {
			continue
		}
		hasSinks = true
		if match(inter.Sink, sink.Connection) {
			return true
		}
	}
	if capability.API != nil && !hasSinks {
		apiInterface := &taxonomy.Interface{Protocol: capability.API.Connection.Name, DataFormat: capability.API.DataFormat}
		if match(apiInterface, sink.Connection) {
			// transfers data in-memory via API
			sink.Virtual = true
			return true
		}
	}
	return false
}

func allowCapability(item *DataInfo, capability taxonomy.Capability) bool {
	return item.Configuration.ConfigDecisions[capability].Deploy != adminconfig.StatusFalse
}

func (env *Environment) validateModuleRestrictions(item *DataInfo, edge *Edge) bool {
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	moduleSpec := edge.Module.Spec
	restrictions := item.Configuration.ConfigDecisions[capability.Capability].DeploymentRestrictions.Modules
	oldPrefix := "capabilities."
	newPrefix := oldPrefix + strconv.Itoa(edge.CapabilityIndex) + "."
	for i := range restrictions {
		if strings.Contains(restrictions[i].Property, oldPrefix) && !strings.Contains(restrictions[i].Property, newPrefix) {
			restrictions[i].Property = strings.Replace(restrictions[i].Property, oldPrefix,
				newPrefix, 1)
		}
	}
	return env.validateRestrictions(restrictions, &moduleSpec, "")
}

func (env *Environment) validateClusterRestrictions(item *DataInfo, edge *ResolvedEdge, cluster multicluster.Cluster) bool {
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	if !env.validateClusterRestrictionsPerCapability(item, capability.Capability, cluster) {
		return false
	}
	if len(edge.Actions) > 0 {
		if !env.validateClusterRestrictionsPerCapability(item, Transform, cluster) {
			return false
		}
	}
	return true
}

func (env *Environment) validateClusterRestrictionsPerCapability(item *DataInfo, capability taxonomy.Capability,
	cluster multicluster.Cluster) bool {
	restrictions := item.Configuration.ConfigDecisions[capability].DeploymentRestrictions.Clusters
	return env.validateRestrictions(restrictions, &cluster, "")
}

// Validation of an object with respect to the admin config restrictions
//nolint:gocyclo
func (env *Environment) validateRestrictions(restrictions []adminconfig.Restriction, spec interface{}, instanceName string) bool {
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
		attributeObj := env.AttributeManager.GetAttribute(taxonomy.Attribute(restrict.Property), instanceName)
		if attributeObj != nil {
			value = attributeObj.Value
			found = true
		} else {
			fields := strings.Split(restrict.Property, ".")
			value, found, err = NestedFieldNoCopy(details, fields...)
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

func NestedFieldNoCopy(obj map[string]interface{}, fields ...string) (interface{}, bool, error) {
	var val interface{} = obj

	for _, field := range fields {
		if val == nil {
			return nil, false, nil
		}
		if reflect.TypeOf(val).Kind() == reflect.Slice {
			s := reflect.ValueOf(val)
			i, err := strconv.Atoi(field)
			if err != nil {
				return nil, false, nil
			}
			val = s.Index(i).Interface()
			continue
		}
		if m, ok := val.(map[string]interface{}); ok {
			val, ok = m[field]
			if !ok {
				return nil, false, nil
			}
		} else {
			return nil, false, nil
		}
	}
	return val, true, nil
}
