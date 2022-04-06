// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"fmt"
	"strconv"
	"strings"

	appApi "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/app"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
)

type moduleAndCapability struct {
	module        *appApi.FybrikModule
	capability    *appApi.ModuleCapability
	capabilityIdx int
}

type DataPathCSP struct {
	problemData         *app.DataInfo
	env                 *app.Environment
	modulesCapabilities []moduleAndCapability      // An enumeration of allowed capabilities in all modules
	interfaceIdx        map[taxonomy.Interface]int // gives an index for each unique interface
	indicators          map[string]bool            // indicators used as part of the problem (to prevent redefinition)
	fzModel             *FlatZincModel
}

func NewDataPathCSP(problemData *app.DataInfo, env *app.Environment) *DataPathCSP {
	dpCSP := DataPathCSP{problemData: problemData, env: env, fzModel: NewFlatZincModel()}
	dpCSP.interfaceIdx = make(map[taxonomy.Interface]int)
	dataSetIntfc := taxonomy.Interface{
		Protocol:   dpCSP.problemData.DataDetails.Details.Connection.Name,
		DataFormat: dpCSP.problemData.DataDetails.Details.DataFormat,
	}
	dpCSP.addInterfaceToMap(dataSetIntfc)
	dpCSP.addInterfaceToMap(dpCSP.problemData.Context.Requirements.Interface)

	dpCSP.fzModel.AddHeaderComment("Encoding of modules and their capabilities:")
	modCapIdx := 1
	comment := ""
	for _, module := range env.Modules {
		for idx, capability := range module.Spec.Capabilities {
			modCap := moduleAndCapability{module, &module.Spec.Capabilities[idx], idx}
			if dpCSP.moduleCapabilityAllowedByRestrictions(modCap) {
				dpCSP.modulesCapabilities = append(dpCSP.modulesCapabilities, modCap)
				dpCSP.addModCapInterfacesToMap(modCap)
				comment = strconv.Itoa(modCapIdx)
				modCapIdx++
			} else {
				comment = "<forbidden>"
			}
			comment = comment + fmt.Sprintf(" - Module: %s, Capability: %d (%s)", module.Name, idx, capability.Capability)
			dpCSP.fzModel.AddHeaderComment(comment)
		}
	}

	dpCSP.fzModel.AddHeaderComment("Encoding of interfaces:")
	for intfc, intfcIdx := range dpCSP.interfaceIdx {
		comment := fmt.Sprintf("%d - Protocol: %s, DataFormat: %s", intfcIdx, intfc.Protocol, intfc.DataFormat)
		dpCSP.fzModel.AddHeaderComment(comment)
	}
	return &dpCSP
}

func (dpc *DataPathCSP) addModCapInterfacesToMap(cap moduleAndCapability) {
	for _, iface := range cap.capability.SupportedInterfaces {
		dpc.addInterfaceToMap(*iface.Source)
		dpc.addInterfaceToMap(*iface.Sink)
	}
}

func (dpc *DataPathCSP) addInterfaceToMap(intfc taxonomy.Interface) {
	_, found := dpc.interfaceIdx[intfc]
	if !found {
		dpc.interfaceIdx[intfc] = len(dpc.interfaceIdx) + 1
	}
}

func (dpc *DataPathCSP) moduleCapabilityAllowedByRestrictions(cap moduleAndCapability) bool {
	decision := dpc.problemData.Configuration.ConfigDecisions[cap.capability.Capability]
	if decision.Deploy == adminconfig.StatusFalse {
		return false // this type of capability should never be deployed
	}
	return dpc.moduleSatisfiesRestrictions(cap.module, decision.DeploymentRestrictions.Modules)
}

// NOTE: Minimal index of FlatZinc arrays is always 1. Hence, we use 1-based modeling all over the place  to avoid confusion

func (dpc *DataPathCSP) BuildFzModel(pathLength uint) error {
	// Variables to select the module capability we use on each data-path location
	moduleCapabilityVarType := rangeVarType(len(dpc.modulesCapabilities))
	dpc.fzModel.AddVariableArray("moduleCapability", moduleCapabilityVarType, pathLength, "", false, true)
	// Variables to select storage-accounts to place on each data-path location (the value 0 means no storage account)
	saTypeVarType := rangeVarType(len(dpc.env.StorageAccounts))
	dpc.fzModel.AddVariableArray("storageAccount", saTypeVarType, pathLength, "", false, true)
	// Variables to select the cluster we allocate to each module on the path
	moduleClusterVarType := rangeVarType(len(dpc.env.Clusters))
	dpc.fzModel.AddVariableArray("moduleCluster", moduleClusterVarType, pathLength, "", false, true)
	// Variables to select the source and sink interface for each module on the path
	moduleInterfaceVarType := rangeVarType(len(dpc.interfaceIdx))
	dpc.fzModel.AddVariableArray("moduleSourceInterface", moduleInterfaceVarType, pathLength, "", false, true)
	dpc.fzModel.AddVariableArray("moduleSinkInterface", moduleInterfaceVarType, pathLength, "", false, true)

	dpc.addGovernanceActionConstraints(pathLength)
	dpc.addAdminConfigRestrictions(int(pathLength))
	dpc.addInterfaceConstraints(pathLength)

	err := dpc.fzModel.Dump("dataPath.fzn")
	return err
}

