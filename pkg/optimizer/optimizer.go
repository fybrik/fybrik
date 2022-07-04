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
	"math"
	"os"
	"os/exec"

	"emperror.dev/errors"

	"github.com/rs/zerolog"

	"fybrik.io/fybrik/pkg/datapath"
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
	opt.log.Debug().Msgf("finding solution of length %d", pathLength)
	modelFile, err := opt.dpc.BuildFzModel(pathLength)
	if len(modelFile) > 0 {
		defer os.Remove(modelFile)
	}
	if err != nil {
		return "", errors.Wrap(err, "error building a model")
	}

	opt.log.Debug().Msgf("Executing %s %s", opt.solverPath, modelFile)
	// #nosec G204 -- Avoid "Subprocess launched with variable" error
	solverSolution, err := exec.Command(opt.solverPath, modelFile).Output()
	if err != nil {
		return "", errors.Wrapf(err, "error executing %s %s", opt.solverPath, modelFile)
	}
	return string(solverSolution), nil
}

// The main method to call for finding a legal and optimal data path
// Attempts short data-paths first, and gradually increases data-path length.
func (opt *Optimizer) Solve() (datapath.Solution, error) {
	bestScore := math.NaN()
	bestSolution := datapath.Solution{}
	for pathLen := 1; pathLen <= MaxDataPathDepth; pathLen++ {
		solverSolution, err := opt.getSolution(pathLen)
		if err != nil {
			return datapath.Solution{}, err
		}
		solution, score, err := opt.dpc.decodeSolverSolution(solverSolution, pathLen)
		if err != nil {
			return datapath.Solution{}, err
		}
		if len(solution.DataPath) > 0 && math.IsNaN(score) { // no optimization goal is specified. prefer shorter paths
			return solution, nil
		}
		if !math.IsNaN(score) && (math.IsNaN(bestScore) || score < bestScore) {
			bestScore = score
			bestSolution = solution
		}
	}
	return bestSolution, nil
}
