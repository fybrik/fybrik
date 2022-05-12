// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"fmt"
	"strconv"
	"strings"

	appApi "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
)

// Names of the primary variables on which we need to take decisions
// Each variable is an array of ints. The i-th cell in the array represents a decision for the i-th Node/Edge in the data path
const (
	modCapVarname    = "moduleCapability"      // Var's value says which capability of which module to use (on each Edge)
	clusterVarname   = "moduleCluster"         // Var's value says which of the available clusters to use (on each Edge)
	saVarname        = "storageAccount"        // Var's value says which storage account to use (0 means no sa)
	srcIntfcVarname  = "moduleSourceInterface" // Var's value says which interface to use as source
	sinkIntfcVarname = "moduleSinkInterface"   // Var's value says which interface to use as sink
	actionVarname    = "action_%s"             // Vars for each required action, say whether the action was applied
	jointGoalVarname = "jointGoal"             // Var's value indicates the quality of the data path w.r.t. optimization goals
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
	problemData         *DataInfo
	env                 *Environment
	modulesCapabilities []moduleAndCapability       // An enumeration of allowed capabilities in all modules
	interfaceIdx        map[taxonomy.Interface]int  // gives an index for each unique interface
	reverseIntfcMap     map[int]*taxonomy.Interface // The reverse mapping (needed when decoding the solution)
	fzModel             *FlatZincModel
}

// The ctor also enumerates all available (module x capabilities) and all available interfaces
// The generated enumerations are listed at the header of the FlatZinc model
func NewDataPathCSP(problemData *DataInfo, env *Environment) *DataPathCSP {
	dpCSP := DataPathCSP{problemData: problemData, env: env, fzModel: NewFlatZincModel()}
	dpCSP.interfaceIdx = map[taxonomy.Interface]int{}
	dpCSP.reverseIntfcMap = map[int]*taxonomy.Interface{}
	dataSetIntfc := getAssetInterface(dpCSP.problemData.DataDetails)
	dpCSP.addAndGetInterface(nil)           // ensure nil interface always gets index 0
	dpCSP.addAndGetInterface(&dataSetIntfc) // data-set interface always gets index 1 (cannot be nil)
	dpCSP.addAndGetInterface(dpCSP.problemData.Context.Requirements.Interface)

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
		comment = fmt.Sprintf("%d - Protocol: %s, DataFormat: %s", intfcIdx, intfc.Protocol, intfc.DataFormat)
		dpCSP.fzModel.AddHeaderComment(comment)
	}
	dpCSP.fzModel.AddHeaderComment("Encoding of clusters:")
	for clusterIdx, cluster := range dpCSP.env.Clusters {
		comment = fmt.Sprintf("%d - %s", clusterIdx+1, cluster.Name)
		dpCSP.fzModel.AddHeaderComment(comment)
	}
	dpCSP.fzModel.AddHeaderComment("Encoding of storage accounts:")
	for saIdx, sa := range dpCSP.env.StorageAccounts {
		comment = fmt.Sprintf("%d - %s", saIdx+1, sa.Name)
		dpCSP.fzModel.AddHeaderComment(comment)
	}
	return &dpCSP
}

// Add the interfaces defined in a given module's capability to the 2 interface maps
func (dpc *DataPathCSP) addModCapInterfacesToMaps(modcap *moduleAndCapability) {
	capability := modcap.capability
	for _, intfc := range capability.SupportedInterfaces {
		if intfc.Source != nil {
			dpc.addAndGetInterface(intfc.Source)
			modcap.hasSource = true
		}
		if intfc.Sink != nil {
			dpc.addAndGetInterface(intfc.Sink)
			modcap.hasSink = true
		}
	}
	if (!modcap.hasSource || !modcap.hasSink) && capability.API != nil {
		apiInterface := &taxonomy.Interface{Protocol: capability.API.Connection.Name, DataFormat: capability.API.DataFormat}
		dpc.addAndGetInterface(apiInterface)
		modcap.virtualSource = !modcap.hasSource
		modcap.virtualSink = !modcap.hasSink
		modcap.hasSource = true
		modcap.hasSink = true
	}
}

