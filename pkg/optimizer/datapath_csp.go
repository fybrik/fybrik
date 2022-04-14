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

// Names of the primary variables on which we need to take decisions
// Each variable is an array of ints. The i-th cell in the array represents a decision for the i-th Node/Edge in the data path
const (
	modCapVarname    = "moduleCapability"      // Var's value says which capability of which module to use (on each Edge)
	clusterVarname   = "moduleCluster"         // Var's value says which of the available clusters to use (on each Edge)
	saVarname        = "storageAccount"        // Var's value says which storage account to use (0 means no sa)
	srcIntfcVarname  = "moduleSourceInterface" // Var's value says which interface to use as source
	sinkIntfcVarname = "moduleSinkInterface"   // Var's value says which interface to use as sink
)

// Couples together a module and one of its capabilities
type moduleAndCapability struct {
	module        *appApi.FybrikModule
	capability    *appApi.ModuleCapability
	capabilityIdx int // The index of capability in module's spec
}

// The main class for producing a CSP from data-path constraints and for decoding solver's solutions
type DataPathCSP struct {
	problemData         *app.DataInfo
	env                 *app.Environment
	modulesCapabilities []moduleAndCapability       // An enumeration of allowed capabilities in all modules
	interfaceIdx        map[taxonomy.Interface]int  // gives an index for each unique interface
	reverseIntfcMap     map[int]*taxonomy.Interface // The reverse mapping (needed when decoding the solution)
	indicators          map[string]bool             // indicator vars used as part of the problem (to prevent redefinition)
	fzModel             *FlatZincModel
}

