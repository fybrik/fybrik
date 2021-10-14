// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"errors"

	"fybrik.io/fybrik/pkg/serde"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/app/asset_details"
	config "fybrik.io/fybrik/manager/controllers/app/config_evaluator"
	"fybrik.io/fybrik/manager/controllers/utils"
	pb "fybrik.io/fybrik/pkg/connectors/protobuf"
	"fybrik.io/fybrik/pkg/multicluster"
	openapiclientmodels "fybrik.io/fybrik/pkg/taxonomy/model/base"
)

// DataInfo defines all the information about the given data set that comes from the fybrikapplication spec and from the connectors.
type DataInfo struct {
	// Source connection details
	DataDetails *asset_details.DataDetails
	// The path to Vault secret which holds the dataset credentials
	VaultSecretPath string
	// Pointer to the relevant data context in the Fybrik application spec
	Context *app.DataContext
	// Evaluated config policies
	Configuration config.EvaluatorOutput
	// Workload cluster
	WorkloadCluster multicluster.Cluster
	// Governance actions to perform on this asset
	Actions []openapiclientmodels.Action
}

// Selector is responsible for finding an appropriate module
type Selector struct {
	Module           *app.FybrikModule
	Dependencies     []*app.FybrikModule
	Message          string
	ModuleCapability *app.ModuleCapability
	Source           *app.InterfaceDetails
	Destination      *app.InterfaceDetails
	// Actions that the module will perform
	Actions []openapiclientmodels.Action
	// StorageAccountRegion for writing data
	StorageAccountRegion string
}

// TODO: Add function to check if module supports recurrence type
// TODO: In the future add support for plugins
// TODO: Add support for scope

// GetModule returns the selected module
func (m *Selector) GetModule() *app.FybrikModule {
	return m.Module
}

// GetDependencies returns dependencies of a selected module
func (m *Selector) GetDependencies() []*app.FybrikModule {
	return m.Dependencies
}

// GetError returns an error message
func (m *Selector) GetError() string {
	return m.Message
}

// SupportsGovernanceActions checks whether the module supports the required governance actions for the capability requested
func (m *Selector) SupportsGovernanceActions(module *app.FybrikModule, actions []openapiclientmodels.Action) bool {
	if m.ModuleCapability == nil {
		return false
	}
	// Loop over the actions requested for the declared capability
	for _, action := range actions {
		// If any one of the actions is not supported, return false
		if !m.SupportsGovernanceAction(module, action) {
			return false
		}
	}
	return true // All actions supported
}

// SupportsGovernanceAction checks whether the module supports the required governance action
func (m *Selector) SupportsGovernanceAction(module *app.FybrikModule, action openapiclientmodels.Action) bool {
	// Loop over the data transforms (actions) performed by the module for this capability
	for _, act := range m.ModuleCapability.Actions {
		if act.ID == action.Name {
			return true
		}
	}
	return false // Action not supported by module
}

// SupportsDependencies checks whether the module supports the dependency requirements
func (m *Selector) SupportsDependencies(module *app.FybrikModule, moduleMap map[string]*app.FybrikModule) bool {
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
func (m *Selector) SupportsInterface(module *app.FybrikModule, requestedCapability app.CapabilityType) bool {
	supportsInterface := false

	// Check if the module supports the capability
	for _, capability := range module.Spec.Capabilities {
		if capability.Capability != requestedCapability {
			continue
		}
		// Check if the source and sink protocols requested are supported
		if requestedCapability == app.Read {
			supportsInterface = capability.API.DataFormat == m.Destination.DataFormat && capability.API.Protocol == m.Destination.Protocol
			if supportsInterface {
				m.ModuleCapability = &capability
				return true
			}
		} else if requestedCapability == app.Copy {
			for _, inter := range capability.SupportedInterfaces {
				if inter.Source.DataFormat != m.Source.DataFormat || inter.Source.Protocol != m.Source.Protocol {
					continue
				}
				if inter.Sink.DataFormat != m.Destination.DataFormat || inter.Sink.Protocol != m.Destination.Protocol {
					continue
				}
				m.ModuleCapability = &capability
				supportsInterface = true
				break
			}
		}
	}
	return supportsInterface
}

// SelectModule finds the module that fits the requirements
func (m *Selector) SelectModule(moduleMap map[string]*app.FybrikModule, requestedCapability app.CapabilityType) bool {
	m.Message = ""
	for _, module := range moduleMap {
		if !m.SupportsInterface(module, requestedCapability) {
			m.Message = app.ModuleNotFound + " for " + string(requestedCapability) + "; requested interface is not supported"
			continue
		}
		if !m.SupportsGovernanceActions(module, m.Actions) {
			m.Message = app.ModuleNotFound + " for " + string(requestedCapability) + "; governance actions are not supported"
			continue
		}
		if !m.SupportsDependencies(module, moduleMap) {
			continue
		}
		return true
	}
	return false
}

// CheckDependencies returns dependent module names
func CheckDependencies(module *app.FybrikModule, moduleMap map[string]*app.FybrikModule) ([]string, []string) {
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

// Transforms a CatalogDatasetInfo into a DataDetails struct
// TODO Think about getting rid of one or the other and reuse
func CatalogDatasetToDataDetails(response *pb.CatalogDatasetInfo) (*asset_details.DataDetails, error) {
	details := response.GetDetails()
	if details == nil {
		return nil, errors.New("no metadata found for " + response.DatasetId)
	}
	format := details.DataFormat
	connection := serde.NewArbitrary(details.DataStore)
	protocol, err := utils.GetProtocol(details)

	return &asset_details.DataDetails{
		Name: details.Name,
		Interface: app.InterfaceDetails{
			Protocol:   protocol,
			DataFormat: format,
		},
		Geography:  details.Geo,
		Connection: *connection,
		Metadata:   details.Metadata,
	}, err
}
