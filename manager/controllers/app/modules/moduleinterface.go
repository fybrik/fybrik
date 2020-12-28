// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"errors"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
)

// Transformations structure defines the governance actions to be taken for a specific flow
type Transformations struct {
	Allowed            bool
	EnforcementActions []pb.EnforcementAction
	Message            string
}

// DataInfo defines all the information about the given data set
type DataInfo struct {
	// Data asset unique identifier, not necessarily the same string appearing in the resource definition
	AssetID string
	// Processing geography
	Geo string
	// Application interface
	AppInterface *app.InterfaceDetails
	// Source connection details
	DataDetails *pb.DatasetDetails
	// Data asset credentials
	Credentials *pb.DatasetCredentials
	// Governance actions
	Actions map[app.ModuleFlow]Transformations
}

// ModuleInstanceSpec consists of the module spec and arguments
type ModuleInstanceSpec struct {
	Module      *app.M4DModule
	Args        *app.ModuleArguments
	AssetID     string
	ClusterName string
}

// Selector is responsible for finding an appropriate module
type Selector struct {
	Module       *app.M4DModule
	Dependencies []*app.M4DModule
	Message      string
	Flow         app.ModuleFlow
	Source       *app.InterfaceDetails
	Destination  *app.InterfaceDetails
	Actions      []pb.EnforcementAction
}

// GetModule returns the selected module
func (m *Selector) GetModule() *app.M4DModule {
	return m.Module
}

// GetDependencies returns dependencies of a selected module
func (m *Selector) GetDependencies() []*app.M4DModule {
	return m.Dependencies
}

// GetError returns an error message
func (m *Selector) GetError() string {
	return m.Message
}

// AddModuleInstances creates module instances for the selected module and its dependencies
func (m *Selector) AddModuleInstances(args *app.ModuleArguments, item DataInfo, cluster string) []ModuleInstanceSpec {
	instances := make([]ModuleInstanceSpec, 0)
	// append moduleinstances to the list
	instances = append(instances, ModuleInstanceSpec{
		AssetID:     item.AssetID,
		Module:      m.GetModule(),
		Args:        args,
		ClusterName: cluster,
	})
	for _, dep := range m.GetDependencies() {
		instances = append(instances, ModuleInstanceSpec{
			AssetID:     item.AssetID,
			Module:      dep,
			Args:        args,
			ClusterName: cluster,
		})
	}
	return instances
}

// SupportsGovernanceActions checks whether the module supports the required agovernance actions
func (m *Selector) SupportsGovernanceActions(module *app.M4DModule, actions []pb.EnforcementAction) bool {
	// Check that the governance actions match
	for i := range actions {
		action := &actions[i]
		supportsAction := false
		for j := range module.Spec.Capabilities.Actions {
			transformation := &module.Spec.Capabilities.Actions[j]
			if transformation.Id == action.Id && transformation.Level == action.Level {
				supportsAction = true
				break
			}
		}
		if !supportsAction {
			return false
		}
	}
	return true
}

// SupportsGovernanceAction checks whether the module supports the required agovernance action
func (m *Selector) SupportsGovernanceAction(module *app.M4DModule, action pb.EnforcementAction) bool {
	// Check that the governance actions match
	for j := range module.Spec.Capabilities.Actions {
		transformation := &module.Spec.Capabilities.Actions[j]
		if transformation.Id == action.Id && transformation.Level == action.Level {
			return true
		}
	}
	return false
}

// SupportsDependencies checks whether the module supports the dependency requirements
func (m *Selector) SupportsDependencies(module *app.M4DModule, moduleMap map[string]*app.M4DModule) bool {
	// check dependencies
	subModuleNames, errNames := CheckDependencies(module, moduleMap)
	if len(errNames) > 0 {
		m.Message += module.Name + " has missing dependencies: "
		for _, name := range errNames {
			m.Message += "\n" + name
		}
		m.Message += "\n"
		return false
	}
	m.Module = module.DeepCopy()
	for _, name := range subModuleNames {
		m.Dependencies = append(m.Dependencies, moduleMap[name])
	}
	return true
}

// SupportsInterface indicates whether the module supports interface requirements and dependencies
func (m *Selector) SupportsInterface(module *app.M4DModule) bool {
	// Check if the module supports the flow
	if !utils.SupportsFlow(module.Spec.Flows, m.Flow) {
		return false
	}
	// Check if the source and sink protocols requested are supported
	supportsInterface := false
	if m.Flow == app.Read {
		supportsInterface = module.Spec.Capabilities.API.DataFormat == m.Destination.DataFormat && module.Spec.Capabilities.API.Protocol == m.Destination.Protocol
	} else if m.Flow == app.Copy {
		for _, inter := range module.Spec.Capabilities.SupportedInterfaces {
			if inter.Flow != m.Flow {
				continue
			}
			if inter.Source.DataFormat != m.Source.DataFormat || inter.Source.Protocol != m.Source.Protocol {
				continue
			}
			if inter.Sink.DataFormat != m.Destination.DataFormat || inter.Sink.Protocol != m.Destination.Protocol {
				continue
			}
			supportsInterface = true
			break
		}
	}
	return supportsInterface
}

// SelectModule finds the module that fits the requirements
func (m *Selector) SelectModule(moduleMap map[string]*app.M4DModule) bool {
	m.Message = ""
	for _, module := range moduleMap {
		if !m.SupportsInterface(module) {
			continue
		}
		if !m.SupportsGovernanceActions(module, m.Actions) {
			continue
		}
		if !m.SupportsDependencies(module, moduleMap) {
			continue
		}
		return true
	}
	m.Message += string(m.Flow) + " : " + app.ModuleNotFound
	return false
}

// CheckDependencies returns dependent module names
func CheckDependencies(module *app.M4DModule, moduleMap map[string]*app.M4DModule) ([]string, []string) {
	var found []string
	var missing []string

	for _, dependency := range module.Spec.Dependencies {
		if dependency.Type != app.Module {
			continue
		}
		if moduleMap[dependency.Name] == nil {
			missing = append(missing, dependency.Name)
		} else {
			found = append(found, dependency.Name)
			names, notFound := CheckDependencies(moduleMap[dependency.Name], moduleMap)
			found = append(found, names...)
			missing = append(missing, notFound...)
		}
	}
	return found, missing
}

// SelectCluster chooses where the module runs
// Current logic:
// Read is done at target (processing geography)
// Copy is done at source when transformations are required, and at target - otherwise
func (m *Selector) SelectCluster(item DataInfo, clusters []multicluster.Cluster) (string, error) {
	var geo string
	if len(m.Actions) > 0 && m.Flow == app.Copy {
		geo = item.DataDetails.Geo
	} else {
		geo = item.Geo
	}
	for _, cluster := range clusters {
		if cluster.Metadata.Region == geo {
			return cluster.Name, nil
		}
	}
	return "", errors.New(app.InvalidClusterConfiguration + "\nNo clusters have been found for running " + m.Module.Name + " in " + geo)
}
