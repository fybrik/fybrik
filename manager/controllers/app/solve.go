// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"errors"

	"github.com/rs/zerolog"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/infrastructure"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
)

// DataInfo defines all the information about the given data set that comes from the fybrikapplication spec and from the connectors.
type DataInfo struct {
	// Source connection details
	DataDetails *datacatalog.GetAssetResponse
	// Pointer to the relevant data context in the Fybrik application spec
	Context *app.DataContext
	// Evaluated config policies
	Configuration adminconfig.EvaluatorOutput
	// Workload cluster
	WorkloadCluster multicluster.Cluster
	// Required governance actions to perform on this asset
	Actions []taxonomy.Action
	// Potential actions to be taken on storing this asset in a specific location
	StorageRequirements map[taxonomy.ProcessingLocation][]taxonomy.Action
}

// Environment defines the available resources (clusters, modules, storageAccounts)
// It also contains the results of queries to policy manager regarding writing data to storage accounts
type Environment struct {
	Modules          map[string]*app.FybrikModule
	Clusters         []multicluster.Cluster
	StorageAccounts  []*app.FybrikStorageAccount
	AttributeManager *infrastructure.AttributeManager
}

// Node represents an access point to data (as a physical source/sink, or a virtual endpoint)
// A virtual endpoint is activated by the workload for read/write actions.
type Node struct {
	Connection *taxonomy.Interface
	Virtual    bool
}

// Edge represents a module capability that gets data via source and returns data via sink interface
type Edge struct {
	Source          *Node
	Sink            *Node
	Module          *app.FybrikModule
	CapabilityIndex int
}

// ResolvedEdge extends an Edge by adding actions that a module should perform, and the cluster where the module will be deployed
// TODO(shlomitk1): add plugins/transformation capabilities to this structure
type ResolvedEdge struct {
	Edge
	Actions        []taxonomy.Action
	Cluster        string
	StorageAccount app.FybrikStorageAccountSpec
}

// Solution is a final solution enabling a plotter construction.
// It represents a full data flow between the data source and the workload.
type Solution struct {
	DataPath []*ResolvedEdge
}

// find a solution for a data path
// satisfying governance and admin policies
// with respect to the optimization strategy
func solve(env *Environment, datasetInfo *DataInfo, log *zerolog.Logger) (Solution, error) {
	if utils.UseCSP() {
		return Solution{}, errors.New("CSP solution is not yet implemented")
	}
	pathBuilder := PathBuilder{Log: log, Env: env, Asset: datasetInfo}
	return pathBuilder.solve()
}
