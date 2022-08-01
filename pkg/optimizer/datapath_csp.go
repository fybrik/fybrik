// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	appApi "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/datapath"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
)

// Names of the primary variables on which we need to make decisions
// Each variable is an array of ints. The i-th cell in the array represents a decision for the i-th Node/Edge in the data path
const (
	modCapVarname    = "moduleCapability"      // Var's value says which capability of which module to use (on each Edge)
	clusterVarname   = "moduleCluster"         // Var's value says which of the available clusters to use (on each Edge)
	saVarname        = "storageAccount"        // Var's value says which storage account to use
	srcIntfcVarname  = "moduleSourceInterface" // Var's value says which interface to use as source
	sinkIntfcVarname = "moduleSinkInterface"   // Var's value says which interface to use as sink
	actionVarname    = "action_%s"             // Vars for each required action, say whether the action was applied
	jointGoalVarname = "jointGoal"             // Var's value indicates the quality of the data path w.r.t. optimization goals

	// The following variables are only allocated and used when inter-region goals are set
	storageLocsVarname               = "storageLocations"          // Var is a concatenation of the value "1" and the values in saVarname
	realSaLocationsVarName           = "realSaLocations"           // Var at pos i is i if storageLocsVarname[i]>0, and 0 otherwise
	maxRealSaVarName                 = "maxRealSA"                 // The maximal value in the realSaLocationsVarName vector
	afterMaxRealSaVarName            = "afterMaxRealSA"            // Var at pos i is true if i >= maxRealSaVarName, and false otherwise
	c2cSelectorVarname               = "c2cSelector"               // Cluster-to-cluster entry to select from a goal's paramArray (at each pos)
	s2cSelectorVarname               = "s2cSelector"               // The storage-to-cluster entry to select from the goals paramArrays
	lastDataStoreVarname             = "lastDataStore"             // The last data store used in the data path
	clusterAfterLastDataStoreVarname = "clusterAfterLastDataStore" // The cluster after the last data store used in the data path

	minusOneStr = "-1"
	zeroStr     = "0"
	oneStr      = "1"
)

// Couples together a module and one of its capabilities
type moduleAndCapability struct {
	module        *appApi.FybrikModule
	capability    *appApi.ModuleCapability
	capabilityIdx int  // The index of capability in module's spec
	virtualSource bool // whether data is consumed via API
	virtualSink   bool // whether data is transferred in memory via API
	hasSource     bool // whether module-capability has source interfaces
	hasSink       bool // whether module-capability has sink interfaces
}

// The main class for producing a CSP from data-path constraints and for decoding solver's solutions
type DataPathCSP struct {
	problemData         *datapath.DataInfo
	env                 *datapath.Environment
	modulesCapabilities []moduleAndCapability       // An enumeration of allowed capabilities in all modules
	interfaceIdx        map[taxonomy.Interface]int  // gives an index for each unique interface
	reverseIntfcMap     map[int]*taxonomy.Interface // The reverse mapping (needed when decoding the solution)
	requiredActions     map[string]taxonomy.Action  // A map from action variables to the actions they represent
	fzModel             *FlatZincModel
	noStorageAccountVal int
}

// The ctor also enumerates all available (module x capabilities) and all available interfaces
// The generated enumerations are listed at the header of the FlatZinc model
func NewDataPathCSP(problemData *datapath.DataInfo, env *datapath.Environment) *DataPathCSP {
	dpCSP := DataPathCSP{problemData: problemData, env: env, fzModel: NewFlatZincModel()}
	dpCSP.requiredActions = map[string]taxonomy.Action{}
	dpCSP.interfaceIdx = map[taxonomy.Interface]int{}
	dpCSP.reverseIntfcMap = map[int]*taxonomy.Interface{}
	dataSetIntfc := getAssetInterface(dpCSP.problemData.DataDetails)
	dpCSP.addInterface(nil)           // ensure nil interface always gets index 0
	dpCSP.addInterface(&dataSetIntfc) // data-set interface always gets index 1 (cannot be nil)
	dpCSP.addInterface(dpCSP.problemData.Context.Requirements.Interface)

	dpCSP.fzModel.AddHeaderComment("Encoding of modules and their capabilities:")
	comment := ""
	for _, module := range env.Modules {
		for idx, capability := range module.Spec.Capabilities {
			modCap := moduleAndCapability{module, &module.Spec.Capabilities[idx], idx, false, false, false, false}
			if dpCSP.moduleCapabilityAllowedByRestrictions(modCap) {
				dpCSP.addModCapInterfacesToMaps(&modCap)
				dpCSP.modulesCapabilities = append(dpCSP.modulesCapabilities, modCap)
				comment = strconv.Itoa(len(dpCSP.modulesCapabilities))
			} else {
				comment = "<forbidden>"
			}
			comment += fmt.Sprintf(" - Module: %s, Capability: %d (%s)", module.Name, idx, capability.Capability)
			dpCSP.fzModel.AddHeaderComment(comment)
		}
	}

	dpCSP.fzModel.AddHeaderComment("Encoding of interfaces:")
	for intfc, intfcIdx := range dpCSP.interfaceIdx {
		dpCSP.fzModel.AddHeaderComment(encodingComment(intfcIdx, fmt.Sprintf("%v", intfc)))
	}
	dpCSP.fzModel.AddHeaderComment("Encoding of clusters:")
	for clusterIdx, cluster := range dpCSP.env.Clusters {
		dpCSP.fzModel.AddHeaderComment(encodingComment(clusterIdx+1, cluster.Name))
	}
	dpCSP.fzModel.AddHeaderComment("Encoding of storage accounts:")
	for saIdx, sa := range dpCSP.env.StorageAccounts {
		dpCSP.fzModel.AddHeaderComment(encodingComment(saIdx+1, sa.Name))
	}
	dpCSP.noStorageAccountVal = len(dpCSP.env.StorageAccounts) + 1
	dpCSP.fzModel.AddHeaderComment(encodingComment(dpCSP.noStorageAccountVal, "No storage account"))
	return &dpCSP
}

// Add the interfaces defined in a given module's capability to the 2 interface maps
func (dpc *DataPathCSP) addModCapInterfacesToMaps(modcap *moduleAndCapability) {
	capability := modcap.capability
	for _, intfc := range capability.SupportedInterfaces {
		if intfc.Source != nil {
			dpc.addInterface(intfc.Source)
			modcap.hasSource = true
		}
		if intfc.Sink != nil {
			dpc.addInterface(intfc.Sink)
			modcap.hasSink = true
		}
	}
	if (!modcap.hasSource || !modcap.hasSink) && capability.API != nil {
		apiInterface := &taxonomy.Interface{Protocol: capability.API.Connection.Name, DataFormat: capability.API.DataFormat}
		dpc.addInterface(apiInterface)
		modcap.virtualSource = !modcap.hasSource
		modcap.virtualSink = !modcap.hasSink
		modcap.hasSource = true
		modcap.hasSink = true
	}
}

