// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package modules

import (
	"errors"

	"github.com/mesh-for-data/mesh-for-data/pkg/serde"

	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	pb "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
	"github.com/mesh-for-data/mesh-for-data/pkg/multicluster"
)

// DataDetails is the information received from the catalog connector
type DataDetails struct {
	// Name of the asset
	Name string
	// Interface is the protocol and format
	Interface app.InterfaceDetails
	// Geography is the geo-location of the asset
	Geography string
	// Connection is the connection details in raw format as received from the connector
	Connection serde.Arbitrary
	// Metadata
	Metadata *pb.DatasetMetadata
}

// DataInfo defines all the information about the given data set that comes from the m4dapplication spec and from the connectors.
type DataInfo struct {
	// Source connection details
	DataDetails *DataDetails
	// The path to Vault secret which holds the dataset credentials
	VaultSecretPath string
	// Pointer to the relevant data context in the M4D application spec
	Context *app.DataContext
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
	// Actions that the module will perform
	Actions []*pb.EnforcementAction
	// Geography where the module will be orchestrated
	Geo string
}

// TODO: Add function to check if module supports recurrence type

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
		AssetID:     item.Context.DataSetID,
		Module:      m.GetModule(),
		Args:        args,
		ClusterName: cluster,
	})
	for _, dep := range m.GetDependencies() {
		instances = append(instances, ModuleInstanceSpec{
			AssetID:     item.Context.DataSetID,
			Module:      dep,
			Args:        args,
			ClusterName: cluster,
		})
	}
	return instances
}

// SupportsGovernanceActions checks whether the module supports the required agovernance actions
func (m *Selector) SupportsGovernanceActions(module *app.M4DModule, actions []*pb.EnforcementAction) bool {
	// Check that the governance actions match
	for _, action := range actions {
		supportsAction := false
		for j := range module.Spec.Capabilities.Actions {
			transformation := &module.Spec.Capabilities.Actions[j]
			if transformation.ID == action.Id && transformation.Level == action.Level {
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
func (m *Selector) SupportsGovernanceAction(module *app.M4DModule, action *pb.EnforcementAction) bool {
	// Check that the governance actions match
	for j := range module.Spec.Capabilities.Actions {
		transformation := &module.Spec.Capabilities.Actions[j]
		if transformation.ID == action.Id && transformation.Level == action.Level {
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
// Write is done at target
func (m *Selector) SelectCluster(item DataInfo, clusters []multicluster.Cluster) (string, error) {
	geo := item.DataDetails.Geography
	if m.Flow == app.Read {
		geo = m.Geo
	} else if m.Flow == app.Copy && len(m.Actions) == 0 {
		geo = m.Geo
	}
	for _, cluster := range clusters {
		if cluster.Metadata.Region == geo {
			return cluster.Name, nil
		}
	}
	return "", errors.New(app.InvalidClusterConfiguration + "\nNo clusters have been found for running " + m.Module.Name + " in " + geo)
}

// Transforms a CatalogDatasetInfo into a DataDetails struct
// TODO Think about getting rid of one or the other and reuse
func CatalogDatasetToDataDetails(response *pb.CatalogDatasetInfo) (*DataDetails, error) {
	details := response.GetDetails()
	protocol, err := utils.GetProtocol(details)
	if err != nil {
		return nil, err
	}
	format, err := utils.GetDataFormat(details)
	if err != nil {
		return nil, err
	}

	connection := serde.NewArbitrary(details.DataStore)

	return &DataDetails{
		Name: details.Name,
		Interface: app.InterfaceDetails{
			Protocol:   protocol,
			DataFormat: format,
		},
		Geography:  details.Geo,
		Connection: *connection,
		Metadata:   details.Metadata,
	}, nil
}
