// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"errors"

	"fybrik.io/fybrik/manager/controllers/utils"
)

// find a solution for a data path
// satisfying governance and admin policies
// with respect to the optimization strategy
func solve(env *Environment, datasetInfo *DataInfo) (Solution, error) {
	if utils.UseCSP() {
		return Solution{}, errors.New("CSP solution is not yet implemented")
	}
	return env.solve(datasetInfo)
}
