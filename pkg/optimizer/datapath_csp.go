// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"fmt"
	"strconv"

	"fybrik.io/fybrik/manager/controllers/app"
)

type DataPathCSP struct {
	problemData *app.DataInfo
	env         *app.Environment
	fzModel     *FlatZincModel
}

func NewDataPathCSP(problemData *app.DataInfo, env *app.Environment) *DataPathCSP {
	dpCSP := DataPathCSP{problemData, env, NewFlatZincModel()}
	return &dpCSP
}

func (dpc *DataPathCSP) BuildFzModel(pathLength uint) error {
	outputArrayAnnotation := fmt.Sprintf("output_array([1..%d])", pathLength)
	// Variables to select the module we place on each data-path location
	moduleTypeVarType := rangeVarType(len(dpc.env.Modules) - 1)
	dpc.fzModel.AddVariable(pathLength, "moduleType", moduleTypeVarType, "", outputArrayAnnotation)
	// Variables to select the cluster we allocate to each module on the path
	moduleClusterVarType := rangeVarType(len(dpc.env.Clusters) - 1)
	dpc.fzModel.AddVariable(pathLength, "moduleCluster", moduleClusterVarType, "", outputArrayAnnotation)

	dpc.addGovernanceActionConstraints(pathLength)

	err := dpc.fzModel.Dump("dataPath.fzn")
	return err
}

func (dpc *DataPathCSP) addGovernanceActionConstraints(pathLength uint) {
	// TODO: Make sure arrays of size one are handled properly
	if len(dpc.problemData.Actions) < 1 {
		return
	}

	// Variables to mark specific actions are applied
	dpc.fzModel.AddVariable(uint(len(dpc.problemData.Actions)), "actionHandled", "bool", "")
	dpc.fzModel.AddConstraint("array_bool_and", []string{"actionHandled", "true"}) // All actions must be handled

	// Variables to mark specific actions are applied at location i
	outputArrayAnnotation := fmt.Sprintf("output_array([1..%d])", pathLength)
	for actionIdx, action := range dpc.problemData.Actions {
		actionVar := fmt.Sprintf("%sAction", action.Name)
		dpc.fzModel.AddVariable(pathLength, actionVar, "bool", "", outputArrayAnnotation)
		actionPosInActionsArray := fmt.Sprintf("actionHandled[%d]", actionIdx+1)
		dpc.fzModel.AddConstraint("array_bool_or", []string{actionVar, actionPosInActionsArray})
	}

}

// ----- helper functions -----

func rangeVarType(rangeEnd int) string {
	return "0.." + strconv.Itoa(rangeEnd)
}
