// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datapath

import (
	"fmt"

	fappv1 "fybrik.io/fybrik/manager/apis/app/v1beta1"
	fappv2 "fybrik.io/fybrik/manager/apis/app/v1beta2"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

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
	Module          *fappv1.FybrikModule
	CapabilityIndex int
}

// ResolvedEdge extends an Edge by adding actions that a module should perform, and the cluster where the module will be deployed
// TODO(shlomitk1): add plugins/transformation capabilities to this structure
type ResolvedEdge struct {
	Edge
	Actions        []taxonomy.Action
	Cluster        string
	StorageAccount fappv2.FybrikStorageAccountSpec
}

// Solution is a final solution enabling a plotter construction.
// It represents a full data flow between the data source and the workload.
type Solution struct {
	DataPath []*ResolvedEdge
}

func (re *ResolvedEdge) String() string {
	return fmt.Sprintf("Source: %v, Sink: %v, Module:%v, CapIndex: %v, Actions: %v, Cluster: %v, SA: %v",
		re.Source, re.Sink, re.Module.Name, re.CapabilityIndex, re.Actions, re.Cluster, re.StorageAccount)
}

func (solution *Solution) Reverse() {
	for elementInd := 0; elementInd < len(solution.DataPath)/2; elementInd++ {
		reversedInd := len(solution.DataPath) - elementInd - 1
		solution.DataPath[elementInd], solution.DataPath[reversedInd] =
			solution.DataPath[reversedInd], solution.DataPath[elementInd]
	}
}
