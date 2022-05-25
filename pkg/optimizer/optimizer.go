// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

/*
	This package is for finding optimal data-path under constraints
	Its main Optimizer class takes data-path and infrastructure metadata, restrictions and optimization goals.
	Optimizer.Solve() returns a valid and optimal data path from a single DataSet to Workload (if such a path exists).
	Note that currently only a single dataset is considered in a given optimization problem.
	Also, more complex data-planes (e.g., DAG shaped) are not yet supported.

	All relevant data gets translated into a Constraint Satisfaction Problem (CSP) in the FlatZinc format
	(see https://www.minizinc.org/doc-latest/en/fzn-spec.html)
	Any FlatZinc-supporting CSP solver can then be called to get an optimal solution.
*/

package optimizer

import (
	"errors"
	"os"
	"os/exec"

	"github.com/rs/zerolog"

	"fybrik.io/fybrik/pkg/datapath"
	"fybrik.io/fybrik/pkg/logging"
)

const (
	MaxDataPathDepth = 4
)

type Optimizer struct {
	dpc         *DataPathCSP
	problemData *datapath.DataInfo
	env         *datapath.Environment
	solverPath  string
	log         *zerolog.Logger
}

func NewOptimizer(env *datapath.Environment, problemData *datapath.DataInfo, solverPath string, log *zerolog.Logger) *Optimizer {
	opt := Optimizer{dpc: NewDataPathCSP(problemData, env), problemData: problemData,
		env: env, solverPath: solverPath, log: log}
	return &opt
}

func (opt *Optimizer) getSolution(pathLength int) (string, error) {
	modelFile, err := opt.dpc.BuildFzModel(pathLength)
	if len(modelFile) > 0 {
		defer os.Remove(modelFile)
	}
	if err != nil {
		return "", err
	}

	// #nosec G204 -- Avoid "Subprocess launched with variable" error
	solverSolution, err := exec.Command(opt.solverPath, modelFile).Output()
	if err != nil {
		return "", err
	}
	return string(solverSolution), nil
}

// The main method to call for finding a legal and optimal data path
// Attempts short data-paths first, and gradually increases data-path length.
func (opt *Optimizer) Solve() (datapath.Solution, error) {
	for pathLen := 1; pathLen <= MaxDataPathDepth; pathLen++ {
		solverSolution, err := opt.getSolution(pathLen)
		if err != nil {
			return datapath.Solution{}, err
		}
		solution, err := opt.dpc.decodeSolverSolution(solverSolution, pathLen)
		if err != nil {
			return datapath.Solution{}, err
		}
		if len(solution.DataPath) > 0 {
			return solution, nil
		}
	}
	msg := "Data path cannot be constructed given the deployed modules and the active restrictions"
	opt.log.Error().Str(logging.DATASETID, opt.problemData.Context.DataSetID).Msg(msg)
	logging.LogStructure("Data Item Context", opt.problemData, opt.log, zerolog.TraceLevel, true, true)
	logging.LogStructure("Module Map", opt.env.Modules, opt.log, zerolog.TraceLevel, true, true)
	return datapath.Solution{}, errors.New(msg + " for " + opt.problemData.Context.DataSetID)
}