// enforce restrictions from admin configuration decisions:
// a. selected capability in each location is not forbidden
// b. module satisifes restrictions for the selected capability
// c. cluster satisfies restrictions for the selected capability
// d. storage account satisfies restrictions for the selected capability
func (dpc *DataPathCSP) addAdminConfigRestrictions(pathLength int) {
	for decCapability, decision := range dpc.problemData.Configuration.ConfigDecisions {
		for modCapIdx, moduleCap := range dpc.modulesCapabilities {
			if moduleCap.capability.Capability != decCapability {
				continue
			}
			for clusterIdx, cluster := range dpc.env.Clusters {
				if !dpc.clusterSatisfiesRestrictions(cluster, decision.DeploymentRestrictions.Clusters) {
					dpc.preventAssignments([]string{"moduleCapability", "moduleCluster"},
						[]int{modCapIdx + 1, clusterIdx + 1}, pathLength)
				}
			}
			for saIdx, sa := range dpc.env.StorageAccounts {
				if !dpc.saSatisfiesRestrictions(sa, decision.DeploymentRestrictions.StorageAccounts) {
					dpc.preventAssignments([]string{"moduleCapability", "storageAccount"},
						[]int{modCapIdx + 1, saIdx + 1}, pathLength)
				}
			}
		}
	}
}

func (dpc *DataPathCSP) moduleSatisfiesRestrictions(module *appApi.FybrikModule, restrictions []adminconfig.Restriction) bool {
	for _, restriction := range restrictions {
		if !restriction.SatisfiedByResource(dpc.env.AttributeManager, module.Spec, "") {
			return false
		}
	}
	return true
}

func (dpc *DataPathCSP) clusterSatisfiesRestrictions(cluster multicluster.Cluster, restrictions []adminconfig.Restriction) bool {
	for _, restriction := range restrictions {
		if !restriction.SatisfiedByResource(dpc.env.AttributeManager, &cluster, "") {
			return false
		}
	}
	return true
}

func (dpc *DataPathCSP) saSatisfiesRestrictions(sa appApi.FybrikStorageAccount, restrictions []adminconfig.Restriction) bool {
	for _, restriction := range restrictions {
		if !restriction.SatisfiedByResource(dpc.env.AttributeManager, &sa.Spec, sa.Name) {
			return false
		}
	}
	return true
}

// Replicates a constraint to block a specific assignment for each position in the path
func (dpc *DataPathCSP) preventAssignments(variables []string, values []int, pathLength int) {
	for pos := 1; pos <= pathLength; pos++ {
		indexedVars := []string{}
		for _, v := range variables {
			indexedVars = append(indexedVars, varAtPos(v, pos))
		}
		dpc.preventAssignment(indexedVars, values)
	}
}

// Adds contraints to prevent the joint assignment of `values` to `variables`
func (dpc *DataPathCSP) preventAssignment(variables []string, values []int) {
	indicators := []string{}
	for idx, variable := range variables {
		indicators = append(indicators, dpc.addIndicator(variable, values[idx], "ne"))
	}
	indicatorsArray := fmt.Sprintf("[%s]", strings.Join(indicators, ", "))
	dpc.fzModel.AddConstraint("array_bool_or", []string{indicatorsArray, "true"})
}

func (dpc *DataPathCSP) addIndicator(variable string, value int, operator string) string {
	indicator := fmt.Sprintf("ind_%s_%s_%d", variable, operator, value)
	indicator = strings.ReplaceAll(indicator, "[", "_")
	indicator = strings.ReplaceAll(indicator, "]", "_")
	if _, defined := dpc.indicators[indicator]; !defined {
		dpc.fzModel.AddVariable(indicator, "bool", "", true, false)
		constraint := fmt.Sprintf("int_%s_reif", operator)
		annotation := fmt.Sprintf("defines_var(%s)", indicator)
		dpc.fzModel.AddConstraint(constraint, []string{variable, strconv.Itoa(value), indicator}, annotation)
	}
	return indicator
}

