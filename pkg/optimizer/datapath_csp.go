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
)

type DataPathCSP struct {
	problemData *app.DataInfo
	env         *app.Environment
	modules     []*appApi.FybrikModule // to ensure consistent iteration order (env.Modules is a map)
	indicators  map[string]bool        // indicators used as part of the problem (to prevent redefinition)
	fzModel     *FlatZincModel
}

func NewDataPathCSP(problemData *app.DataInfo, env *app.Environment) *DataPathCSP {
	dpCSP := DataPathCSP{problemData: problemData, env: env, fzModel: NewFlatZincModel()}
	dpCSP.modules = []*appApi.FybrikModule{}
	for _, module := range env.Modules {
		dpCSP.modules = append(dpCSP.modules, module)
	}
	return &dpCSP
}

// NOTE: Minimal index of FlatZinc arrays is always 1. Hence, we use 1-based modeling allover to avoid confusion

func (dpc *DataPathCSP) BuildFzModel(pathLength uint) error {
	outputArrayAnnotation := fmt.Sprintf("output_array([1..%d])", pathLength)
	// Variables to select the module we place on each data-path location
	moduleTypeVarType := rangeVarType(len(dpc.env.Modules))
	dpc.fzModel.AddVariable(pathLength, "moduleType", moduleTypeVarType, "", outputArrayAnnotation)
	// Variables to select the module capability we use on each data-path location
	maxCapabilityIndex, maxCapabilityPerModule := getMaxCapabilityIdx(dpc.modules)
	dpc.fzModel.AddParam(uint(len(dpc.env.Modules)), "MaxModuleCapability", "int", "["+maxCapabilityPerModule+"]")
	moduleCapabilityVarType := rangeVarType(maxCapabilityIndex)
	dpc.fzModel.AddVariable(pathLength, "moduleCapability", moduleCapabilityVarType, "", outputArrayAnnotation)
	// Variables to select storage-accounts to place on each data-path location (the value 0 means no storage account)
	saTypeVarType := rangeVarType(len(dpc.env.StorageAccounts))
	dpc.fzModel.AddVariable(pathLength, "storageAccount", saTypeVarType, "", outputArrayAnnotation)
	// Variables to select the cluster we allocate to each module on the path
	moduleClusterVarType := rangeVarType(len(dpc.env.Clusters))
	dpc.fzModel.AddVariable(pathLength, "moduleCluster", moduleClusterVarType, "", outputArrayAnnotation)

	dpc.addGovernanceActionConstraints(pathLength, maxCapabilityIndex)
	dpc.addModuleCapabilitiesConstraints(int(pathLength))
	dpc.addForbiddenCapabilitiesConstraints(int(pathLength))

	err := dpc.fzModel.Dump("dataPath.fzn")
	return err
}

// Ensure that a module at position i is not assigned a capability it doesn't have
func (dpc *DataPathCSP) addModuleCapabilitiesConstraints(pathLength int) {
	for pathPos := 1; pathPos <= pathLength; pathPos++ {
		moduleAtPosVar := fmt.Sprintf("moduleType[%d]", pathPos)
		CapabilityAtPosVar := fmt.Sprintf("moduleCapability[%d]", pathPos)
		MaxCapabilityForModuleAtPosVar := fmt.Sprintf("maxCapabilityForModuleAt%d", pathPos)
		dpc.fzModel.AddVariable(1, MaxCapabilityForModuleAtPosVar, "int", "")
		dpc.fzModel.AddConstraint("array_var_int_element", []string{moduleAtPosVar, "MaxModuleCapability", MaxCapabilityForModuleAtPosVar})
		dpc.fzModel.AddConstraint("int_le", []string{CapabilityAtPosVar, MaxCapabilityForModuleAtPosVar})
	}
}

//ensure selected capability is not forbidden
func (dpc *DataPathCSP) addForbiddenCapabilitiesConstraints(pathLength int) {
	for decCapability, decision := range dpc.problemData.Configuration.ConfigDecisions {
		if decision.Deploy == adminconfig.StatusFalse {
			for modIdx, module := range dpc.modules {
				for capIdx, modCapability := range module.Spec.Capabilities {
					if modCapability.Capability == decCapability {
						for pos := 1; pos <= pathLength; pos++ {
							modulesVar := fmt.Sprintf("moduleType[%d]", pos)
							capabilityVar := fmt.Sprintf("moduleCapability[%d]", pos)
							dpc.preventAssignment([]string{modulesVar, capabilityVar}, []int{modIdx + 1, capIdx + 1})
						}
					}
				}
			}
		}
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
		dpc.fzModel.AddVariable(1, indicator, "bool", "", "is_defined_var")
		constraint := fmt.Sprintf("int_%s_reif", operator)
		annotation := fmt.Sprintf("defines_var(%s)", indicator)
		dpc.fzModel.AddConstraint(constraint, []string{variable, strconv.Itoa(value), indicator}, annotation)
	}
	return indicator
}

