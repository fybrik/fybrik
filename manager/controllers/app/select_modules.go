// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"emperror.dev/errors"
	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/app/assetmetadata"
	"fybrik.io/fybrik/manager/controllers/app/modules"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/multicluster"
	taxonomymodels "fybrik.io/fybrik/pkg/taxonomy/model/policymanager/base"
	v1 "k8s.io/api/core/v1"
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

// SelectModules selects the specific modules and the relevant capabilities in order to construct a data flow for the given asset
// It restricts the choice for the deployment clusters and selects the suitable storage account region for the copy
// Algorithm:
// For a read scenario without the requirement for the copy, try to find a single read module. If not possible - find read + copy.
// For a read scenario with the requirement for the copy - construct a read + copy flow
// For a copy scenario - construct a copy flow.
func (p *PlotterGenerator) SelectModules(item *DataInfo, appContext *app.FybrikApplication) (map[app.CapabilityType]*modules.Selector, error) {
	selectors := map[app.CapabilityType]*modules.Selector{}
	var err error
	// read flow, copy is not required (either allowed or forbidden)
	if item.Configuration.ConfigDecisions[app.Read].Deploy == v1.ConditionTrue &&
		item.Configuration.ConfigDecisions[app.Copy].Deploy != v1.ConditionTrue {
		p.Log.Info("Looking for a data path with no copy")
		selectors, err = p.buildReadFlow(item, appContext)
		if err == nil {
			// read flow is ready
			return selectors, nil
		}
		// read flow, no read module was selected
		// if copy is forbidden - report an error
		if item.Configuration.ConfigDecisions[app.Copy].Deploy == v1.ConditionFalse {
			p.Log.Info("Could not build a read flow for " + item.Context.DataSetID + " : " + err.Error())
			return selectors, err
		}
	}
	// read + copy
	if item.Configuration.ConfigDecisions[app.Read].Deploy == v1.ConditionTrue {
		p.Log.Info("Copy is required in addition to the read module")
		return p.buildReadFlowWithCopy(item, appContext)
	}

	// copy flow (ingest)
	if (item.Configuration.ConfigDecisions[app.Copy].Deploy == v1.ConditionTrue) && (item.Configuration.ConfigDecisions[app.Read].Deploy == v1.ConditionFalse) {
		return p.buildCopyFlow(item, appContext)
	}
	// no data flow has been constructed
	return selectors, errors.New("Failed to generate a plotter: no capabilities are required")
}

func (p *PlotterGenerator) buildReadFlow(item *DataInfo, appContext *app.FybrikApplication) (map[app.CapabilityType]*modules.Selector, error) {
	// choose a read module that supports user interface requirements, data source interface and all required actions
	selectors := map[app.CapabilityType]*modules.Selector{app.Read: nil, app.Copy: nil}
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
			clusters := utils.Intersection(
				item.Configuration.ConfigDecisions[app.Read].DeploymentRestrictions.Clusters,
				item.Configuration.ConfigDecisions[app.Transform].DeploymentRestrictions.Clusters,
			)
			readDecision := item.Configuration.ConfigDecisions[app.Read]
			readDecision.DeploymentRestrictions.Clusters = clusters
			item.Configuration.ConfigDecisions[app.Read] = readDecision
		}
		if len(item.Configuration.ConfigDecisions[app.Read].DeploymentRestrictions.Clusters) > 0 {
			selectors[app.Read] = selector
			return selectors, nil
		}
		return selectors, errors.New("No deployment clusters for read are available")
	}
	return selectors, errors.New(selector.GetError())
}

