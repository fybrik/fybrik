// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package modules

import (
	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
)

// Transformations structure defines the governance actions to be taken for a specific flow
type Transformations struct {
	Allowed            bool
	EnforcementActions []pb.EnforcementAction
	Message            string

	// In some cases copy is required to perform transformations at source
	// Temporary solution: in these cases mark copy actions as required until rules for transformations at data source are implemented in policy manager
	Required bool
}

// DataInfo defines all the information about the given data set
type DataInfo struct {
	// Data asset unique identifier, not necessarily the same string appearing in the resource definition
	AssetID string
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
	Module  *app.M4DModule
	Args    *app.ModuleArguments
	AssetID string
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
func (m *Selector) AddModuleInstances(args *app.ModuleArguments, item DataInfo) []ModuleInstanceSpec {
	instances := make([]ModuleInstanceSpec, 0)
	// append moduleinstances to the list
	instances = append(instances, ModuleInstanceSpec{
		AssetID: item.AssetID,
		Module:  m.GetModule(),
		Args:    args,
	})
	for _, dep := range m.GetDependencies() {
		instances = append(instances, ModuleInstanceSpec{
			AssetID: item.AssetID,
			Module:  dep,
			Args:    args,
		})
	}
	return instances
}

// SelectModule finds the module that fits the requirements
func (m *Selector) SelectModule(moduleMap map[string]*app.M4DModule) bool {
	m.Message = ""
	for _, module := range moduleMap {
		// Check if the module supports the flow
		if !utils.SupportsFlow(module.Spec.Flows, m.Flow) {
			continue
		}
		// Check if the source and sink protocols requested are supported
		supportsInterface := false
		var supportedInterfaceLog, requiredInterfaceLog string
		if m.Flow == app.Read {
			supportsInterface = module.Spec.Capabilities.API.DataFormat == m.Destination.DataFormat && module.Spec.Capabilities.API.Protocol == m.Destination.Protocol
			supportedInterfaceLog = "supports: " + string(module.Spec.Capabilities.API.DataFormat) + "," + string(module.Spec.Capabilities.API.Protocol) + "\n"
			requiredInterfaceLog = "requires: " + string(m.Destination.DataFormat) + "," + string(m.Destination.Protocol) + "\n"
		} else if m.Flow == app.Copy {
			for _, inter := range module.Spec.Capabilities.SupportedInterfaces {
				if inter.Flow != m.Flow {
					continue
				}
				supportedInterfaceLog = "supports: " + string(inter.Source.DataFormat) + "," + string(inter.Source.Protocol) + "\n"
				requiredInterfaceLog = "requires: " + string(m.Source.DataFormat) + "," + string(m.Source.Protocol) + "\n"
				supportedInterfaceLog += "supports: " + string(inter.Sink.DataFormat) + "," + string(inter.Sink.Protocol) + "\n"
				requiredInterfaceLog += "requires: " + string(m.Destination.DataFormat) + "," + string(m.Destination.Protocol) + "\n"

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
		if !supportsInterface {
			m.Message += module.Name + " does not support user interface:\n" + supportedInterfaceLog + requiredInterfaceLog
			continue
		}
		// Check that the governance actions match
		for i := range m.Actions {
			action := &m.Actions[i]
			supportsAction := false
			for j := range module.Spec.Capabilities.Actions {
				transformation := &module.Spec.Capabilities.Actions[j]
				if transformation.Id == action.Id && transformation.Level == action.Level {
					supportsAction = true
					break
				}
			}
			if !supportsAction {
				m.Message += module.Name + " does not support action " + action.Id + ";"
				continue
			}
		}
		// check dependencies
		subModuleNames, errNames := CheckDependencies(module, moduleMap)
		if len(errNames) > 0 {
			m.Message += module.Name + " has missing dependencies: "
			for _, name := range errNames {
				m.Message += "\n" + name
			}
			m.Message += "\n"
			continue
		}
		m.Module = module.DeepCopy()
		for _, name := range subModuleNames {
			m.Dependencies = append(m.Dependencies, moduleMap[name])
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
