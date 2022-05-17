// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"errors"
	"os"
	"os/exec"
)

const (
	MaxDataPathDepth = 4
)

// The main class for finding optimal data-path under constraints
// Takes data-path parameters, restrictions and optimization goals and returns a valid and optimal data path
// (if such a path exists)
// Translates all relevant data into a Constraint Satisfaction Problem (CSP) and calls a CSP solver to get an optimal solution
// Attempts short data-paths first, and gradually increases data-path length.
type Optimizer struct {
	dpc         *DataPathCSP
	problemData *DataInfo
	env         *Environment
	solverPath  string
}

func NewOptimizer(env *Environment, problemData *DataInfo, solverPath string) *Optimizer {
	opt := Optimizer{dpc: NewDataPathCSP(problemData, env), problemData: problemData, env: env, solverPath: solverPath}
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
func (opt *Optimizer) Solve() (Solution, error) {
	for pathLen := 1; pathLen <= MaxDataPathDepth; pathLen++ {
		solverSolution, err := opt.getSolution(pathLen)
		if err != nil {
			return Solution{}, err
		}
		solution, err := opt.dpc.decodeSolverSolution(solverSolution, pathLen)
		if err != nil {
			return Solution{}, err
		}
		if len(solution.DataPath) > 0 {
			return solution, nil
		}
	}
	msg := "Data path cannot be constructed given the deployed modules and the active restrictions"
	return Solution{}, errors.New(msg + " for " + opt.problemData.Context.DataSetID)
}