// Add the given interface to the 2 interface maps (but avoid duplicates)
func (dpc *DataPathCSP) addAndGetInterface(intfc *taxonomy.Interface) int {
	if intfc == nil {
		intfc = &taxonomy.Interface{}
	}
	intfcIdx, found := dpc.interfaceIdx[*intfc]
	if !found {
		intfcIdx = len(dpc.interfaceIdx)
		dpc.interfaceIdx[*intfc] = intfcIdx
		dpc.reverseIntfcMap[intfcIdx] = intfc
	}
	return intfcIdx
}

// This is the main method for building a FlatZinc CSP out of the data-path parameters and constraints.
// Returns a file name where the model was dumped
// NOTE: Minimal index of FlatZinc arrays is always 1. Hence, we use 1-based modeling all over the place to avoid confusion
//       The two exceptions are storage accounts, where a value of 0 means no storage account and interfaces (0 means nil)
func (dpc *DataPathCSP) BuildFzModel(pathLength int) (string, error) {
	dpc.fzModel.Clear() // This function can be called multiple times - clear vars and constraints from last call
	// Variables to select the module capability we use on each data-path location
	moduleCapabilityVarType := fznRangeVarType(1, len(dpc.modulesCapabilities))
	dpc.fzModel.AddVariableArray(modCapVarname, moduleCapabilityVarType, pathLength, false, true)
	// Variables to select storage-accounts to place on each data-path location (the value 0 means no storage account)
	saTypeVarType := fznRangeVarType(0, len(dpc.env.StorageAccounts))
	dpc.fzModel.AddVariableArray(saVarname, saTypeVarType, pathLength, false, true)
	// Variables to select the cluster we allocate to each module on the path
	moduleClusterVarType := fznRangeVarType(1, len(dpc.env.Clusters))
	dpc.fzModel.AddVariableArray(clusterVarname, moduleClusterVarType, pathLength, false, true)
	// Variables to select the source and sink interface for each module on the path
	moduleInterfaceVarType := fznRangeVarType(0, len(dpc.interfaceIdx)-1)
	dpc.fzModel.AddVariableArray(srcIntfcVarname, moduleInterfaceVarType, pathLength, false, true)
	dpc.fzModel.AddVariableArray(sinkIntfcVarname, moduleInterfaceVarType, pathLength, false, true)

	dpc.addGovernanceActionConstraints(pathLength)
	err := dpc.addAdminConfigRestrictions(pathLength)
	if err != nil {
		return "", err
	}
	dpc.addInterfaceConstraints(pathLength)
	err = dpc.addOptimizationGoals(pathLength)
	if err != nil {
		return "", err
	}

	return dpc.fzModel.Dump()
}

