// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"fmt"
	"testing"
)

func TestWriteModel(t *testing.T) {
	myWriter := NewFlatZincModel()
	myWriter.AddParam(1, "pi", "float", "3.1415")
	myWriter.AddParam(7, "fib", "int", "[1, 1, 2, 3, 5, 8, 13]")
	myWriter.AddVariable(1, "y", "int", "")
	myWriter.AddVariable(3, "y3", "int", "", "mip2", "mip3")
	myWriter.AddConstraint("int_le", []string{"0", "x"}, "domain")
	myWriter.SetSolveTarget(Minimize, "x", "int_search(xs, input_order, indomain_min, complete)")
	err := myWriter.Dump("test.fzn")
	if err != nil {
		t.Errorf("Failed writing FlatZinc file: %s ", err)
	}
}

func TestReadingResults(t *testing.T) {
	myReader := NewFlatZincModel()
	res, err := myReader.ReadSolutions("test1.fzn_solution")
	if err != nil {
		t.Errorf("%s", err)
	}
	fmt.Println(res)
}

func TestReadingUNSATResults(t *testing.T) {
	myReader := NewFlatZincModel()
	_, err := myReader.ReadSolutions("unsat.fzn_solution")
	if err == nil {
		t.Errorf("Expected an error on this test")
	}
}

func TestReadingBadResults(t *testing.T) {
	myReader := NewFlatZincModel()
	_, err := myReader.ReadSolutions("bad.fzn_solution")
	if err == nil {
		t.Errorf("Expected an error on this test")
	}
}