// Add the given interface to the 2 interface maps (but avoid duplicates)
func (dpc *DataPathCSP) addInterface(intfc *taxonomy.Interface) {
	if intfc == nil {
		intfc = &taxonomy.Interface{}
	}
	_, found := dpc.interfaceIdx[*intfc]
	if !found {
		intfcIdx := len(dpc.interfaceIdx)
		dpc.interfaceIdx[*intfc] = intfcIdx
		dpc.reverseIntfcMap[intfcIdx] = intfc
	}
}

// This is the main method for building a FlatZinc CSP out of the data-path parameters and constraints.
// Returns a file name where the model was dumped
// NOTE: Minimal index of FlatZinc arrays is always 1. Hence, we use 1-based modeling all over the place to avoid confusion
//       The only exception is with interfaces (0 means nil)
func (dpc *DataPathCSP) BuildFzModel(pathLength int) (string, error) {
	dpc.fzModel.Clear() // This function can be called multiple times - clear vars and constraints from last call
	// Variables to select the module capability we use on each data-path location
	moduleCapabilityVarType := fznRangeVarType(1, len(dpc.modulesCapabilities))
	dpc.fzModel.AddVariableArray(modCapVarname, moduleCapabilityVarType, pathLength, false, true)
	// Variables to select storage-accounts to place on each data-path location (last value means no storage account)
	saTypeVarType := fznRangeVarType(1, dpc.noStorageAccountVal)
	dpc.fzModel.AddVariableArray(saVarname, saTypeVarType, pathLength, false, true)
	// Variables to select the cluster we allocate to each module on the path
	moduleClusterVarType := fznRangeVarType(1, len(dpc.env.Clusters))
	dpc.fzModel.AddVariableArray(clusterVarname, moduleClusterVarType, pathLength+1, false, true)
	// Fix moduleCluster[pathLength+1] to the workload cluster
	workloadCluster := getWorkloadClusterIndex(dpc.problemData.WorkloadCluster, dpc.env.Clusters)
	dpc.fzModel.AddConstraint(IntEqConstraint, []string{varAtPos(clusterVarname, pathLength+1), workloadCluster, TrueValue})
	// Variables to select the source and sink interface for each module on the path (0 means nil interface)
	moduleInterfaceVarType := fznRangeVarType(0, len(dpc.interfaceIdx)-1)
	dpc.fzModel.AddVariableArray(srcIntfcVarname, moduleInterfaceVarType, pathLength, false, true)
	dpc.fzModel.AddVariableArray(sinkIntfcVarname, moduleInterfaceVarType, pathLength, false, true)

	dpc.addInterfaceConstraints(pathLength)
	dpc.addGovernanceActionConstraints(pathLength)
	err := dpc.addAdminConfigRestrictions(pathLength)
	if err != nil {
		return "", err
	}
	err = dpc.addOptimizationGoals(pathLength)
	if err != nil {
		return "", err
	}

	return dpc.fzModel.Dump()
}

// enforce restrictions from admin configuration decisions.
// Note: module+capability that does not satisfy capability restrictions is already filtered from dpc.modulesCapabilities
// a. cluster satisfies restrictions for the selected capability
// b. storage account satisfies restrictions for the selected capability
// c. module satisfies "transform" restrictions if a governance action is selected
// d. cluster satisfies "transform" restrictions if a governance action is selected
// e. storage account satisfies "transform" restrictions if a governance action is selected
// f. capabilities that must be deployed are indeed deployed
func (dpc *DataPathCSP) addAdminConfigRestrictions(pathLength int) error {
	for decCapability := range dpc.problemData.Configuration.ConfigDecisions {
		decision := dpc.problemData.Configuration.ConfigDecisions[decCapability]
		relevantModCaps := []string{}
		for modCapIdx, moduleCap := range dpc.modulesCapabilities {
			if moduleCap.capability.Capability != decCapability {
				continue
			}
			relevantModCaps = append(relevantModCaps, strconv.Itoa(modCapIdx+1))
			dpc.enforceClusterRestrictions(decision.DeploymentRestrictions, modCapVarname, modCapIdx+1, pathLength)
			dpc.enforceStorageRestrictions(decision.DeploymentRestrictions, modCapVarname, modCapIdx+1, pathLength)
		}
		if decCapability == "transform" {
			for actionVar := range dpc.requiredActions {
				dpc.enforceModuleRestrictions(decision.DeploymentRestrictions, actionVar, 1, pathLength)
				dpc.enforceClusterRestrictions(decision.DeploymentRestrictions, actionVar, 1, pathLength)
				dpc.enforceStorageRestrictions(decision.DeploymentRestrictions, actionVar, 1, pathLength)
			}
		}
		if decision.Deploy == adminconfig.StatusTrue { // this capability must be deployed
			if len(relevantModCaps) == 0 {
				return fmt.Errorf("capability %v is required, but it is not supported by any module", decCapability)
			}
			dpc.ensureCapabilityIsDeployed(relevantModCaps, pathLength)
		}
	}
	return nil
}

// checks given restrictions on each moduleCapability, and if the restriction is violated for the given module,
// blocks the assignment varToBlock==valueToBlock && modCapability==moduleCapabilityIndex
func (dpc *DataPathCSP) enforceModuleRestrictions(restrictions adminconfig.Restrictions,
	varToBlock string, valueToBlock, pathLen int) {
	for modCapIdx := range dpc.modulesCapabilities {
		if !dpc.modcapSatisfiesRestrictions(&dpc.modulesCapabilities[modCapIdx], restrictions.Modules) {
			dpc.preventAssignments([]string{varToBlock, modCapVarname}, []int{valueToBlock, modCapIdx + 1}, pathLen)
		}
	}
}

// checks given restrictions on each cluster, and if the restriction is violated for a given cluster,
// blocks the assignment varToBlock==valueToBlock && cluster==clusterIndex
func (dpc *DataPathCSP) enforceClusterRestrictions(restrictions adminconfig.Restrictions,
	varToBlock string, valueToBlock, pathLen int) {
	for clusterIdx, cluster := range dpc.env.Clusters {
		if !dpc.clusterSatisfiesRestrictions(cluster, restrictions.Clusters) {
			dpc.preventAssignments([]string{varToBlock, clusterVarname}, []int{valueToBlock, clusterIdx + 1}, pathLen)
		}
	}
}

// checks given restrictions on each storage account, and if the restriction is violated for a given SA,
// blocks the assignment varToBlock==valueToBlock && saVarname==saIndex
func (dpc *DataPathCSP) enforceStorageRestrictions(restrictions adminconfig.Restrictions,
	varToBlock string, valueToBlock, pathLen int) {
	for saIdx, sa := range dpc.env.StorageAccounts {
		if !dpc.saSatisfiesRestrictions(sa, restrictions.StorageAccounts) {
			dpc.preventAssignments([]string{varToBlock, saVarname}, []int{valueToBlock, saIdx + 1}, pathLen)
		}
	}
}

