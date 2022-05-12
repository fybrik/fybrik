// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"os"
	"reflect"
	"testing"
)

func TestWriteModel(t *testing.T) {
	myWriter := NewFlatZincModel()
	myWriter.AddHeaderComment("hello")
	myWriter.AddHeaderComment("this is a test")
	myWriter.AddParam("pi", "float", "3.1415")
	myWriter.AddParamArray("fib", "int", 7, "[1, 1, 2, 3, 5, 8, 13]")
	myWriter.AddVariable("y", "int", false, false)
	myWriter.AddVariableArray("y3", "int", 3, false, true)
	myWriter.AddConstraint("int_le", []string{"0", "x"}, "domain")
	myWriter.SetSolveTarget(Minimize, "x", "int_search(xs, input_order, indomain_min, complete)")
	fileName, err := myWriter.Dump()
	if fileName != "" {
		os.Remove(fileName)
	}
	if err != nil {
		t.Errorf("Failed writing FlatZinc file: %s ", err)
	}
}

var test1SolutionExpected = []CPSolution{
	{
		"Beamtime":   {"8"},
		"K":          {"2"},
		"ladder_num": {"1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1"},
		"x": {"23", "15", "18", "11", "6", "15", "18", "11", "6", "15", "18", "5", "16", "15", "18", "5", "16", "1", "18", "5", "16", "1",
			"18", "11", "16", "5", "18", "11", "16", "5", "1", "11", "16", "5", "1", "18", "6", "5", "1", "18", "6", "5", "1", "20", "6", "12", "1",
			"20", "6", "12", "1", "25", "16", "12", "1", "25"},
	},
	{
		"Beamtime":   {"21"},
		"K":          {"7"},
		"ladder_num": {"14", "6", "5", "13", "7", "8", "11", "9", "10", "1", "2", "3", "4", "12"},
		"x": {"23", "15", "18", "11", "6", "15", "18", "11", "6", "15", "18", "5", "16", "15", "18", "5", "16", "1", "18", "5", "16", "1",
			"18", "11", "16", "5", "18", "11", "16", "5", "1", "11", "16", "5", "1", "18", "6", "5", "1", "18", "6", "5", "1", "20", "6", "12", "1",
			"20", "6", "12", "1", "25", "16", "12", "1", "25"},
	},
}

func TestReadingResults(t *testing.T) {
	myReader := NewFlatZincModel()
	res, err := myReader.ReadSolutionsFromFile("testdata/test1.fzn_solution")
	if err != nil {
		t.Errorf("%s", err)
	}
	if !reflect.DeepEqual(res, test1SolutionExpected) {
		t.Errorf("Unexpected result.\nExpected: %v\nActual: %v", test1SolutionExpected, res)
	}
}

func TestReadingBestResults(t *testing.T) {
	myReader := NewFlatZincModel()
	res, err := myReader.ReadBestSolutionFromFile("testdata/test1.fzn_solution")
	if err != nil {
		t.Errorf("%s", err)
	}
	expected := test1SolutionExpected[len(test1SolutionExpected)-1]
	if !reflect.DeepEqual(res, expected) {
		t.Errorf("Unexpected result.\nExpected: %v\nActual: %v", expected, res)
	}
}

func TestReadingUNSATResults(t *testing.T) {
	myReader := NewFlatZincModel()
	res, err := myReader.ReadSolutionsFromFile("testdata/unsat.fzn_solution")
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(res) != 1 || len(res[0]) > 0 {
		t.Errorf("Expecting a single empty solution")
	}
}

func TestReadingBestUNSATResults(t *testing.T) {
	myReader := NewFlatZincModel()
	res, err := myReader.ReadBestSolutionFromFile("testdata/unsat.fzn_solution")
	if err != nil {
		t.Errorf("%s", err)
	}
	if len(res) > 0 {
		t.Errorf("Expecting a single empty solution")
	}
}

func TestReadingUnknownResults(t *testing.T) {
	myReader := NewFlatZincModel()
	_, err := myReader.ReadSolutionsFromFile("testdata/unknown.fzn_solution")
	if err == nil {
		t.Errorf("Expecting an error when result is unknown")
	}
}

func TestReadingBadResults(t *testing.T) {
	myReader := NewFlatZincModel()
	_, err := myReader.ReadSolutionsFromFile("testdata/bad.fzn_solution")
	if err == nil {
		t.Errorf("Expected a parse error on an ill-formatted file")
	}
}
