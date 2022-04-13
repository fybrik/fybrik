// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"errors"
	"fmt"
	"os/exec"

	"fybrik.io/fybrik/manager/controllers/app"
)

const (
	MaxDataPathDepth = 4
)

type Optimizer struct {
	dpc         *DataPathCSP
	problemData *app.DataInfo
	env         *app.Environment
	solverPath  string
}

func NewOptimizer(env *app.Environment, problemData *app.DataInfo, solverPath string) *Optimizer {
	opt := Optimizer{dpc: NewDataPathCSP(problemData, env), problemData: problemData, env: env, solverPath: solverPath}
	return &opt
}

func (opt *Optimizer) Solve() (app.Solution, error) {
	for pathLen := 1; pathLen < MaxDataPathDepth; pathLen++ {
		fzModelFile := fmt.Sprintf("DataPathCSP%s_%d.fzn", opt.problemData.Context.DataSetID, pathLen)
		err := opt.dpc.BuildFzModel(fzModelFile, uint(pathLen))
		if err != nil {
			return app.Solution{}, err
		}
		// #nosec G204 -- Avoid "Subprocess launched with variable" error
		solverSolution, err := exec.Command(opt.solverPath, fzModelFile).Output()
		if err != nil {
			return app.Solution{}, err
		}
		solution, err := opt.dpc.decodeSolverSolution(string(solverSolution), pathLen)
		if err != nil {
			return app.Solution{}, err
		}
		if len(solution.DataPath) > 0 {
			return solution, nil
		}
	}
	msg := "Data path cannot be constructed given the deployed modules and the active restrictions"
	return app.Solution{}, errors.New(msg + " for " + opt.problemData.Context.DataSetID)
}