// Adds a constraint to ensure that at least one moduleCapability is chosen from the given set "modCaps"
func (dpc *DataPathCSP) ensureCapabilityIsDeployed(modCaps []string, pathLength int) {
	reqCapIndicator := dpc.addSetInIndicator(modCapVarname, modCaps, pathLength)
	dpc.fzModel.AddConstraint( // the weighted sum of the indicators (with all weights set to -1) should be <= -1
		BoolLinLeConstraint,
		[]string{fznCompoundLiteral(arrayOfSameInt(-1, pathLength), false), reqCapIndicator, strconv.Itoa(-1)},
	)
}

// Decide if a given module and its given capability satisfy all administrator's restrictions
func (dpc *DataPathCSP) moduleCapabilityAllowedByRestrictions(modcap moduleAndCapability) bool {
	decision := dpc.problemData.Configuration.ConfigDecisions[modcap.capability.Capability]
	if decision.Deploy == adminconfig.StatusFalse {
		return false // this type of capability should never be deployed
	}
	return dpc.modcapSatisfiesRestrictions(&modcap, decision.DeploymentRestrictions.Modules)
}

// Decide if a given module satisfies all administrator's restrictions
func (dpc *DataPathCSP) modcapSatisfiesRestrictions(modcap *moduleAndCapability, restrictions []adminconfig.Restriction) bool {
	oldPrefix := "capabilities."
	newPrefix := oldPrefix + strconv.Itoa(modcap.capabilityIdx) + "."
	for _, restriction := range restrictions {
		restriction.Property = strings.Replace(restriction.Property, oldPrefix, newPrefix, 1)
		if !restriction.SatisfiedByResource(dpc.env.AttributeManager, modcap.module.Spec, "") {
			return false
		}
	}
	return true
}

// Decide if a given cluster satisfies all administrator's restrictions
func (dpc *DataPathCSP) clusterSatisfiesRestrictions(cluster multicluster.Cluster, restrictions []adminconfig.Restriction) bool {
	for _, restriction := range restrictions {
		if !restriction.SatisfiedByResource(dpc.env.AttributeManager, &cluster, "") {
			return false
		}
	}
	return true
}

// Decide if a given storage account satisfies all administrator's restrictions
func (dpc *DataPathCSP) saSatisfiesRestrictions(sa *appApi.FybrikStorageAccount, restrictions []adminconfig.Restriction) bool {
	for _, restriction := range restrictions {
		if !restriction.SatisfiedByResource(dpc.env.AttributeManager, &sa.Spec, sa.Name) {
			return false
		}
	}
	return true
}

// Replicates a constraint to block a specific combination of assignments for each position in the path
func (dpc *DataPathCSP) preventAssignments(variables []string, values []int, pathLength int) {
	// Prepare an indicator for each variable which is true iff the variable is NOT assigned its given value
	indicators := []string{}
	for idx, variable := range variables {
		if dpc.fzModel.GetVariableType(variable) == BoolType {
			if values[idx] == 0 { // if the variable is Boolean, we assume that "false" is 0 and "true" is anything else
				indicators = append(indicators, variable) // the var is an indicator for itself not being "false"
			} else {
				indicators = append(indicators, dpc.addBoolNotIndicator(variable, pathLength))
			}
		} else {
			indicators = append(indicators, dpc.addEqualityIndicator(variable, values[idx], pathLength, false))
		}
	}

	for pos := 1; pos <= pathLength; pos++ {
		indexedIndicators := []string{}
		for _, v := range indicators {
			indexedIndicators = append(indexedIndicators, varAtPos(v, pos))
		}
		indicatorsArray := fznCompoundLiteral(indexedIndicators, false)
		dpc.fzModel.AddConstraint(ArrBoolOrConstraint, []string{indicatorsArray, TrueValue})
	}
}

// Given a Boolean variable, returns indicator variable array which is true iff the variable is false at each pos
func (dpc *DataPathCSP) addBoolNotIndicator(variable string, pathLength int) string {
	indicator := fmt.Sprintf("ind_not_%s", variable)
	if _, defined := dpc.fzModel.VarMap[indicator]; defined {
		return indicator
	}
	dpc.fzModel.AddVariableArray(indicator, BoolType, pathLength, true, false)
	for pathPos := 1; pathPos <= pathLength; pathPos++ {
		indicatorAtPos := varAtPos(indicator, pathPos)
		annotation := GetDefinesVarAnnotation(indicatorAtPos)
		dpc.fzModel.AddConstraint(BoolNotEqConstraint, []string{varAtPos(variable, pathPos), indicatorAtPos}, annotation)
	}

	return indicator
}

// Adds an indicator array whose elements are true iff a given integer variable EQUALS a given value in a given pos
// Setting equality to false will check if the integer DOES NOT EQUAL the given value
func (dpc *DataPathCSP) addEqualityIndicator(variable string, value, pathLength int, equality bool) string {
	constraint := IntEqConstraint
	if !equality {
		constraint = IntNotEqConstraint
	}
	indicator := fmt.Sprintf("ind_%s_%s_%d", variable, constraint, value)
	if _, defined := dpc.fzModel.VarMap[indicator]; defined {
		return indicator
	}

	dpc.fzModel.AddVariableArray(indicator, BoolType, pathLength, true, false)
	strVal := strconv.Itoa(value)
	for pathPos := 1; pathPos <= pathLength; pathPos++ {
		variableAtPos := varAtPos(variable, pathPos)
		indicatorAtPos := varAtPos(indicator, pathPos)
		annotation := GetDefinesVarAnnotation(indicatorAtPos)
		dpc.fzModel.AddConstraint(constraint, []string{variableAtPos, strVal, indicatorAtPos}, annotation)
	}
	return indicator
}

// Adds an indicator per path location to check if the value of "variable" in this location is in the given set of values
func (dpc *DataPathCSP) addSetInIndicator(variable string, valueSet []string, pathLength int) string {
	indicator := fmt.Sprintf("ind_%s_in_%s", variable, strings.Join(valueSet, "_"))
	if _, defined := dpc.fzModel.VarMap[indicator]; defined {
		return indicator
	}

	dpc.fzModel.AddVariableArray(indicator, BoolType, pathLength, true, false)
	if len(valueSet) > 0 {
		for pathPos := 1; pathPos <= pathLength; pathPos++ {
			variableAtPos := varAtPos(variable, pathPos)
			indicatorAtPos := varAtPos(indicator, pathPos)
			setLiteral := fznCompoundLiteral(valueSet, true)
			annotation := GetDefinesVarAnnotation(indicatorAtPos)
			dpc.fzModel.AddConstraint(SetInConstraint, []string{variableAtPos, setLiteral, indicatorAtPos}, annotation)
		}
	} else { // value set is empty - indicators should always be false as variable value is never in the given set
		dpc.fzModel.AddConstraint(ArrBoolOrConstraint, []string{indicator, FalseValue})
	}
	return indicator
}