// enforce restrictions from admin configuration decisions:
// a. cluster satisfies restrictions for the selected capability
// b. storage account satisfies restrictions for the selected capability
// c. Ensure capabilities that must be deployed are indeed deployed
func (dpc *DataPathCSP) addAdminConfigRestrictions(pathLength int) error {
	for decCapability := range dpc.problemData.Configuration.ConfigDecisions {
		decision := dpc.problemData.Configuration.ConfigDecisions[decCapability]
		relevantModCaps := []string{}
		for modCapIdx, moduleCap := range dpc.modulesCapabilities {
			if moduleCap.capability.Capability != decCapability {
				continue
			}
			relevantModCaps = append(relevantModCaps, strconv.Itoa(modCapIdx+1))
			for clusterIdx, cluster := range dpc.env.Clusters {
				if !dpc.clusterSatisfiesRestrictions(cluster, decision.DeploymentRestrictions.Clusters) {
					dpc.preventAssignments([]string{modCapVarname, clusterVarname},
						[]int{modCapIdx + 1, clusterIdx + 1}, pathLength)
				}
			}
			for saIdx, sa := range dpc.env.StorageAccounts {
				if !dpc.saSatisfiesRestrictions(sa, decision.DeploymentRestrictions.StorageAccounts) {
					dpc.preventAssignments([]string{modCapVarname, saVarname},
						[]int{modCapIdx + 1, saIdx + 1}, pathLength)
				}
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

// Replicates a constraint to block a specific assignment for each position in the path
func (dpc *DataPathCSP) preventAssignments(variables []string, values []int, pathLength int) {
	indicators := []string{}
	for idx, variable := range variables {
		indicators = append(indicators, dpc.addEqualityIndicator(variable, values[idx], pathLength, false))
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

	if !dpc.problemData.Context.Requirements.FlowParams.IsNewDataSet {
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

	// Now, ensure interfaces match along the datapath from dataset to workload
	startIntfcIdx := strconv.Itoa(1)
	endIntfcIdx := strconv.Itoa(dpc.addAndGetInterface(dpc.problemData.Context.Requirements.Interface))
	if dpc.problemData.Context.Flow == taxonomy.WriteFlow {
		startIntfcIdx, endIntfcIdx = endIntfcIdx, startIntfcIdx // swap start and end for write flows
	}
	dpc.fzModel.AddConstraint(IntEqConstraint, []string{varAtPos(srcIntfcVarname, 1), startIntfcIdx, TrueValue})
	for pathPos := 1; pathPos < pathLength; pathPos++ {
		dpc.fzModel.AddConstraint(IntEqConstraint,
			[]string{varAtPos(sinkIntfcVarname, pathPos), varAtPos(srcIntfcVarname, pathPos+1), TrueValue})
	}
	dpc.fzModel.AddConstraint(IntEqConstraint, []string{varAtPos(sinkIntfcVarname, pathLength), endIntfcIdx, TrueValue})

	// Finally, make sure a storage account is assigned iff there is a sink interface and it is non-virtual
	if dpc.problemData.Context.Flow == taxonomy.WriteFlow && !dpc.problemData.Context.Requirements.FlowParams.IsNewDataSet {
		return // no need to allocate storage, write destination is known
	}

	noSaRequiredModCaps := []string{}
	for modCapIdx, modCap := range dpc.modulesCapabilities {
		if modCap.virtualSink || !modCap.hasSink {
			noSaRequiredModCaps = append(noSaRequiredModCaps, strconv.Itoa(modCapIdx+1))
		}
	}
	noSaRequiredVarName := dpc.addSetInIndicator(modCapVarname, noSaRequiredModCaps, pathLength)
	realSA := dpc.addEqualityIndicator(saVarname, 0, pathLength, false) // a value >0 means a storage account is allocated
	for pathPos := 1; pathPos <= pathLength; pathPos++ {
		noSaRequiredAtPos := varAtPos(noSaRequiredVarName, pathPos)
		realSAAtPos := varAtPos(realSA, pathPos)
		dpc.fzModel.AddConstraint(BoolNotEqConstraint, []string{realSAAtPos, noSaRequiredAtPos})
	}
}

// Add constraints to ensure interface selection matches module-capability selection
func (dpc *DataPathCSP) modCapSupportsIntfc(pathLength int) {
	for intfc, intfcIdx := range dpc.interfaceIdx {
		for modCapIdx, modCap := range dpc.modulesCapabilities {
			modcapSupportsIntfcSrc := false
			modcapSupportsIntfcSink := false
			for _, modifc := range modCap.capability.SupportedInterfaces {
				modcapSupportsIntfcSrc = modcapSupportsIntfcSrc || interfacesMatch(modifc.Source, &intfc)
				modcapSupportsIntfcSink = modcapSupportsIntfcSink || interfacesMatch(modifc.Sink, &intfc)
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
		floatWeight, err := strconv.ParseFloat(weight, 64)
		if err != nil {
			return err
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
	if goal.Directive == adminconfig.Maximize {
		weight = "-" + weight
	}

	attribute := goal.Attribute
	goalVarname := fmt.Sprintf("goal%s", attribute)
	dpc.fzModel.AddVariableArray(goalVarname, IntType, pathLen, true, false)

	goalVarNames := []string{}
	selectorVar, paramArray, err := dpc.getAttributeMapping(attribute)
	if err != nil {
		return "", "", err
	}
	for pos := 1; pos <= pathLen; pos++ {
		selectorVarAtPos := varAtPos(selectorVar, pos)
		goalAtPos := varAtPos(goalVarname, pos)
		goalVarNames = append(goalVarNames, goalAtPos)
		definesAnnotation := GetDefinesVarAnnotation(goalAtPos)
		dpc.fzModel.AddConstraint(ArrIntElemConstraint, []string{selectorVarAtPos, paramArray, goalAtPos}, definesAnnotation)
	}

	goalSumVarname := fmt.Sprintf("goal%sSum", attribute)
	dpc.fzModel.AddVariable(goalSumVarname, IntType, true, false)
	dpc.setVarAsWeightedSum(goalSumVarname, goalVarNames, arrayOfSameInt(1, pathLen))
	return goalSumVarname, weight, nil
}

// This creates a param array with the values of the given attribute for each cluster/module/storage account instance
// NOTE: We currently assume all values are integers. Code should be changed if some values are floats.
func (dpc *DataPathCSP) getAttributeMapping(attr taxonomy.Attribute) (string, string, error) {
	instanceType := dpc.env.AttributeManager.GetInstanceType(attr)
	if instanceType == nil {
		return "", "", fmt.Errorf("there are no clusters, modules or storage accounts with an attribute %s", attr)
	}

	resArray := []string{}
	varName := ""
	switch *instanceType {
	case taxonomy.Cluster:
		varName = clusterVarname
		for _, cluster := range dpc.env.Clusters {
			infraElement := dpc.env.AttributeManager.GetAttribute(attr, cluster.Name)
			if infraElement == nil {
				return "", "", fmt.Errorf("attribute %s is not defined for cluster %s", attr, cluster.Name)
			}
			resArray = append(resArray, infraElement.Value)
		}
	case taxonomy.StorageAccount:
		varName = saVarname
		for _, sa := range dpc.env.StorageAccounts {
			infraElement := dpc.env.AttributeManager.GetAttribute(attr, sa.Name)
			if infraElement == nil {
				return "", "", fmt.Errorf("attribute %s is not defined for storage account %s", attr, sa.Name)
			}
			resArray = append(resArray, infraElement.Value)
		}
	default: // should be taxonomy.Module
		varName = modCapVarname
		for _, modCap := range dpc.modulesCapabilities {
			infraElement := dpc.env.AttributeManager.GetAttribute(attr, modCap.module.Name)
			if infraElement == nil {
				return "", "", fmt.Errorf("attribute %s is not defined for module %s", attr, modCap.module.Name)
			}
			resArray = append(resArray, infraElement.Value)
		}
	}
	paramName := varName + string(attr)
	dpc.fzModel.AddParamArray(paramName, IntType, len(resArray), fznCompoundLiteral(resArray, false))
	return varName, paramName, nil
}

// Sets the CSP int variable sumVarname to be the weighted sum of int elements in arrayToSum.
// The integer weight of each element is given in the array "weights".
// FlatZinc doesn't give us a "weighted sum" constraint (and not even sum constraint).
// The trick is to use the dot-product constraint, add the summing var with weight -1 and force the result to be 0
func (dpc *DataPathCSP) setVarAsWeightedSum(sumVarname string, arrayToSum, weights []string) {
	arrayToSum = append(arrayToSum, sumVarname)
	weights = append(weights, "-1")
	dpc.fzModel.AddConstraint(
		IntLinEqConstraint,
		[]string{fznCompoundLiteral(weights, false), fznCompoundLiteral(arrayToSum, false), "0"},
		GetDefinesVarAnnotation(sumVarname),
	)
}

// Returns which actions should be activated by the module at position pathPos (according to the solver's solution)
func (dpc *DataPathCSP) getSolutionActionsAtPos(solverSolution CPSolution, pathPos int) []taxonomy.Action {
	actions := []taxonomy.Action{}
	for _, action := range dpc.problemData.Actions {
		actionVarname := getActionVarname(action)
		actionSolution := solverSolution[actionVarname]
		if actionSolution[pathPos] == TrueValue {
			actions = append(actions, action)
		}
	}
	return actions
}

// Translates a solver's solution into a FybrikApplication Solution for a given data-path
// TODO: better handle error messages
func (dpc *DataPathCSP) decodeSolverSolution(solverSolutionStr string, pathLen int) (Solution, error) {
	solverSolution, err := dpc.fzModel.ReadBestSolution(solverSolutionStr)
	if err != nil {
		return Solution{}, err
	}
	if len(solverSolution) == 0 {
		return Solution{}, nil // UNSAT
	}

	modCapSolution := solverSolution[modCapVarname]
	clusterSolution := solverSolution[clusterVarname]
	saSolution := solverSolution[saVarname]
	srcIntfcSolution := solverSolution[srcIntfcVarname]
	sinkIntfcSolution := solverSolution[sinkIntfcVarname]

	srcIntfcIdx, _ := strconv.Atoi(srcIntfcSolution[0])
	srcNode := &Node{Connection: dpc.reverseIntfcMap[srcIntfcIdx]}

	solution := Solution{}
	for pathPos := 0; pathPos < pathLen; pathPos++ {
		modCapIdx, _ := strconv.Atoi(modCapSolution[pathPos])
		modCap := dpc.modulesCapabilities[modCapIdx-1]
		clusterIdx, _ := strconv.Atoi(clusterSolution[pathPos])
		saIdx, _ := strconv.Atoi(saSolution[pathPos])
		sa := appApi.FybrikStorageAccountSpec{}
		if saIdx > 0 { // recall that a value of 0 means no storage account
			sa = dpc.env.StorageAccounts[saIdx-1].Spec
		}
		sinkIntfcIdx, _ := strconv.Atoi(sinkIntfcSolution[pathPos])
		sinkNode := &Node{Connection: dpc.reverseIntfcMap[sinkIntfcIdx], Virtual: modCap.virtualSink}
		edge := Edge{Module: modCap.module, CapabilityIndex: modCap.capabilityIdx, Source: srcNode, Sink: sinkNode}
		resolvedEdge := ResolvedEdge{
			Edge:           edge,
			Actions:        dpc.getSolutionActionsAtPos(solverSolution, pathPos),
			Cluster:        dpc.env.Clusters[clusterIdx-1].Name,
			StorageAccount: sa,
		}
		solution.DataPath = append(solution.DataPath, &resolvedEdge)
		srcNode = sinkNode
	}

	if dpc.problemData.Context.Flow == taxonomy.WriteFlow { // reverse solution
		for elementInd := 0; elementInd < len(solution.DataPath)/2; elementInd++ {
			reversedInd := len(solution.DataPath) - elementInd - 1
			solution.DataPath[elementInd], solution.DataPath[reversedInd] =
				solution.DataPath[reversedInd], solution.DataPath[elementInd]
		}
	}

	return solution, nil
}

// ----- helper functions -----

func getActionVarname(action taxonomy.Action) string {
	return fmt.Sprintf(actionVarname, action.Name)
}

func varAtPos(variable string, pos int) string {
	return fmt.Sprintf("%s[%d]", variable, pos)
}

func arrayOfSameInt(num, arrayLen int) []string {
	return strings.Fields(strings.Repeat(strconv.Itoa(num)+" ", arrayLen))
}

func getAssetInterface(connection *datacatalog.GetAssetResponse) taxonomy.Interface {
	if connection == nil || connection.Details.Connection.Name == "" {
		return taxonomy.Interface{Protocol: appApi.S3, DataFormat: ""}
	}
	return taxonomy.Interface{Protocol: connection.Details.Connection.Name, DataFormat: connection.Details.DataFormat}
}

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

	// an empty DataFormat value is not checked
	// either a module supports any format, or any format can be selected (no requirements)
	return moduleIntfc.DataFormat == "" || moduleIntfc.DataFormat == otherIntfc.DataFormat
}
