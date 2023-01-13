// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/rs/zerolog"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/datapath"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
)

// Temporary hard-coded capability representing Actions of read/copy capability.
// A change to modules is requested to add "transform" as an additional capability to the existing read and copy modules.
const Transform = "transform"

// component responsible for data path construction
type PathBuilder struct {
	Log   *zerolog.Logger
	Env   *datapath.Environment
	Asset *datapath.DataInfo
}

// find a solution for data plane orchestration
func (p *PathBuilder) solve() (datapath.Solution, error) {
	p.Log.Trace().Str(logging.DATASETID, p.Asset.Context.DataSetID).Msg("Choose modules for dataset")
	solutions := p.FindPaths()
	// No data path found for the asset
	if len(solutions) == 0 {
		msg := "Deployed modules do not provide the functionality required to construct a data path"
		p.Log.Error().Str(logging.DATASETID, p.Asset.Context.DataSetID).Msg(msg)
		logging.LogStructure("Data Item Context", p.Asset, p.Log, zerolog.TraceLevel, true, true)
		logging.LogStructure("Module Map", p.Env.Modules, p.Log, zerolog.TraceLevel, true, true)
		return datapath.Solution{}, errors.New(msg + " for " + p.Asset.Context.DataSetID)
	}
	return solutions[0], nil
}

// FindPaths finds all valid data paths between the data source and the workload
// First, data paths are constructed using interface connections, starting from data source.
// Then, transformations are added to the found paths, and clusters are matched to satisfy restrictions from admin config policies.
// Optimization is done by the shortest path (the paths are sorted by the length). To be changed in future versions.
func (p *PathBuilder) FindPaths() []datapath.Solution {
	nodeFromAssetMetadata := p.getAssetConnectionNode()
	nodeFromAppRequirements := p.getRequiredConnectionNode()

	// find data paths of length up to DATAPATH_LIMIT from data source to the workload, not including transformations or branches
	// If an error exists it is logged in LogEnvVariables and a default value is used
	bound, _ := environment.GetDataPathMaxSize()
	var solutions []datapath.Solution
	if p.Asset.Context.Flow != taxonomy.WriteFlow {
		solutions = p.findPathsWithinLimit(nodeFromAssetMetadata, nodeFromAppRequirements, bound)
	} else {
		solutions = p.findPathsWithinLimit(nodeFromAppRequirements, nodeFromAssetMetadata, bound)
		// reverse each solution to start with the application requirements, e.g. workload
		for ind := range solutions {
			solutions[ind].Reverse()
		}
	}
	// get valid solutions by extending data paths with transformations and selecting an appropriate cluster for each capability
	solutions = p.validSolutions(solutions)

	return solutions
}

// extend the received data paths with transformations and select an appropriate cluster for each capability in a data path
func (p *PathBuilder) validSolutions(solutions []datapath.Solution) []datapath.Solution {
	validPaths := []datapath.Solution{}
	for ind := range solutions {
		if p.validate(solutions[ind]) {
			validPaths = append(validPaths, solutions[ind])
		}
	}
	return validPaths
}

// if new storage should be located, check the requirements:
// where storage can be allocated
// what additional actions to perform
func (p *PathBuilder) validateStorageRequirements(element *datapath.ResolvedEdge) bool {
	var found bool
	var actions []taxonomy.Action

	if p.Asset.Context.Flow == taxonomy.WriteFlow && !p.Asset.Context.Requirements.FlowParams.IsNewDataSet {
		// no need to allocate storage, write destination is known
		return true
	}
	// select a storage account that
	// 1. satisfies admin config restrictions on storage
	// 2. writing to this storage is not forbidden by governance policies
	for accountInd := range p.Env.StorageAccounts {
		// validate restrictions
		moduleCapability := element.Module.Spec.Capabilities[element.CapabilityIndex]
		account := p.Env.StorageAccounts[accountInd]
		if !p.validateRestrictions(
			p.Asset.Configuration.ConfigDecisions[moduleCapability.Capability].DeploymentRestrictions.StorageAccounts,
			&account.Spec, account.Name) {
			p.Log.Debug().Str(logging.DATASETID, p.Asset.Context.DataSetID).Msgf("storage account %s does not match the requirements",
				account.Name)
			continue
		}
		// query the policy manager whether WRITE operation is allowed
		actions, found = p.Asset.StorageRequirements[account.Spec.Geography]
		if !found {
			continue
		}

		// add the selected storage account geography
		element.StorageAccount = account.Spec
		break
	}
	if element.StorageAccount.Geography == "" {
		p.Log.Debug().Str(logging.DATASETID, p.Asset.Context.DataSetID).Msg("Could not find a storage account, aborting data path construction")
		return false
	}
	// add WRITE actions
	element.Actions = actions
	return true
}