// Adds a Boolean indicator variable that is implied by either the given var or its negation
func (dpc *DataPathCSP) addImpliedByIndicator(variable string, pathLen int, impliedByNegatedVar bool) string {
	notStr := ""
	if impliedByNegatedVar {
		notStr = "not_"
	}
	indicator := fmt.Sprintf("ind_implied_by_%s%s", notStr, variable)
	if _, defined := dpc.fzModel.VarMap[indicator]; defined {
		return indicator
	}
	dpc.fzModel.AddVariableArray(indicator, BoolType, pathLen, true, false)
	for pathPos := 1; pathPos <= pathLen; pathPos++ {
		variableAtPos := varAtPos(variable, pathPos)
		indicatorAtPos := varAtPos(indicator, pathPos)
		annotations := GetDefinesVarAnnotation(indicatorAtPos)
		if impliedByNegatedVar {
			arrayToOr := fznCompoundLiteral([]string{variableAtPos, indicatorAtPos}, false)
			dpc.fzModel.AddConstraint(ArrBoolOrConstraint, []string{arrayToOr, TrueValue}, annotations)
		} else {
			dpc.fzModel.AddConstraint(BoolLeConstraint, []string{variableAtPos, indicatorAtPos}, annotations)
		}
	}

	return indicator
}

// Make sure that every required governance action is implemented exactly one time.
func (dpc *DataPathCSP) addGovernanceActionConstraints(pathLength int) {
	allOnesArrayLiteral := arrayOfSameInt(1, pathLength)
	for _, action := range dpc.problemData.Actions {
		// An *output* array of Booleans variable to mark whether the current action is applied at location i
		actionVar := dpc.addActionIndicator(action, pathLength)
		// ensuring action is implemented once
		dpc.fzModel.AddConstraint(
			BoolLinEqConstraint, []string{fznCompoundLiteral(allOnesArrayLiteral, false), actionVar, strconv.Itoa(1)})
	}

	if dpc.problemData.Context.Flow != taxonomy.WriteFlow || dpc.problemData.Context.Requirements.FlowParams.IsNewDataSet {
		for saIdx, sa := range dpc.env.StorageAccounts {
			actions, found := dpc.problemData.StorageRequirements[sa.Spec.Region]
			if !found { //
				dpc.preventAssignments([]string{saVarname}, []int{saIdx + 1}, pathLength)
			} else {
				for _, action := range actions {
					actionIndicator := dpc.addActionIndicator(action, pathLength)
					// ensuring action is implemented no more than once
					dpc.fzModel.AddConstraint(BoolLinLeConstraint,
						[]string{fznCompoundLiteral(allOnesArrayLiteral, false), actionIndicator, strconv.Itoa(1)})

					// Ensuring that if sa is chosen, then the required action is also applied
					saChosenIndicator := dpc.addEqualityIndicator(saVarname, saIdx+1, pathLength, true)
					saChosenVar := dpc.orOfIndicators(saChosenIndicator)
					actionChosenVar := dpc.orOfIndicators(actionIndicator)
					dpc.fzModel.AddConstraint(BoolLeConstraint, []string{saChosenVar, actionChosenVar})
				}
			}
		}
	}
}

// Returns a variable which holds the OR of all boolean variables in indicatorArray
func (dpc *DataPathCSP) orOfIndicators(indicatorArray string) string {
	bigOrVarname := indicatorArray + "_OR"
	dpc.fzModel.AddVariable(bigOrVarname, BoolType, true, false)
	annotation := GetDefinesVarAnnotation(bigOrVarname)
	dpc.fzModel.AddConstraint(ArrBoolOrConstraint, []string{indicatorArray, bigOrVarname}, annotation)
	return bigOrVarname
}

// Returns an *output* array of Booleans variable to mark whether the current action is applied at location i
func (dpc *DataPathCSP) addActionIndicator(action taxonomy.Action, pathLength int) string {
	actionVar := getActionVarname(action)
	dpc.requiredActions[actionVar] = action
	if _, found := dpc.fzModel.VarMap[actionVar]; found {
		return actionVar
	}
	dpc.fzModel.AddVariableArray(actionVar, BoolType, pathLength, false, true)

	// accumulate module-capabilities that support the current action
	moduleCapabilitiesStrs := []string{}
	for modCapIdx, modCap := range dpc.modulesCapabilities {
		for _, capAction := range modCap.capability.Actions {
			if capAction.Name == action.Name {
				moduleCapabilitiesStrs = append(moduleCapabilitiesStrs, strconv.Itoa(modCapIdx+1))
			}
		}
	}

	// add vars (and constraints) indicating if an action is supported at each path location
	setInIndicator := dpc.addSetInIndicator(modCapVarname, moduleCapabilitiesStrs, pathLength)
	for pathPos := 1; pathPos <= pathLength; pathPos++ {
		indicatorAtPos := varAtPos(setInIndicator, pathPos)
		actionVarAtPos := varAtPos(actionVar, pathPos)
		dpc.fzModel.AddConstraint(BoolLeConstraint, []string{actionVarAtPos, indicatorAtPos})
	}
	return actionVar
}

// prevent setting source/sink interfaces which are not supported by module capability
func (dpc *DataPathCSP) addInterfaceConstraints(pathLength int) {
	// First, make sure interface selection matches module-capability selection
	dpc.modCapSupportsIntfc(pathLength)

	// Now, ensure interfaces match along the data-path from dataset to workload
	startIntfcIndexes := fznCompoundLiteral(dpc.getMatchingInterfaces(dpc.reverseIntfcMap[1]), true)
	endIntfcIndexes := fznCompoundLiteral(dpc.getMatchingInterfaces(dpc.problemData.Context.Requirements.Interface), true)
	if dpc.problemData.Context.Flow == taxonomy.WriteFlow {
		startIntfcIndexes, endIntfcIndexes = endIntfcIndexes, startIntfcIndexes // swap start and end for write flows
	}
	dpc.fzModel.AddConstraint(SetInConstraint, []string{varAtPos(srcIntfcVarname, 1), startIntfcIndexes, TrueValue})
	for pathPos := 1; pathPos < pathLength; pathPos++ {
		dpc.fzModel.AddConstraint(IntEqConstraint,
			[]string{varAtPos(sinkIntfcVarname, pathPos), varAtPos(srcIntfcVarname, pathPos+1), TrueValue})
	}
	dpc.fzModel.AddConstraint(SetInConstraint, []string{varAtPos(sinkIntfcVarname, pathLength), endIntfcIndexes, TrueValue})

	// Finally, make sure a storage account is assigned iff there is a sink interface and it is non-virtual
	if dpc.problemData.Context.Flow == taxonomy.WriteFlow && !dpc.problemData.Context.Requirements.FlowParams.IsNewDataSet {
		for pathPos := 1; pathPos <= pathLength; pathPos++ {
			dpc.fzModel.AddConstraint(IntEqConstraint,
				[]string{varAtPos(saVarname, pathPos), strconv.Itoa(dpc.noStorageAccountVal), TrueValue})
		}
		return // no need to allocate storage, write destination is known
	}

	noSaRequiredModCaps := []string{}
	for modCapIdx, modCap := range dpc.modulesCapabilities {
		if modCap.virtualSink || !modCap.hasSink {
			noSaRequiredModCaps = append(noSaRequiredModCaps, strconv.Itoa(modCapIdx+1))
		}
	}
	noSaRequiredVarName := dpc.addSetInIndicator(modCapVarname, noSaRequiredModCaps, pathLength)
	realSA := dpc.addEqualityIndicator(saVarname, dpc.noStorageAccountVal, pathLength, false)
	for pathPos := 1; pathPos <= pathLength; pathPos++ {
		noSaRequiredAtPos := varAtPos(noSaRequiredVarName, pathPos)
		realSAAtPos := varAtPos(realSA, pathPos)
		dpc.fzModel.AddConstraint(BoolNotEqConstraint, []string{realSAAtPos, noSaRequiredAtPos})
	}
}

