// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"emperror.dev/errors"
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	config "fybrik.io/fybrik/manager/controllers/app/config"
	"fybrik.io/fybrik/manager/controllers/app/dataset"
	"fybrik.io/fybrik/manager/controllers/app/modules"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/serde"
	openapiclientmodels "fybrik.io/fybrik/pkg/taxonomy/model/base"
	v1 "k8s.io/api/core/v1"
)

// DataInfo defines all the information about the given data set that comes from the fybrikapplication spec and from the connectors.
type DataInfo struct {
	// Source connection details
	DataDetails *dataset.DataDetails
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

func (p *PlotterGenerator) SelectModules(item DataInfo, appContext *app.FybrikApplication) (map[app.CapabilityType]*modules.Selector, error) {
	selectors := map[app.CapabilityType]*modules.Selector{}
	var readSelector, copySelector *modules.Selector
	// read flow, copy is not required (either allowed or forbidden)
	if item.Configuration.ConfigDecisions[app.Read].Deploy == v1.ConditionTrue &&
		item.Configuration.ConfigDecisions[app.Copy].Deploy != v1.ConditionTrue {
		// select a read module that supports user interface requirements, data source interface and all required actions
		p.Log.Info("Looking for a data path with no copy")
		selector := &modules.Selector{
			Destination:  &item.Context.Requirements.Interface,
			Actions:      item.Actions,
			Source:       &item.DataDetails.Interface,
			Dependencies: []*app.FybrikModule{},
			Module:       nil,
			Message:      "",
		}
		if selector.SelectModule(p.Modules, app.Read) {
			// check deployment cluster
			if len(item.Actions) > 0 {
				clusters := utils.Intersection(item.Configuration.ConfigDecisions[app.Read].Clusters, item.Configuration.ConfigDecisions[app.Transform].Clusters)
				readDecision := item.Configuration.ConfigDecisions[app.Read]
				readDecision.Clusters = clusters
				item.Configuration.ConfigDecisions[app.Read] = readDecision
			}
			if len(item.Configuration.ConfigDecisions[app.Read].Clusters) > 0 {
				readSelector = selector
				item.Configuration.ConfigDecisions[app.Copy] = config.Decision{Deploy: v1.ConditionFalse}
			}
		}
		// read flow, no read module was selected
		// if copy is forbidden - report an error
		if readSelector == nil && item.Configuration.ConfigDecisions[app.Copy].Deploy == v1.ConditionFalse {
			p.Log.Info("Could not select a read module for " + item.Context.DataSetID + " : " + selector.GetError())
			return nil, errors.New(selector.GetError())
		}
	}

	// read + copy
	if item.Configuration.ConfigDecisions[app.Read].Deploy == v1.ConditionTrue && readSelector == nil {
		p.Log.Info("Copy is required in addition to the read module")
		// remove source interface support and action support
		readSelector = &modules.Selector{
			Destination:  &item.Context.Requirements.Interface,
			Actions:      []openapiclientmodels.Action{},
			Source:       nil,
			Dependencies: []*app.FybrikModule{},
			Module:       nil,
			Message:      "",
		}
		if !readSelector.SelectModule(p.Modules, app.Read) {
			p.Log.Info("Could not select a read module for " + item.Context.DataSetID + "; no module supports the requested interface")
			return nil, errors.New(readSelector.GetError())
		}

		interfaces := GetSupportedReadSources(readSelector.GetModule())
		actionsOnRead := []openapiclientmodels.Action{}
		actionsOnCopy := []openapiclientmodels.Action{}
		if len(item.Actions) > 0 {
			readAndTransformClusters := utils.Intersection(item.Configuration.ConfigDecisions[app.Read].Clusters, item.Configuration.ConfigDecisions[app.Transform].Clusters)
			copyAndTransformClusters := utils.Intersection(item.Configuration.ConfigDecisions[app.Copy].Clusters, item.Configuration.ConfigDecisions[app.Transform].Clusters)
			if len(readAndTransformClusters) == 0 {
				actionsOnCopy = item.Actions
			} else {
				// ensure that copy + read support all needed actions
				// actions that the read module can not perform are required to be done during copy
				for _, action := range item.Actions {
					if !readSelector.SupportsGovernanceAction(readSelector.GetModule(), action) {
						actionsOnCopy = append(actionsOnCopy, action)
					} else {
						actionsOnRead = append(actionsOnRead, action)
					}
				}
				readSelector.Actions = actionsOnRead
			}
			// WRITE actions
			operation := new(openapiclientmodels.PolicyManagerRequestAction)
			operation.SetActionType(openapiclientmodels.WRITE)
			operation.SetDestination(item.WorkloadCluster.Metadata.Region)
			actions, err := LookupPolicyDecisions(item.Context.DataSetID, p.PolicyManager, appContext, operation)
			actionsOnCopy = append(actionsOnCopy, actions...)
			if err != nil {
				return nil, err
			}
			if len(copyAndTransformClusters) == 0 && len(actionsOnCopy) > 0 {
				// copy module can not perform governance actions
				return nil, errors.New("Violation of a policy on deployment clusters for running governance actions")
			}
		}
		// select a module that supports COPY, supports required governance actions, has the required dependencies, with source in module sources and a non-empty intersection between requested and supported interfaces.
		for _, copyDest := range interfaces {
			selector := &modules.Selector{
				Source:               &item.DataDetails.Interface,
				Actions:              actionsOnCopy,
				Destination:          copyDest,
				Dependencies:         make([]*app.FybrikModule, 0),
				Module:               nil,
				Message:              "",
				StorageAccountRegion: item.WorkloadCluster.Metadata.Region,
			}
			if selector.SelectModule(p.Modules, app.Copy) {
				copySelector = selector
				break
			}
		}
		if copySelector == nil {
			p.Log.Info("Could not select a copy module for " + item.Context.DataSetID)
			return nil, errors.New("Copy is required but the data path could not be constructed")
		}
	}

	// copy flow (ingest)
	if (item.Configuration.ConfigDecisions[app.Copy].Deploy == v1.ConditionTrue) && (item.Configuration.ConfigDecisions[app.Read].Deploy == v1.ConditionFalse) {
		for _, region := range p.StorageAccountRegions {
			operation := new(openapiclientmodels.PolicyManagerRequestAction)
			operation.SetActionType(openapiclientmodels.WRITE)
			operation.SetDestination(region)

			actionsOnCopy, err := LookupPolicyDecisions(item.Context.DataSetID, p.PolicyManager, appContext, operation)
			if err != nil && err.Error() == app.WriteNotAllowed {
				continue
			}
			if err != nil {
				return nil, err
			}
			selector := &modules.Selector{
				Source:               &item.DataDetails.Interface,
				Actions:              actionsOnCopy,
				Destination:          &item.Context.Requirements.Interface,
				Dependencies:         make([]*app.FybrikModule, 0),
				Module:               nil,
				Message:              "",
				StorageAccountRegion: region,
			}
			if selector.SelectModule(p.Modules, app.Copy) {
				copySelector = selector
				break
			}
		}
		if copySelector == nil {
			p.Log.Info("Could not select a copy module for " + item.Context.DataSetID)
			return nil, errors.New(string(app.Copy) + " : " + app.ModuleNotFound)
		}
	}
	selectors[app.Copy] = copySelector
	selectors[app.Read] = readSelector
	return selectors, nil
}

// GetSupportedReadSources returns a list of supported READ interfaces of a module
func GetSupportedReadSources(module *app.FybrikModule) []*app.InterfaceDetails {
	var list []*app.InterfaceDetails

	// Check if the module supports READ
	if hasCapability, caps := utils.GetModuleCapabilities(module, app.Read); hasCapability {
		for _, cap := range caps {
			// Collect the interface sources
			for _, inter := range cap.SupportedInterfaces {
				list = append(list, inter.Source)
			}
		}
	}
	return list
}

func actionsToArbitrary(actions []openapiclientmodels.Action) []serde.Arbitrary {
	result := []serde.Arbitrary{}
	for _, action := range actions {
		raw := serde.NewArbitrary(action)
		result = append(result, *raw)
	}
	return result
}