// The ctor also enumerates all available (module x capabilities) and all available interfaces
// The generated enumerations are listed at the header of the FlatZinc model
func NewDataPathCSP(problemData *app.DataInfo, env *app.Environment) *DataPathCSP {
	dpCSP := DataPathCSP{problemData: problemData, env: env, fzModel: NewFlatZincModel()}
	dpCSP.interfaceIdx = map[taxonomy.Interface]int{}
	dpCSP.reverseIntfcMap = map[int]*taxonomy.Interface{}
	dataSetIntfc := taxonomy.Interface{
		Protocol:   dpCSP.problemData.DataDetails.Details.Connection.Name,
		DataFormat: dpCSP.problemData.DataDetails.Details.DataFormat,
	}
	dpCSP.addInterfaceToMaps(&dataSetIntfc)
	dpCSP.addInterfaceToMaps(&dpCSP.problemData.Context.Requirements.Interface)

	dpCSP.fzModel.AddHeaderComment("Encoding of modules and their capabilities:")
	comment := ""
	for _, module := range env.Modules {
		for idx, capability := range module.Spec.Capabilities {
			modCap := moduleAndCapability{module, &module.Spec.Capabilities[idx], idx}
			if dpCSP.moduleCapabilityAllowedByRestrictions(modCap) {
				dpCSP.modulesCapabilities = append(dpCSP.modulesCapabilities, modCap)
				dpCSP.addModCapInterfacesToMaps(modCap)
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
	return &dpCSP
}

// Add the interfaces defined in a given module's capability to the 2 interface maps
func (dpc *DataPathCSP) addModCapInterfacesToMaps(modcap moduleAndCapability) {
	for _, intfc := range modcap.capability.SupportedInterfaces {
		dpc.addInterfaceToMaps(intfc.Source)
		dpc.addInterfaceToMaps(intfc.Sink)
	}
}

// Add the given interface to the 2 interface maps (but avoid duplicates)
func (dpc *DataPathCSP) addInterfaceToMaps(intfc *taxonomy.Interface) {
	_, found := dpc.interfaceIdx[*intfc]
	if !found {
		intfcIdx := len(dpc.interfaceIdx) + 1
		dpc.interfaceIdx[*intfc] = intfcIdx
		dpc.reverseIntfcMap[intfcIdx] = intfc
	}
}

// This is the main method for building a FlatZinc CSP out of the data-path parameters and constraints.
// NOTE: Minimal index of FlatZinc arrays is always 1. Hence, we use 1-based modeling all over the place to avoid confusion
//       The one exception is storage accounts, where a value of 0 means no storage account
func (dpc *DataPathCSP) BuildFzModel(fzModelFile string, pathLength uint) error {
	dpc.fzModel.Clear() // This function can be called multiple times - clear vars and constraints from last call
	// Variables to select the module capability we use on each data-path location
	moduleCapabilityVarType := rangeVarType(1, len(dpc.modulesCapabilities))
	dpc.fzModel.AddVariableArray(modCapVarname, moduleCapabilityVarType, pathLength, false, true)
	// Variables to select storage-accounts to place on each data-path location (the value 0 means no storage account)
	saTypeVarType := rangeVarType(0, len(dpc.env.StorageAccounts))
	dpc.fzModel.AddVariableArray(saVarname, saTypeVarType, pathLength, false, true)
	// Variables to select the cluster we allocate to each module on the path
	moduleClusterVarType := rangeVarType(1, len(dpc.env.Clusters))
	dpc.fzModel.AddVariableArray(clusterVarname, moduleClusterVarType, pathLength, false, true)
	// Variables to select the source and sink interface for each module on the path
	moduleInterfaceVarType := rangeVarType(1, len(dpc.interfaceIdx))
	dpc.fzModel.AddVariableArray(srcIntfcVarname, moduleInterfaceVarType, pathLength, false, true)
	dpc.fzModel.AddVariableArray(sinkIntfcVarname, moduleInterfaceVarType, pathLength, false, true)

	dpc.addGovernanceActionConstraints(pathLength)
	dpc.addAdminConfigRestrictions(int(pathLength))
	dpc.addInterfaceConstraints(pathLength)

	err := dpc.fzModel.Dump(fzModelFile)
	return err
}

// enforce restrictions from admin configuration decisions:
// a. cluster satisfies restrictions for the selected capability
// b. storage account satisfies restrictions for the selected capability
func (dpc *DataPathCSP) addAdminConfigRestrictions(pathLength int) {
	for decCapability := range dpc.problemData.Configuration.ConfigDecisions {
		for modCapIdx, moduleCap := range dpc.modulesCapabilities {
			if moduleCap.capability.Capability != decCapability {
				continue
			}
			decision := dpc.problemData.Configuration.ConfigDecisions[decCapability]
			for clusterIdx, cluster := range dpc.env.Clusters {
				if !dpc.clusterSatisfiesRestrictions(cluster, decision.DeploymentRestrictions.Clusters) {
					dpc.preventAssignments([]string{modCapVarname, clusterVarname},
						[]int{modCapIdx + 1, clusterIdx + 1}, pathLength)
				}
			}
			for saIdx := range dpc.env.StorageAccounts {
				if !dpc.saSatisfiesRestrictions(&dpc.env.StorageAccounts[saIdx], decision.DeploymentRestrictions.StorageAccounts) {
					dpc.preventAssignments([]string{modCapVarname, saVarname},
						[]int{modCapIdx + 1, saIdx + 1}, pathLength)
				}
			}
		}
	}
}

// Decide if a given module and its given capability satisfy all administrator's restrictions
func (dpc *DataPathCSP) moduleCapabilityAllowedByRestrictions(modcap moduleAndCapability) bool {
	decision := dpc.problemData.Configuration.ConfigDecisions[modcap.capability.Capability]
	if decision.Deploy == adminconfig.StatusFalse {
		return false // this type of capability should never be deployed
	}
	return dpc.moduleSatisfiesRestrictions(modcap.module, decision.DeploymentRestrictions.Modules)
}

// Decide if a given module satisfies all administrator's restrictions
func (dpc *DataPathCSP) moduleSatisfiesRestrictions(module *appApi.FybrikModule, restrictions []adminconfig.Restriction) bool {
	for _, restriction := range restrictions {
		if !restriction.SatisfiedByResource(dpc.env.AttributeManager, module.Spec, "") {
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
	for pos := 1; pos <= pathLength; pos++ {
		indexedVars := []string{}
		for _, v := range variables {
			indexedVars = append(indexedVars, varAtPos(v, pos))
		}
		dpc.preventAssignment(indexedVars, values)
	}
}

// Adds constraints to prevent the joint assignment of `values` to `variables`
func (dpc *DataPathCSP) preventAssignment(variables []string, values []int) {
	indicators := []string{}
	for idx, variable := range variables {
		indicators = append(indicators, dpc.addInequalityIndicator(variable, values[idx]))
	}
	indicatorsArray := fznCompoundLiteral(indicators, false)
	dpc.fzModel.AddConstraint(ArrBoolOrConstraint, []string{indicatorsArray, "true"})
}

// Adds an indicator variable which is true iff a given integer variable DOES NOT EQUAL a given value
func (dpc *DataPathCSP) addInequalityIndicator(variable string, value int) string {
	indicator := fmt.Sprintf("ind_%s_ne_%d", variable, value)
	indicator = strings.ReplaceAll(indicator, "[", "")
	indicator = strings.ReplaceAll(indicator, "]", "")
	if _, defined := dpc.indicators[indicator]; !defined {
		dpc.fzModel.AddVariable(indicator, BoolType, true, false)
		annotation := fmt.Sprintf("defines_var(%s)", indicator)
		dpc.fzModel.AddConstraint(IntNotEqConstraint, []string{variable, strconv.Itoa(value), indicator}, annotation)
	}
	return indicator
}

// Make sure that every required governance action is implemented exactly one time.
func (dpc *DataPathCSP) addGovernanceActionConstraints(pathLength uint) {
	repeatingOnes := strings.Repeat("1 ", int(pathLength))
	allOnesArrayLiteral := fznCompoundLiteral(strings.Fields(repeatingOnes), false)

	for _, action := range dpc.problemData.Actions {
		// An *output* array of Booleans variable to mark whether the current action is applied at location i
		actionVar := getActionVarname(action)
		dpc.fzModel.AddVariableArray(actionVar, BoolType, pathLength, false, true)
		// ensuring action is implemented once
		dpc.fzModel.AddConstraint(BoolLinEqConstraint, []string{allOnesArrayLiteral, actionVar, strconv.Itoa(1)})

		// accumulate module-capabilities that support the current action
		moduleCapabilitiesStrs := []string{}
		for modCapIdx, modCap := range dpc.modulesCapabilities {
			for _, capAction := range modCap.capability.Actions {
				if capAction.Name == action.Name {
					moduleCapabilitiesStrs = append(moduleCapabilitiesStrs, strconv.Itoa(modCapIdx+1))
				}
			}
		}
		modCapsSupportingAction := fznCompoundLiteral(moduleCapabilitiesStrs, true)

		// add vars (and constraints) indicating if an action is supported at each path location
		for pathPos := 1; pathPos <= int(pathLength); pathPos++ {
			modCapAtPos := varAtPos(modCapVarname, pathPos)
			actSupportedVarname := fmt.Sprintf("action%sSupportedAt%d", action.Name, pathPos)
			dpc.fzModel.AddVariable(actSupportedVarname, BoolType, true, false)
			dpc.fzModel.AddConstraint(SetInConstraint, []string{modCapAtPos, modCapsSupportingAction, actSupportedVarname})
			dpc.fzModel.AddConstraint(BoolLeConstraint, []string{varAtPos(actionVar, pathPos), actSupportedVarname})
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
				dpc.preventAssignments([]string{modCapVarname, srcIntfcVarname}, []int{modCapIdx + 1, intfcIdx}, int(pathLength))
			}
			if !modcapSupportsIntfcSink {
				dpc.preventAssignments([]string{modCapVarname, sinkIntfcVarname}, []int{modCapIdx + 1, intfcIdx}, int(pathLength))
			}
		}
	}

	// ensuring interface matching along the datapath from dataset to workload
	dpc.fzModel.AddConstraint(IntEqConstraint, []string{varAtPos(srcIntfcVarname, 1), strconv.Itoa(1)})
	for pathPos := 1; pathPos < int(pathLength); pathPos++ {
		dpc.fzModel.AddConstraint(IntEqConstraint, []string{varAtPos(sinkIntfcVarname, pathPos), varAtPos(srcIntfcVarname, pathPos+1)})
	}
	dpc.fzModel.AddConstraint(IntEqConstraint, []string{varAtPos(sinkIntfcVarname, int(pathLength)),
		strconv.Itoa(dpc.interfaceIdx[dpc.problemData.Context.Requirements.Interface])})
}

// Returns which actions should be activated by the module at position pathPos
func (dpc *DataPathCSP) getSolutionActionsAtPos(solverSolution CPSolution, pathPos int) []taxonomy.Action {
	actions := []taxonomy.Action{}
	for _, action := range dpc.problemData.Actions {
		actionVarname := getActionVarname(action)
		actionSolution := solverSolution[actionVarname]
		if actionSolution[pathPos] == "true" {
			actions = append(actions, action)
		}
	}
	return actions
}

// Translates a solver's solution into a FybrikApplication Solution for a given data-path
// TODO: better handle error messages
func (dpc *DataPathCSP) decodeSolverSolution(solverSolutionStr string, pathLen int) (app.Solution, error) {
	solverSolution, err := dpc.fzModel.ReadBestSolution(solverSolutionStr)
	if err != nil {
		return app.Solution{}, err
	}
	if len(solverSolution) == 0 {
		return app.Solution{}, nil // UNSAT
	}

	modCapSolution := solverSolution[modCapVarname]
	clusterSolution := solverSolution[clusterVarname]
	saSolution := solverSolution[saVarname]
	srcIntfcSolution := solverSolution[srcIntfcVarname]
	sinkIntfcSolution := solverSolution[sinkIntfcVarname]

	srcIntfcIdx, _ := strconv.Atoi(srcIntfcSolution[0])
	srcNode := &app.Node{Connection: dpc.reverseIntfcMap[srcIntfcIdx]}

	solution := app.Solution{}
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
		sinkNode := &app.Node{Connection: dpc.reverseIntfcMap[sinkIntfcIdx]}
		edge := app.Edge{Module: modCap.module, CapabilityIndex: modCap.capabilityIdx, Source: srcNode, Sink: sinkNode}
		resolvedEdge := app.ResolvedEdge{
			Edge:           edge,
			Actions:        dpc.getSolutionActionsAtPos(solverSolution, pathPos),
			Cluster:        dpc.env.Clusters[clusterIdx-1].Name,
			StorageAccount: sa,
		}
		solution.DataPath = append(solution.DataPath, &resolvedEdge)
		srcNode = sinkNode
	}

	return solution, nil
}

// ----- helper functions -----

func getActionVarname(action taxonomy.Action) string {
	return fmt.Sprintf("action%s", action.Name)
}

func rangeVarType(rangeStart, rangeEnd int) string {
	if rangeEnd < rangeStart {
		rangeEnd = rangeStart
	}
	return fmt.Sprintf("%d..%d", rangeStart, rangeEnd)
}

func varAtPos(variable string, pos int) string {
	return fmt.Sprintf("%s[%d]", variable, pos)
}

func fznCompoundLiteral(values []string, isSet bool) string {
	jointValues := strings.Join(values, ", ")
	if isSet {
		return fmt.Sprintf("{%s}", jointValues)
	}
	return fmt.Sprintf("[%s]", jointValues)
}

func interfacesMatch(intfc1, intfc2 taxonomy.Interface) bool {
	return intfc1.Protocol == intfc2.Protocol && intfc1.DataFormat == intfc2.DataFormat
}