// Return a list of indexes of interfaces that match the input interface
func (dpc *DataPathCSP) getMatchingInterfaces(refIntfc *taxonomy.Interface) []string {
	res := []string{}
	for intfc, intfcIdx := range dpc.interfaceIdx {
		if interfacesMatch(refIntfc, &intfc) {
			res = append(res, strconv.Itoa(intfcIdx))
		}
	}
	return res
}

// Add constraints to ensure interface selection matches module-capability selection
func (dpc *DataPathCSP) modCapSupportsIntfc(pathLength int) {
	for intfc, intfcIdx := range dpc.interfaceIdx {
		for modCapIdx, modCap := range dpc.modulesCapabilities {
			modcapSupportsIntfcSrc := false
			modcapSupportsIntfcSink := false
			for _, modIntfc := range modCap.capability.SupportedInterfaces {
				modcapSupportsIntfcSrc = modcapSupportsIntfcSrc || interfacesMatch(modIntfc.Source, &intfc)
				modcapSupportsIntfcSink = modcapSupportsIntfcSink || interfacesMatch(modIntfc.Sink, &intfc)
			}
			if modCap.virtualSource || modCap.virtualSink {
				capAPI := modCap.capability.API
				apiIntfc := &taxonomy.Interface{Protocol: capAPI.Connection.Name, DataFormat: capAPI.DataFormat}
				modcapSupportsIntfcSrc = modcapSupportsIntfcSrc || modCap.virtualSource && interfacesMatch(apiIntfc, &intfc)
				modcapSupportsIntfcSink = modcapSupportsIntfcSink || modCap.virtualSink && interfacesMatch(apiIntfc, &intfc)
			}
			if !modcapSupportsIntfcSrc {
				dpc.preventAssignments([]string{modCapVarname, srcIntfcVarname}, []int{modCapIdx + 1, intfcIdx}, pathLength)
			}
			if !modcapSupportsIntfcSink {
				dpc.preventAssignments([]string{modCapVarname, sinkIntfcVarname}, []int{modCapIdx + 1, intfcIdx}, pathLength)
			}
		}
	}
}

// If there are optimization goals set, defines appropriate variables and sets the CSP-solver optimization goal
// Otherwise, just sets the CSP-solver goal as "satisfy"
func (dpc *DataPathCSP) addOptimizationGoals(pathLength int) error {
	const floatToIntRatio = 100.
	goalVarnames := []string{}
	weights := []string{}
	for _, goal := range dpc.problemData.Configuration.OptimizationStrategy {
		goalVarname, weight, err := dpc.addAnOptimizationGoal(goal, pathLength)
		if err != nil {
			return err
		}
		if goalVarname == "" {
			continue
		}
		floatWeight := 1.
		if weight != "" {
			floatWeight, err = strconv.ParseFloat(weight, 64) //nolint:revive,gomnd // Ignore magic number 64
			if err != nil {
				return err
			}
		}
		goalVarnames = append(goalVarnames, goalVarname)
		weights = append(weights, strconv.Itoa(int(floatWeight*floatToIntRatio)))
	}

	if len(goalVarnames) == 0 { // No optimization goals. Just satisfy constraints
		dpc.fzModel.SetSolveTarget(Satisfy, "")
	} else {
		dpc.fzModel.AddVariable(jointGoalVarname, IntType, true, true)
		dpc.setVarAsWeightedSum(jointGoalVarname, goalVarnames, weights)
		dpc.fzModel.SetSolveTarget(Minimize, jointGoalVarname)
	}
	return nil
}

// Adds variables to calculate the value of a single optimization goal
// Returns the variable containing the goal's value and its relative weight (as a string)
func (dpc *DataPathCSP) addAnOptimizationGoal(goal adminconfig.AttributeOptimization, pathLen int) (string, string, error) {
	weight := goal.Weight
	if goal.Directive == adminconfig.Maximize && weight != "" {
		weight = "-" + weight
	}

	attribute := goal.Attribute
	instanceTypes := dpc.env.AttributeManager.GetInstanceTypes(attribute)
	if len(instanceTypes) == 0 {
		return "", "", fmt.Errorf("no infrastructure data for attribute %s", attribute)
	}
	sanitizedAttr := sanitizeFznIdentifier(attribute)
	goalVarname := fmt.Sprintf("goal%s", sanitizedAttr)

	var err error
	if instanceTypes[0] == taxonomy.InterRegion { // The attribute is defined over region-pairs (e.g., bandwidth)
		err = dpc.setInterRegionGoalVarArray(attribute, goalVarname, pathLen)
	} else {
		err = dpc.setSimpleGoalVarArray(attribute, instanceTypes, goalVarname, pathLen)
	}
	if err != nil {
		return "", "", err
	}

	goalSumVarname := fmt.Sprintf("goal%sSum", sanitizedAttr)
	dpc.fzModel.AddVariable(goalSumVarname, IntType, true, false)
	dpc.setVarAsSimpleSumOfVarArray(goalSumVarname, goalVarname)
	return goalSumVarname, weight, nil
}

// For the given attribute, add a variable in the goalVarname array for each relevant instance type
// to hold the sum of the attribute values specified for the selected instances of this instance type.
// For example, goalVarname[1] may hold the sum of all storage account costs, while goalVarname[2] may hold the sum of all cluster costs.
// Summing the entries in goalVarname[1] and goalVarname[2] will yield the total cost of the selected instances.
func (dpc *DataPathCSP) setSimpleGoalVarArray(attr string, instanceTypes []taxonomy.InstanceType, goalVarname string, pathLen int) error {
	dpc.fzModel.AddVariableArray(goalVarname, IntType, len(instanceTypes), true, false)
	for idx, instanceType := range instanceTypes {
		instanceTypeGoalVarName := fmt.Sprintf("%s%s", goalVarname, instanceType)
		err := dpc.setGoalVarArrayForInstanceType(attr, instanceType, instanceTypeGoalVarName, pathLen)
		if err != nil {
			return err
		}
		dpc.setVarAsSimpleSumOfVarArray(varAtPos(goalVarname, idx+1), instanceTypeGoalVarName)
	}
	return nil
}