func (p *PlotterGenerator) buildReadFlowWithCopy(item *DataInfo, appContext *app.FybrikApplication) (map[app.CapabilityType]*modules.Selector, error) {
	selectors := map[app.CapabilityType]*modules.Selector{app.Read: nil, app.Copy: nil}
	// find a read module that supports api requirements
	// not looking for source interface support and action support
	// TODO(shlomitk1): consider multiple options for read module compatibility
	readSelector := &modules.Selector{
		Destination:  &item.Context.Requirements.Interface,
		Actions:      []taxonomymodels.Action{},
		Source:       nil,
		Dependencies: []*app.FybrikModule{},
		Module:       nil,
		Message:      "",
	}
	if !readSelector.SelectModule(p.Modules, app.Read) {
		p.Log.Info("Could not find a read module for " + item.Context.DataSetID + "; no module supports the requested interface")
		return selectors, errors.New(readSelector.GetError())
	}
	// find a copy module that matches the selected read module (common interface and action support)
	interfaces := GetSupportedReadSources(readSelector.GetModule())
	actionsOnRead := []taxonomymodels.Action{}
	actionsOnCopy := []taxonomymodels.Action{}
	if len(item.Actions) > 0 {
		// intersect deployment clusters for read+transform, copy+transform
		readAndTransformClusters := utils.Intersection(
			item.Configuration.ConfigDecisions[app.Read].DeploymentRestrictions.Clusters,
			item.Configuration.ConfigDecisions[app.Transform].DeploymentRestrictions.Clusters,
		)
		copyAndTransformClusters := utils.Intersection(
			item.Configuration.ConfigDecisions[app.Copy].DeploymentRestrictions.Clusters,
			item.Configuration.ConfigDecisions[app.Transform].DeploymentRestrictions.Clusters,
		)
		if len(readAndTransformClusters) == 0 {
			// read module can not run transformations because of the cluster restriction
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
		if len(actionsOnRead) > 0 {
			readDecision := item.Configuration.ConfigDecisions[app.Read]
			readDecision.DeploymentRestrictions.Clusters = readAndTransformClusters
			item.Configuration.ConfigDecisions[app.Read] = readDecision
		}
		if len(actionsOnCopy) > 0 {
			copyDecision := item.Configuration.ConfigDecisions[app.Copy]
			copyDecision.DeploymentRestrictions.Clusters = copyAndTransformClusters
			copyDecision.Deploy = v1.ConditionTrue
			item.Configuration.ConfigDecisions[app.Copy] = copyDecision
		}

		// WRITE actions that should be done by the copy module
		// TODO(shlomitk1): generalize the regions the temporary copy can be done to, currently assumes workload geography
		operation := new(taxonomymodels.PolicyManagerRequestAction)
		operation.SetActionType(taxonomymodels.WRITE)
		operation.SetDestination(item.WorkloadCluster.Metadata.Region)
		actions, err := LookupPolicyDecisions(item.Context.DataSetID, p.PolicyManager, appContext, operation)
		actionsOnCopy = append(actionsOnCopy, actions...)
		if err != nil {
			return selectors, err
		}
		if len(copyAndTransformClusters) == 0 && len(actionsOnCopy) > 0 {
			// copy module can not perform governance actions
			return selectors, errors.New("Violation of a policy on deployment clusters for running governance actions")
		}
	}
	// find a module that supports COPY, supports required governance actions, has the required dependencies, with source in module sources and a non-empty intersection between requested and supported interfaces.
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
			selectors[app.Read] = readSelector
			selectors[app.Copy] = selector
			// the flow is ready
			return selectors, nil
		}
	}
	p.Log.Info("Could not find a copy module for " + item.Context.DataSetID)
	return selectors, errors.New("Copy is required but the data path could not be constructed")
}

func (p *PlotterGenerator) buildCopyFlow(item *DataInfo, appContext *app.FybrikApplication) (map[app.CapabilityType]*modules.Selector, error) {
	selectors := map[app.CapabilityType]*modules.Selector{app.Read: nil, app.Copy: nil}
	// find a region for storage allocation
	// TODO(shlomitk1): prefer the workload cluster if specified
	for _, region := range p.StorageAccountRegions {
		operation := new(taxonomymodels.PolicyManagerRequestAction)
		operation.SetActionType(taxonomymodels.WRITE)
		operation.SetDestination(region)

		actionsOnCopy, err := LookupPolicyDecisions(item.Context.DataSetID, p.PolicyManager, appContext, operation)
		if err != nil && err.Error() == app.WriteNotAllowed {
			continue
		}
		if err != nil {
			return selectors, err
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
			// the flow is ready
			if len(actionsOnCopy) != 0 {
				copyDecision := item.Configuration.ConfigDecisions[app.Copy]
				copyDecision.DeploymentRestrictions.Clusters = utils.Intersection(
					item.Configuration.ConfigDecisions[app.Copy].DeploymentRestrictions.Clusters,
					item.Configuration.ConfigDecisions[app.Transform].DeploymentRestrictions.Clusters,
				)
				item.Configuration.ConfigDecisions[app.Copy] = copyDecision
			}
			selectors[app.Copy] = selector
			return selectors, nil
		}
	}
	p.Log.Info("Could not find a copy module for " + item.Context.DataSetID)
	return nil, errors.New(string(app.Copy) + " : " + app.ModuleNotFound)
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

func createActionStructure(actions []taxonomymodels.Action) []app.SupportedAction {
	result := []app.SupportedAction{}
	for _, action := range actions {
		supportedAction := app.SupportedAction{Action: action}
		result = append(result, supportedAction)
	}
	return result
}