// Make sure that every required governance action is implemented exactly one time.
func (dpc *DataPathCSP) addGovernanceActionConstraints(pathLength uint) {
	if len(dpc.problemData.Actions) < 1 {
		return
	}

	repeatingOnes := strings.Repeat("1, ", int(pathLength))
	allOnesArrayLiteral := fmt.Sprintf("[%s]", repeatingOnes[0:len(repeatingOnes)-2])
	dpc.fzModel.AddParamArray("AllOnes", "int", pathLength, allOnesArrayLiteral)

	for _, action := range dpc.problemData.Actions {
		// Variables to mark specific actions are applied at location i
		actionVar := fmt.Sprintf("action%s", action.Name)
		dpc.fzModel.AddVariableArray(actionVar, "bool", pathLength, "", false, true)
		dpc.fzModel.AddConstraint("bool_lin_eq", []string{"AllOnes", actionVar, "1"}) // ensuring action is implemented once

		// accumulate module-capabilities that support the current action
		moduleCapabilitiesStrs := []string{}
		for modCapIdx, modCap := range dpc.modulesCapabilities {
			for _, capAction := range modCap.capability.Actions {
				if capAction.Name == action.Name {
					moduleCapabilitiesStrs = append(moduleCapabilitiesStrs, strconv.Itoa(modCapIdx+1))
				}
			}
		}
		modCapsSupportingAction := fmt.Sprintf("{%s}", strings.Join(moduleCapabilitiesStrs, ", "))

		// add vars (and constraints) indicating if an action is supported at each path location
		for pathPos := 1; pathPos <= int(pathLength); pathPos++ {
			modCapAtPos := varAtPos("moduleCapability", pathPos)
			actSupportedVarname := fmt.Sprintf("action%sSupportedAt%d", action.Name, pathPos)
			dpc.fzModel.AddVariable(actSupportedVarname, "bool", "", true, false)
			dpc.fzModel.AddConstraint("set_in_reif", []string{modCapAtPos, modCapsSupportingAction, actSupportedVarname})
			dpc.fzModel.AddConstraint("bool_le", []string{varAtPos(actionVar, pathPos), actSupportedVarname})
		}
	}
}

// prevent setting source/sink interfaces which are not supported by module capability
func (dpc *DataPathCSP) addInterfaceConstraints(pathLength uint) {
	for intfc, intfcIdx := range dpc.interfaceIdx {
		for modCapIdx, modCap := range dpc.modulesCapabilities {
			modcapSupportsIntfcSrc := false
			modcapSupportsIntfcSink := false
			for _, modifc := range modCap.capability.SupportedInterfaces {
				if interfacesMatch(*modifc.Source, intfc) {
					modcapSupportsIntfcSrc = true
				}
				if interfacesMatch(*modifc.Sink, intfc) {
					modcapSupportsIntfcSink = true
				}
			}
			if !modcapSupportsIntfcSrc {
				dpc.preventAssignments([]string{"moduleCapability", "moduleSourceInterface"}, []int{modCapIdx + 1, intfcIdx}, int(pathLength))
			}
			if !modcapSupportsIntfcSink {
				dpc.preventAssignments([]string{"moduleCapability", "moduleSinkInterface"}, []int{modCapIdx + 1, intfcIdx}, int(pathLength))
			}
		}
	}

	// ensuring interface matching along the datapath from dataset to workload
	dpc.fzModel.AddConstraint("int_eq", []string{varAtPos("moduleSourceInterface", 1), "1"})
	for pathPos := 1; pathPos < int(pathLength); pathPos++ {
		dpc.fzModel.AddConstraint("int_eq", []string{varAtPos("moduleSinkInterface", pathPos), varAtPos("moduleSourceInterface", pathPos+1)})
	}
	dpc.fzModel.AddConstraint("int_eq", []string{varAtPos("moduleSinkInterface", int(pathLength)),
		strconv.Itoa(dpc.interfaceIdx[dpc.problemData.Context.Requirements.Interface])})
}

// ----- helper functions -----

func rangeVarType(rangeEnd int) string {
	if rangeEnd < 1 {
		rangeEnd = 1
	}
	return "1.." + strconv.Itoa(rangeEnd)
}

func varAtPos(variable string, pos int) string {
	return fmt.Sprintf("%s[%d]", variable, pos)
}

func interfacesMatch(intfc1, intfc2 taxonomy.Interface) bool {
	return intfc1.Protocol == intfc2.Protocol && intfc1.DataFormat == intfc2.DataFormat
}