// Make sure that every required governance action is implemented exactly one time.
// TODO: check if multiple copies of same action allowed
// TODO: check if a module can only use some of its supported actions
func (dpc *DataPathCSP) addGovernanceActionConstraints(pathLength uint, maxCapabilityIndex int) {
	if len(dpc.problemData.Actions) < 1 {
		return
	}

	repeatingOnes := strings.Repeat("1, ", int(pathLength))
	allOnesArrayLiteral := fmt.Sprintf("[%s]", repeatingOnes[0:len(repeatingOnes)-2])
	dpc.fzModel.AddParam(pathLength, "AllOnes", "int", allOnesArrayLiteral)

	// Variables to mark specific actions are applied at location i
	outputArrayAnnotation := fmt.Sprintf("output_array([1..%d])", pathLength)
	for _, action := range dpc.problemData.Actions {
		actionVar := fmt.Sprintf("action%s", action.Name)
		dpc.fzModel.AddVariable(pathLength, actionVar, "bool", "", outputArrayAnnotation)
		dpc.fzModel.AddConstraint("bool_lin_eq", []string{"AllOnes", actionVar, "1"})

		// for each modules store which capabilities support the current action
		moduleCapabilitiesStrs := []string{}
		for _, module := range dpc.modules {
			for _, capability := range module.Spec.Capabilities {
				actionFound := false
				for _, capAction := range capability.Actions {
					if capAction.Name == action.Name {
						actionFound = true
					}
				}
				moduleCapabilitiesStrs = append(moduleCapabilitiesStrs, strconv.FormatBool(actionFound))
			}
			for capIdx := len(module.Spec.Capabilities); capIdx < maxCapabilityIndex; capIdx++ {
				moduleCapabilitiesStrs = append(moduleCapabilitiesStrs, strconv.FormatBool(false)) // padding
			}
		}
		moduleCapabilitiesStr := fmt.Sprintf("[%s]", strings.Join(moduleCapabilitiesStrs, ", "))
		actModCapVarname := fmt.Sprintf("ModuleCapabilitiesSupporting%s", action.Name)
		dpc.fzModel.AddParam(uint(len(dpc.modules)*maxCapabilityIndex), actModCapVarname, "bool", moduleCapabilitiesStr)

		for pathPos := 1; pathPos <= int(pathLength); pathPos++ {
			actSupportedVarname := fmt.Sprintf("action%sSupportedAt%d", action.Name, pathPos)
			actSupportedIdxVarname := actSupportedVarname + "Idx"
			dpc.fzModel.AddVariable(1, actSupportedIdxVarname, "int", "", "output_var")
			dpc.fzModel.AddVariable(1, actSupportedVarname, "bool", "", "output_var")
			idxVectorParams := fmt.Sprintf("[%s, moduleType[%d], moduleCapability[%d] , 1]", actSupportedIdxVarname, pathPos, pathPos)
			dpc.fzModel.AddConstraint("int_lin_eq", []string{"[-1, 2, 1, -2]", idxVectorParams, "0"})
			dpc.fzModel.AddConstraint("array_bool_element", []string{actSupportedIdxVarname, actModCapVarname, actSupportedVarname})
			dpc.fzModel.AddConstraint("bool_le", []string{fmt.Sprintf("%s[%d]", actionVar, pathPos), actSupportedVarname})
		}
	}
}

// ----- helper functions -----

func rangeVarType(rangeEnd int) string {
	if rangeEnd < 1 {
		rangeEnd = 1
	}
	return "1.." + strconv.Itoa(rangeEnd)
}

// returns the maximal capability index over all modules (as int), and per module (as a string)
func getMaxCapabilityIdx(modules []*appApi.FybrikModule) (int, string) {
	maxCapabilityIndex := 0
	moduleMaxCapabilityIdx := []string{}
	for _, module := range modules {
		numModuleCapabilities := len(module.Spec.Capabilities)
		moduleMaxCapabilityIdx = append(moduleMaxCapabilityIdx, strconv.Itoa(numModuleCapabilities))
		if numModuleCapabilities > maxCapabilityIndex {
			maxCapabilityIndex = numModuleCapabilities
		}
	}
	return maxCapabilityIndex, strings.Join(moduleMaxCapabilityIdx, ", ")
}
