// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"fmt"
	"strconv"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/multicluster"
)

// DataInfo defines all the information about the given data set that comes from the fybrikapplication spec and from the connectors.
type DataInfo struct {
	// All available modules
	Modules map[string]*app.FybrikModule
	// All available clusters
	Clusters []multicluster.Cluster
	// Source connection details
	//DataDetails *datacatalog.GetAssetResponse
	// Pointer to the relevant data context in the Fybrik application spec
	//Context *app.DataContext
	// Evaluated config policies
	//Configuration adminconfig.EvaluatorOutput
	// Workload cluster
	//WorkloadCluster multicluster.Cluster
	// Governance actions to perform on this asset
	//Actions []taxonomy.Action
}

type DataPathCSP struct {
	problemData *DataInfo
	fzModel     *FlatZincModel
}

func NewDataPathCSP(problemData *DataInfo) *DataPathCSP {
	dpCSP := DataPathCSP{problemData, NewFlatZincModel()}
	return &dpCSP
}

func (dpc *DataPathCSP) BuildFzModel(pathLength uint) error {
	arrayAnnotation := fmt.Sprintf("output_array([1..%d])", pathLength)
	// Variables to select the module we place on each data-path location
	moduleTypeVarType := rangeVarType(len(dpc.problemData.Modules) - 1)
	dpc.fzModel.AddVariable(pathLength, "moduleType", moduleTypeVarType, "", arrayAnnotation)
	// Variables to select the cluster we allocate to each module on the path
	moduleClusterVarType := rangeVarType(len(dpc.problemData.Clusters) - 1)
	dpc.fzModel.AddVariable(pathLength, "moduleCluster", moduleClusterVarType, "", arrayAnnotation)

	err := dpc.fzModel.Dump("dataPath.fzn")
	return err
}

// ----- helper functions -----

func rangeVarType(rangeEnd int) string {
	return "0.." + strconv.Itoa(rangeEnd)
}