func (p *PathBuilder) validate(solution datapath.Solution) bool {
	// start from data source, check supported actions and cluster restrictions
	requiredActions := p.Asset.Actions
	for ind := range solution.DataPath {
		element := solution.DataPath[ind]
		if element.Edge.Sink != nil && !element.Edge.Sink.Virtual {
			if !p.validateStorageRequirements(element) {
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
		if !p.findCluster(element) {
			p.Log.Debug().Str(logging.DATASETID, p.Asset.Context.DataSetID).Msg("Could not find an available cluster for " +
				string(moduleCapability.Capability))
			return false
		}
	}
	// Are all actions supported by the capabilities in this data path?
	if len(requiredActions) > 0 {
		p.Log.Debug().Str(logging.DATASETID, p.Asset.Context.DataSetID).
			Msg("Not all governance actions are supported, aborting data path construction")
		return false
	}
	// Are all capabilities that need to be deployed supported in this data path?
	supportedCapabilities := map[taxonomy.Capability]bool{}
	for _, element := range solution.DataPath {
		supportedCapabilities[element.Module.Spec.Capabilities[element.CapabilityIndex].Capability] = true
	}
	for capability := range p.Asset.Configuration.ConfigDecisions {
		if p.Asset.Configuration.ConfigDecisions[capability].Deploy == adminconfig.StatusTrue {
			// check that it is supported
			if !supportedCapabilities[capability] {
				return false
			}
		}
	}
	return true
}

// find a cluster that satisfies the requirements
func (p *PathBuilder) findCluster(element *datapath.ResolvedEdge) bool {
	for _, cluster := range p.Env.Clusters {
		if p.validateClusterRestrictions(element, cluster) {
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
func (p *PathBuilder) findPathsWithinLimit(source, sink *datapath.Node, n int) []datapath.Solution {
	solutions := []datapath.Solution{}
	for _, module := range p.Env.Modules {
		for capabilityInd, capability := range module.Spec.Capabilities {
			// check if capability is allowed
			if !p.allowCapability(capability.Capability) {
				continue
			}
			edge := datapath.Edge{Module: module, CapabilityIndex: capabilityInd, Source: nil, Sink: nil}
			// check that the module + module capability satisfy the requirements from the admin config policies
			if !p.validateModuleRestrictions(&edge) {
				p.Log.Debug().Msgf("module %s does not satisfy requirements for capability %s", module.Name, capability.Capability)
				continue
			}
			// check whether the module supports the final destination
			if !supportsSinkInterface(&edge, sink) {
				p.Log.Debug().Msgf("module %s does not support sink requirements for capability %s", module.Name, capability.Capability)
				continue
			}

			edge.Sink = sink
			// if a module supports both source and sink interfaces, it's an end of the recursion
			if supportsSourceInterface(&edge, source) {
				edge.Source = source
				// found a path
				var path []*datapath.ResolvedEdge
				path = append(path, &datapath.ResolvedEdge{Edge: edge})
				solutions = append(solutions, datapath.Solution{DataPath: path})
			} else {
				p.Log.Debug().Msgf("module %s does not satisfy source requirements for capability %s", module.Name, capability.Capability)
			}
			// try to build data paths using the selected module capability
			if n > 1 {
				sources := []*taxonomy.Interface{}
				for _, inter := range capability.SupportedInterfaces {
					if inter.Source != nil {
						sources = append(sources, inter.Source)
					}
				}
				if (len(sources) == 0) && (capability.API != nil) {
					sources = append(sources, &taxonomy.Interface{
						Protocol:   capability.API.Connection.Name,
						DataFormat: capability.API.DataFormat})
				}
				for _, inter := range sources {
					node := datapath.Node{Connection: inter}
					// recursive call to find paths of length = n-1 using the supported source of the selected module capability
					paths := p.findPathsWithinLimit(source, &node, n-1)
					// add the selected module to the found paths
					for i := range paths {
						auxEdge := datapath.Edge{Module: module, CapabilityIndex: capabilityInd, Source: &node, Sink: sink}
						paths[i].DataPath = append(paths[i].DataPath, &datapath.ResolvedEdge{Edge: auxEdge})
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
func CheckDependencies(module *fapp.FybrikModule, moduleMap map[string]*fapp.FybrikModule) ([]*fapp.FybrikModule, []string) {
	var found []*fapp.FybrikModule
	var missing []string
	for _, dependency := range module.Spec.Dependencies {
		if dependency.Type != fapp.Module {
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
func SupportsDependencies(module *fapp.FybrikModule, moduleMap map[string]*fapp.FybrikModule) bool {
	// check dependencies
	_, missingModules := CheckDependencies(module, moduleMap)
	return len(missingModules) == 0
}

// GetDependencies returns dependencies of a selected module
func GetDependencies(module *fapp.FybrikModule, moduleMap map[string]*fapp.FybrikModule) ([]*fapp.FybrikModule, error) {
	dependencies, missingModules := CheckDependencies(module, moduleMap)
	if len(missingModules) > 0 {
		return dependencies, errors.New("Module " + module.Name + " has missing dependencies")
	}
	return dependencies, nil
}

// supportsGovernanceAction checks whether the module supports the required governance action
func supportsGovernanceAction(edge *datapath.Edge, action taxonomy.Action) bool {
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
	if source.Protocol != sink.Protocol {
		return false
	}
	// an empty DataFormat value is not checked
	// either a module supports any format, or any format can be selected (no requirements)
	if source.DataFormat != "" && sink.DataFormat != "" && source.DataFormat != sink.DataFormat {
		return false
	}
	return true
}

// supportsSourceInterface indicates whether the source interface requirements are met.
//nolint:dupl
func supportsSourceInterface(edge *datapath.Edge, sourceNode *datapath.Node) bool {
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	hasSources := false
	for _, inter := range capability.SupportedInterfaces {
		if inter.Source == nil {
			continue
		}
		hasSources = true
		// connection via Source
		if sourceNode != nil && match(inter.Source, sourceNode.Connection) {
			return true
		}
	}
	if sourceNode != nil && capability.API != nil && !hasSources {
		apiInterface := &taxonomy.Interface{Protocol: capability.API.Connection.Name, DataFormat: capability.API.DataFormat}
		if match(apiInterface, sourceNode.Connection) {
			// consumes data via API
			sourceNode.Virtual = true
			return true
		}
	}
	if !hasSources && (sourceNode == nil) {
		return true
	}
	return false
}

// supportsSinkInterface indicates whether the sink interface requirements are met.
//nolint:dupl
func supportsSinkInterface(edge *datapath.Edge, sinkNode *datapath.Node) bool {
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	hasSinks := false
	for _, inter := range capability.SupportedInterfaces {
		if inter.Sink == nil {
			continue
		}
		hasSinks = true
		if sinkNode != nil && match(inter.Sink, sinkNode.Connection) {
			return true
		}
	}
	if sinkNode != nil && capability.API != nil && !hasSinks {
		apiInterface := &taxonomy.Interface{Protocol: capability.API.Connection.Name, DataFormat: capability.API.DataFormat}
		if match(apiInterface, sinkNode.Connection) {
			// transfers data in-memory via API
			sinkNode.Virtual = true
			return true
		}
	}
	if !hasSinks && (sinkNode == nil) {
		return true
	}
	return false
}

func (p *PathBuilder) getAssetConnectionNode() *datapath.Node {
	var protocol taxonomy.ConnectionType
	var dataFormat taxonomy.DataFormat
	// If the connection name is empty, the default protocol is s3.
	if p.Asset.DataDetails == nil || p.Asset.DataDetails.Details.Connection.Name == "" {
		protocol = utils.GetDefaultConnectionType()
	} else {
		protocol = p.Asset.DataDetails.Details.Connection.Name
		dataFormat = p.Asset.DataDetails.Details.DataFormat
	}
	return &datapath.Node{
		Connection: &taxonomy.Interface{
			Protocol:   protocol,
			DataFormat: dataFormat,
		},
	}
}

func (p *PathBuilder) getRequiredConnectionNode() *datapath.Node {
	if p.Asset.Context.Requirements.Interface == nil {
		return nil
	}
	return &datapath.Node{Connection: p.Asset.Context.Requirements.Interface}
}

func (p *PathBuilder) allowCapability(capability taxonomy.Capability) bool {
	return p.Asset.Configuration.ConfigDecisions[capability].Deploy != adminconfig.StatusFalse
}

func (p *PathBuilder) validateModuleRestrictions(edge *datapath.Edge) bool {
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	moduleSpec := edge.Module.Spec
	restrictions := []adminconfig.Restriction{}
	oldPrefix := "capabilities."
	newPrefix := oldPrefix + strconv.Itoa(edge.CapabilityIndex) + "."
	for i := range p.Asset.Configuration.ConfigDecisions[capability.Capability].DeploymentRestrictions.Modules {
		restrict := p.Asset.Configuration.ConfigDecisions[capability.Capability].DeploymentRestrictions.Modules[i]
		restrict.Property = strings.Replace(restrict.Property, oldPrefix, newPrefix, 1)
		restrictions = append(restrictions, restrict)
	}
	return p.validateRestrictions(restrictions, &moduleSpec, edge.Module.Name)
}

func (p *PathBuilder) validateClusterRestrictions(edge *datapath.ResolvedEdge, cluster multicluster.Cluster) bool {
	capability := edge.Module.Spec.Capabilities[edge.CapabilityIndex]
	if !p.validateClusterRestrictionsPerCapability(capability.Capability, cluster) {
		return false
	}
	if len(edge.Actions) > 0 {
		if !p.validateClusterRestrictionsPerCapability(Transform, cluster) {
			return false
		}
	}
	return true
}

func (p *PathBuilder) validateClusterRestrictionsPerCapability(capability taxonomy.Capability,
	cluster multicluster.Cluster) bool {
	restrictions := p.Asset.Configuration.ConfigDecisions[capability].DeploymentRestrictions.Clusters
	return p.validateRestrictions(restrictions, &cluster, cluster.Name)
}

// Validation of an object with respect to the admin config restrictions
func (p *PathBuilder) validateRestrictions(restrictions []adminconfig.Restriction, spec interface{}, instanceName string) bool {
	for _, restrict := range restrictions {
		if !restrict.SatisfiedByResource(p.Env.AttributeManager, spec, instanceName) {
			return false
		}
	}
	return true
}
