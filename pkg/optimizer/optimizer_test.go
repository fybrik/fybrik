// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"os"
	"testing"
)

func TestOptimizer(t *testing.T) {
	env := getTestEnv()
	opt := NewOptimizer(env, getDataInfo(env), os.Getenv("ABSTOOLBIN")+"/fzn-or-tools")
	solution, err := opt.Solve()
	if err != nil {
		t.Fatalf("Failed solving constraint problem: %s", err)
	}

	solutionLen := len(solution.DataPath)
	if solutionLen < 1 {
		t.Error("Solution is too short")
	} else if solutionLen > 3 {
		t.Errorf("Solution is too long: %d", solutionLen)
	}
	for _, edge := range solution.DataPath {
		t.Log(edge)
	}
}