// For a given attribute and a given instance type (module/cluster/storage-account), build a var array called goalVarName,
// where the i-th element is the value of the attribute of the selected module/cluster/storage-account at path position i
func (dpc *DataPathCSP) setGoalVarArrayForInstanceType(attr string, instanceType taxonomy.InstanceType,
	goalVarname string, pathLen int) error {
	selectorVar, paramArray, err := dpc.getAttributeMapping(attr, instanceType)
	if err != nil {
		return err
	}

	dpc.fzModel.AddVariableArray(goalVarname, IntType, pathLen, true, false)
	for pos := 1; pos <= pathLen; pos++ {
		selectorVarAtPos := varAtPos(selectorVar, pos)
		goalAtPos := varAtPos(goalVarname, pos)
		definesAnnotation := GetDefinesVarAnnotation(goalAtPos)
		dpc.fzModel.AddConstraint(ArrIntElemConstraint, []string{selectorVarAtPos, paramArray, goalAtPos}, definesAnnotation)
	}
	return nil
}

// This creates a param array with the values of the given attribute for each cluster/module/storage account instance
// NOTE: We currently assume all values are integers. Code should be changed if some values are floats.
func (dpc *DataPathCSP) getAttributeMapping(attr string, instanceType taxonomy.InstanceType) (string, string, error) {
	resArray := []string{}
	varName := ""
	switch instanceType {
	case taxonomy.Cluster:
		varName = clusterVarname
		for _, cluster := range dpc.env.Clusters {
			infraElementValue, err := dpc.env.AttributeManager.GetNormalizedAttributeValue(attr, cluster.Name)
			if err != nil {
				return "", "", err
			}
			resArray = append(resArray, infraElementValue)
		}
	case taxonomy.StorageAccount:
		varName = saVarname
		for _, sa := range dpc.env.StorageAccounts {
			infraElementValue, err := dpc.env.AttributeManager.GetNormalizedAttributeValue(attr, sa.Name)
			if err != nil {
				infraElementValue, err = dpc.env.AttributeManager.GetNormalizedAttributeValue(attr, sa.GenerateName)
				if err != nil {
					return "", "", err
				}
			}
			resArray = append(resArray, infraElementValue)
		}
		resArray = append(resArray, zeroStr) // Assuming attribute == 0 if no storage account is set
	case taxonomy.Module:
		varName = modCapVarname
		for _, modCap := range dpc.modulesCapabilities {
			infraElementValue, err := dpc.env.AttributeManager.GetNormalizedAttributeValue(attr, modCap.module.Name)
			if err != nil {
				return "", "", err
			}
			resArray = append(resArray, infraElementValue)
		}
	default:
		return "", "", fmt.Errorf("unknown instance type %s", instanceType)
	}
	if len(resArray) < 1 { // e.g. if there are no storage accounts
		return "", "", nil
	}
	paramName := varName + sanitizeFznIdentifier(attr)
	dpc.fzModel.AddParamArray(paramName, IntType, len(resArray), fznCompoundLiteral(resArray, false))
	return varName, paramName, nil
}

// This will set the goal array according to the given attribute, which is defined over region-pairs (e.g., bandwidth)
// goalArray[i] is the attr value from cluster i to cluster i+1 (cluster pathLen+1 is the workload)
// If i <= (the position of the last data-store), then goalArray[i] is 0
// goalArray[pathLen+1] is the attr value from the last data-store on the pipe to the next cluster (or the workload)
func (dpc *DataPathCSP) setInterRegionGoalVarArray(attr, goalVarname string, pathLen int) error {
	dpc.setInterRegionGoalsCommonVars(pathLen)

	c2cParamName, err := dpc.getCluster2ClusterParamArray(attr)
	if err != nil {
		return err
	}
	s2cParamName, err := dpc.getStorageToClusterParamArray(attr)
	if err != nil {
		return err
	}

	dpc.fzModel.AddVariableArray(goalVarname, IntType, pathLen+1, true, false)
	c2cSelectedValueName := c2cParamName + "SelectedValue" // The value selected from the c2cParamArray
	// val at pos i is the attr value from cluster i to cluster i+1 (cluster pathLen+1 is the workload)
	dpc.fzModel.AddVariableArray(c2cSelectedValueName, IntType, pathLen, true, false)

	for pos := 1; pos <= pathLen; pos++ {
		c2cSelectorAtPos := varAtPos(c2cSelectorVarname, pos)
		c2cSelectedValueAtPos := varAtPos(c2cSelectedValueName, pos)
		c2cDefinesAnnotation := GetDefinesVarAnnotation(c2cSelectedValueAtPos)
		dpc.fzModel.AddConstraint(ArrIntElemConstraint, []string{c2cSelectorAtPos, c2cParamName, c2cSelectedValueAtPos}, c2cDefinesAnnotation)

		// goalVarname is set to the attr value from cluster i to cluster i+1 if i > maxRealSa, otherwise it is set to 0
		dpc.assignWithSelector(goalVarname, afterMaxRealSaVarName,
			arrayOfVarPositions(c2cSelectedValueName, pathLen), arrayOfSameInt(0, pathLen), pathLen)
	}

	// Finally, set goalVarname[pathLen+1] to be the attr value between the last data-store and the next cluster
	goalAtPos := varAtPos(goalVarname, pathLen+1)
	dpc.fzModel.AddConstraint(ArrIntElemConstraint, []string{s2cSelectorVarname, s2cParamName, goalAtPos})
	return nil
}

// This function declares and sets variables which are common to setting the value of all inter-region goals
// In particular, it declares the following variables, which are used in setInterRegionGoalVarArray()
//  c2cSelectorVarname[i] is the selector variable for the c2cParamArray at position i
//                        that is, the attribute value between which two cluster to take
//  s2cSelectorVarname is the selector variable for the s2cParamArray (),
//                     that is, the attribute value between which storage-account and cluster to take
// afterMaxRealSaVarName[i] is true iff the cluster at position i is after the last data store

func (dpc *DataPathCSP) setInterRegionGoalsCommonVars(pathLen int) {
	if _, defined := dpc.fzModel.VarMap[realSaLocationsVarName]; defined {
		return // Common vars already set
	}

	dpc.fzModel.AddVariableArray(storageLocsVarname, IntType, pathLen+1, true, false)
	storageLocsAssignment := append([]string{zeroStr}, dpc.fzModel.varArrayElements(saVarname)...)
	dpc.fzModel.SetVariableAssignment(storageLocsVarname, fznCompoundLiteral(storageLocsAssignment, false))
	realSA := dpc.addEqualityIndicator(storageLocsVarname, dpc.noStorageAccountVal, pathLen+1, false)
	virtualSA := dpc.addEqualityIndicator(storageLocsVarname, dpc.noStorageAccountVal, pathLen+1, true)
	realSaLocationsVarType := fznRangeVarType(0, pathLen+1)
	dpc.fzModel.AddVariableArray(realSaLocationsVarName, realSaLocationsVarType, pathLen+1, true, false)
	for pos := 1; pos <= pathLen+1; pos++ {
		// If a storage account is allocated at pos, set realSaLocations[pos] to pos, otherwise, set to 0
		realSAAtPos := varAtPos(realSA, pos)
		virtualSAAtPos := varAtPos(virtualSA, pos)
		realSaLocationsAtPos := varAtPos(realSaLocationsVarName, pos)
		dpc.fzModel.AddConstraint(IntEqConstraint, []string{realSaLocationsAtPos, strconv.Itoa(pos), realSAAtPos})
		dpc.fzModel.AddConstraint(IntEqConstraint, []string{realSaLocationsAtPos, zeroStr, virtualSAAtPos})
	}

	dpc.fzModel.AddVariable(maxRealSaVarName, realSaLocationsVarType, true, false)
	dpc.fzModel.AddConstraint(IntMaxConstraint, []string{maxRealSaVarName, realSaLocationsVarName})

	dpc.fzModel.AddVariableArray(afterMaxRealSaVarName, BoolType, pathLen, true, false)
	c2cSelectorType := fznRangeVarType(1, len(dpc.env.Clusters)*len(dpc.env.Clusters))
	dpc.fzModel.AddVariableArray(c2cSelectorVarname, c2cSelectorType, pathLen, true, false)
	numClustersStr := strconv.Itoa(len(dpc.env.Clusters))
	selectorWeights := []string{numClustersStr, oneStr, minusOneStr}
	for pos := 1; pos <= pathLen; pos++ {
		// set afterMaxRealSA[pos] to true iff pos is after the last storage account on the pipe
		afterMaxRealSaAtPos := varAtPos(afterMaxRealSaVarName, pos)
		c2cSelectorAtPos := varAtPos(c2cSelectorVarname, pos)
		clusterAtPos := varAtPos(clusterVarname, pos)
		clusterAtNextPos := varAtPos(clusterVarname, pos+1)
		dpc.fzModel.AddConstraint(IntLeConstraint, []string{maxRealSaVarName, strconv.Itoa(pos), afterMaxRealSaAtPos})
		dpc.setVarAsWeightedSum(c2cSelectorAtPos, []string{clusterAtPos, clusterAtNextPos, numClustersStr}, selectorWeights)
	}

	lastDataStoreVarType := fznRangeVarType(0, len(dpc.env.StorageAccounts))
	dpc.fzModel.AddVariable(lastDataStoreVarname, lastDataStoreVarType, true, false)
	dpc.fzModel.AddConstraint(ArrIntElemConstraint, []string{maxRealSaVarName, storageLocsVarname, lastDataStoreVarname})
	clusterAfterLastDataStoreVarType := fznRangeVarType(1, len(dpc.env.Clusters))
	dpc.fzModel.AddVariable(clusterAfterLastDataStoreVarname, clusterAfterLastDataStoreVarType, true, false)
	dpc.fzModel.AddConstraint(ArrIntElemConstraint, []string{maxRealSaVarName, clusterVarname, clusterAfterLastDataStoreVarname})
	s2cSelectorType := fznRangeVarType(1, len(dpc.env.Clusters)*(len(dpc.env.StorageAccounts)+1))
	dpc.fzModel.AddVariable(s2cSelectorVarname, s2cSelectorType, true, false)
	selectorWeights = []string{numClustersStr, oneStr}
	dpc.setVarAsWeightedSum(s2cSelectorVarname, []string{lastDataStoreVarname, clusterAfterLastDataStoreVarname}, selectorWeights)
}

// Produces a paramArray containing the attr value for each pair of clusters
func (dpc *DataPathCSP) getCluster2ClusterParamArray(attr string) (string, error) {
	c2cParamArray := []string{}
	for _, cluster1 := range dpc.env.Clusters {
		for _, cluster2 := range dpc.env.Clusters {
			value, err := dpc.env.AttributeManager.GetNormAttrValFromArgs(attr, cluster1.Metadata.Region, cluster2.Metadata.Region)
			if err != nil {
				return "", err
			}
			c2cParamArray = append(c2cParamArray, value)
		}
	}
	c2cParamName := "cluster2cluster" + sanitizeFznIdentifier(attr)
	dpc.fzModel.AddParamArray(c2cParamName, IntType, len(c2cParamArray), fznCompoundLiteral(c2cParamArray, false))
	return c2cParamName, nil
}

// Produces a paramArray containing the attr value for each pair of storage-account and cluster
// The first line of the resulting matrix describes the attr value of the dataset-region vs each cluster
func (dpc *DataPathCSP) getStorageToClusterParamArray(attr string) (string, error) {
	s2cParamArray := []string{}
	dataSetRegion := dpc.problemData.DataDetails.ResourceMetadata.Geography
	for _, cluster := range dpc.env.Clusters {
		value, err := dpc.env.AttributeManager.GetNormAttrValFromArgs(attr, dataSetRegion, cluster.Metadata.Region)
		if err != nil {
			return "", err
		}
		s2cParamArray = append(s2cParamArray, value)
	}
	for _, sa := range dpc.env.StorageAccounts {
		for _, cluster := range dpc.env.Clusters {
			value, err := dpc.env.AttributeManager.GetNormAttrValFromArgs(attr, string(sa.Spec.Region), cluster.Metadata.Region)
			if err != nil {
				return "", err
			}
			s2cParamArray = append(s2cParamArray, value)
		}
	}

	s2cParamName := "sa2cluster" + sanitizeFznIdentifier(attr)
	dpc.fzModel.AddParamArray(s2cParamName, IntType, len(s2cParamArray), fznCompoundLiteral(s2cParamArray, false))
	return s2cParamName, nil
}

// Sets the CSP int variable sumVarname to be the weighted sum of int elements in arrayToSum.
// The integer weight of each element is given in the array "weights".
// FlatZinc doesn't give us a "weighted sum" constraint (and not even sum constraint).
// The trick is to use the dot-product constraint, add the summing var with weight -1 and force the result to be 0
func (dpc *DataPathCSP) setVarAsWeightedSum(sumVarname string, arrayToSum, weights []string) {
	arrayToSum = append(arrayToSum, sumVarname)
	weights = append(weights, minusOneStr)
	dpc.fzModel.AddConstraint(
		IntLinEqConstraint,
		[]string{fznCompoundLiteral(weights, false), fznCompoundLiteral(arrayToSum, false), strconv.Itoa(0)},
		GetDefinesVarAnnotation(sumVarname),
	)
}

// Sets the CSP int variable sumVarname to be the weighted sum of the elements in the variable array varArrayToSum.
func (dpc *DataPathCSP) setVarAsWeightedSumOfVarArray(sumVarname, varArrayToSum string, weightsArray []string) {
	arrayToSum := arrayOfVarPositions(varArrayToSum, len(weightsArray))
	dpc.setVarAsWeightedSum(sumVarname, arrayToSum, weightsArray)
}

// Sets the CSP int variable sumVarname to be the sum of the elements in the variable array varArrayToSum.
func (dpc *DataPathCSP) setVarAsSimpleSumOfVarArray(sumVarname, varArrayToSum string) {
	arrayLen := dpc.fzModel.GetVariableSize(varArrayToSum)
	dpc.setVarAsWeightedSumOfVarArray(sumVarname, varArrayToSum, arrayOfSameInt(1, arrayLen))
}

// "varToAssign" gets assigned with "valIfTrue" if "selectorVar" is true, and with "valIfFalse" otherwise
func (dpc *DataPathCSP) assignWithSelector(varToAssign, selectorVar string, valIfTrue, valIfFalse []string, pathLen int) {
	impliedBySelector := dpc.addImpliedByIndicator(selectorVar, pathLen, false)
	impliedByNotSelector := dpc.addImpliedByIndicator(selectorVar, pathLen, true)
	for pos := 1; pos <= pathLen; pos++ {
		varToAssignAtPos := varAtPos(varToAssign, pos)
		impliedBySelectorAtPos := varAtPos(impliedBySelector, pos)
		impliedByNotSelectorAtPos := varAtPos(impliedByNotSelector, pos)
		dpc.fzModel.AddConstraint(IntEqConstraint, []string{varToAssignAtPos, valIfTrue[pos-1], impliedBySelectorAtPos})
		dpc.fzModel.AddConstraint(IntEqConstraint, []string{varToAssignAtPos, valIfFalse[pos-1], impliedByNotSelectorAtPos})
	}
}

// Returns which actions should be activated by the module at position pathPos (according to the solver's solution)
func (dpc *DataPathCSP) getSolutionActionsAtPos(solverSolution CPSolution, pathPos int) []taxonomy.Action {
	actions := []taxonomy.Action{}
	for actionVarname, action := range dpc.requiredActions {
		actionSolution := solverSolution[actionVarname]
		if actionSolution[pathPos] == TrueValue {
			actions = append(actions, action)
		}
	}
	return actions
}

// Translates a solver's solution into a FybrikApplication Solution for a given data-path
// Also returns the score of the solution (the smaller the better) if such exists, and NaN otherwise
// TODO: better handle error messages
func (dpc *DataPathCSP) decodeSolverSolution(solverSolutionStr string, pathLen int) (datapath.Solution, float64, error) {
	solverSolution, err := dpc.fzModel.ReadBestSolution(solverSolutionStr)
	if err != nil {
		return datapath.Solution{}, math.NaN(), err
	}
	if len(solverSolution) == 0 {
		return datapath.Solution{}, math.NaN(), nil // UNSAT
	}

	modCapSolution := solverSolution[modCapVarname]
	clusterSolution := solverSolution[clusterVarname]
	saSolution := solverSolution[saVarname]
	srcIntfcSolution := solverSolution[srcIntfcVarname]
	sinkIntfcSolution := solverSolution[sinkIntfcVarname]

	srcIntfcIdx, _ := strconv.Atoi(srcIntfcSolution[0])
	srcNode := &datapath.Node{Connection: dpc.reverseIntfcMap[srcIntfcIdx]}

	solution := datapath.Solution{}
	for pathPos := 0; pathPos < pathLen; pathPos++ {
		modCapIdx, _ := strconv.Atoi(modCapSolution[pathPos])
		modCap := dpc.modulesCapabilities[modCapIdx-1]
		clusterIdx, _ := strconv.Atoi(clusterSolution[pathPos])
		saIdx, _ := strconv.Atoi(saSolution[pathPos])
		sa := appApi.FybrikStorageAccountSpec{}
		if saIdx != dpc.noStorageAccountVal {
			sa = dpc.env.StorageAccounts[saIdx-1].Spec
		}
		sinkIntfcIdx, _ := strconv.Atoi(sinkIntfcSolution[pathPos])
		sinkNode := &datapath.Node{Connection: dpc.reverseIntfcMap[sinkIntfcIdx], Virtual: modCap.virtualSink}
		edge := datapath.Edge{Module: modCap.module, CapabilityIndex: modCap.capabilityIdx, Source: srcNode, Sink: sinkNode}
		resolvedEdge := datapath.ResolvedEdge{
			Edge:           edge,
			Actions:        dpc.getSolutionActionsAtPos(solverSolution, pathPos),
			Cluster:        dpc.env.Clusters[clusterIdx-1].Name,
			StorageAccount: sa,
		}
		solution.DataPath = append(solution.DataPath, &resolvedEdge)
		srcNode = sinkNode
	}

	if dpc.problemData.Context.Flow == taxonomy.WriteFlow {
		solution.Reverse()
	}

	score := math.NaN()
	if scoreStr, found := solverSolution[jointGoalVarname]; found {
		score, err = strconv.ParseFloat(scoreStr[0], 64) //nolint:revive,gomnd // Ignore magic number 64
		if err != nil {
			score = math.NaN()
		}
	}

	return solution, score, nil
}

// ----- helper functions -----

func encodingComment(index int, encodedVal string) string {
	return fmt.Sprintf("%d - %s", index, encodedVal)
}

func getActionVarname(action taxonomy.Action) string {
	return fmt.Sprintf(actionVarname, action.Name)
}

func arrayOfSameStr(str string, arrayLen int) []string {
	array := make([]string, arrayLen)
	for i := 0; i < arrayLen; i++ {
		array[i] = str
	}
	return array
}

func arrayOfSameInt(num, arrayLen int) []string {
	return arrayOfSameStr(strconv.Itoa(num), arrayLen)
}

func getAssetInterface(connection *datacatalog.GetAssetResponse) taxonomy.Interface {
	if connection == nil || connection.Details.Connection.Name == "" {
		return taxonomy.Interface{Protocol: utils.GetDefaultConnectionType(), DataFormat: ""}
	}
	return taxonomy.Interface{Protocol: connection.Details.Connection.Name, DataFormat: connection.Details.DataFormat}
}

func arrayOfVarPositions(variableArray string, arrayLen int) []string {
	array := make([]string, arrayLen)
	for i := 1; i <= arrayLen; i++ {
		array[i-1] = varAtPos(variableArray, i)
	}
	return array
}

func getWorkloadClusterIndex(wlCluster multicluster.Cluster, clusters []multicluster.Cluster) string {
	for i, cluster := range clusters {
		if cluster.Name == wlCluster.Name {
			return strconv.Itoa(i + 1)
		}
	}
	return oneStr // Note: this shouldn't really happen. We assume the workload cluster should always be one of the cluster in env.Clusters
}

// returns whether an interface supported by a module (at source or at sink) matches another interface
func interfacesMatch(moduleIntfc, otherIntfc *taxonomy.Interface) bool {
	if moduleIntfc == nil {
		moduleIntfc = &taxonomy.Interface{}
	}
	if otherIntfc == nil {
		otherIntfc = &taxonomy.Interface{}
	}
	if moduleIntfc.Protocol != otherIntfc.Protocol {
		return false
	}

	// an empty DataFormat in the module's interface means it supports all formats
	return moduleIntfc.DataFormat == "" || moduleIntfc.DataFormat == otherIntfc.DataFormat
}
