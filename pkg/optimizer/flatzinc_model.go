// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

/*
This file implements FlatZincModel: a class to hold a constraint problem, based on the FlatZinc format.
The class can dump the a constraint program to a file, using the FlatZinc specification.
The class can also read solver solutions, written using the FlatZinc specification.
The FlatZinc specification: https://www.minizinc.org/doc-latest/en/fzn-spec.html
*/

// Data for a single FlatZinc parameter
type FlatZincParam struct {
	NumInstances uint // NumInstances > 1 means an array of params
	Name         string
	Type         string
	Assignment   string
}

// formats a paramater declaraion in FlatZinc format
func (fzp FlatZincParam) paramDeclaration() string {
	if fzp.NumInstances == 1 {
		return fmt.Sprintf("%s: %s = %s;\n", fzp.Type, fzp.Name, fzp.Assignment)
	}
	return fmt.Sprintf("array [1..%d] of %s: %s = %s;\n", fzp.NumInstances, fzp.Type, fzp.Name, fzp.Assignment)
}

type Annotations []string

func (annots Annotations) annotationString() string {
	annotsStr := strings.Join(annots, " :: ")
	if len(annotsStr) > 0 {
		annotsStr = " :: " + annotsStr
	}
	return annotsStr
}

// Data for a single FlatZinc variable
type FlatZincVariable struct {
	NumInstances uint // NumInstances > 1 means an array of vars
	Name         string
	Type         string
	Assignment   string
	Annotations
}

// formats a variable declaraion in FlatZinc format
func (fzv FlatZincVariable) varDeclaration() string {
	if fzv.NumInstances == 1 {
		return fmt.Sprintf("var %s: %s%s;\n", fzv.Type, fzv.Name, fzv.annotationString())
	}
	return fmt.Sprintf("array [1..%d] of var %s: %s%s;\n", fzv.NumInstances, fzv.Type, fzv.Name, fzv.annotationString())
}

// Data for a single FlatZinc constraint
type FlatZincConstraint struct {
	Identifier  string
	Expressions []string
	Annotations
}

// formats a constraint statement in FlatZinc format
func (cnstr FlatZincConstraint) constraintStatement() string {
	exprs := strings.Join(cnstr.Expressions, ", ")
	return fmt.Sprintf("constraint %s(%s)%s;\n", cnstr.Identifier, exprs, cnstr.annotationString())
}

// FlatZinc solve goal must be one of three types: satisfy, minimize, maximize
type SolveGoal int64

const (
	Satisfy SolveGoal = iota
	Minimize
	Maximize
)

func (s SolveGoal) String() string {
	switch s {
	case Satisfy:
		return "satisfy"
	case Minimize:
		return "minimize"
	case Maximize:
		return "maximize"
	}
	return "unknown"
}

// Data for a FlatZinc-model solve item
type FlatZincSolveItem struct {
	goal SolveGoal
	expr string
	Annotations
}

// formats a solve item in FlatZinc format
func (slv FlatZincSolveItem) solveItemStatement() string {
	return fmt.Sprintf("solve%s %s %s;\n", slv.annotationString(), slv.goal, slv.expr)
}

// The main class for holding a FlatZinc constraint problem
type FlatZincModel struct {
	ParamMap    map[string]FlatZincParam
	VarMap      map[string]FlatZincVariable
	Constraints []FlatZincConstraint
	SolveTarget FlatZincSolveItem
}

func NewFlatZincModel() *FlatZincModel {
	var fzw FlatZincModel
	fzw.ParamMap = make(map[string]FlatZincParam)
	fzw.VarMap = make(map[string]FlatZincVariable)
	fzw.Constraints = make([]FlatZincConstraint, 0)
	return &fzw
}

func (fzw *FlatZincModel) AddParam(numInst uint, name, vartype, assignment string) {
	fzw.ParamMap[name] = FlatZincParam{numInst, name, vartype, assignment}
}

func (fzw *FlatZincModel) AddVariable(numInst uint, name, vartype, assignment string, annotations ...string) {
	fzw.VarMap[name] = FlatZincVariable{numInst, name, vartype, assignment, annotations}
}

func (fzw *FlatZincModel) AddConstraint(identifier string, exprs []string, annotations ...string) {
	fzw.Constraints = append(fzw.Constraints, FlatZincConstraint{identifier, exprs, annotations})
}

func (fzw *FlatZincModel) SetSolveTarget(goal SolveGoal, expr string, annotations ...string) {
	fzw.SolveTarget = FlatZincSolveItem{goal, expr, annotations}
}

// dumps a FlatZinc model to a file, using the FlatZinc syntax
func (fzw *FlatZincModel) Dump(fileName string) error {
	file, err := os.Create(path.Clean(fileName))
	if err != nil {
		return fmt.Errorf("failed opening file %s for writing", fileName)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %s\n", err)
		}
	}()

	fileContent := ""
	for _, fzparam := range fzw.ParamMap {
		fileContent += fzparam.paramDeclaration()
	}

	fileContent += "\n"
	for _, fzvar := range fzw.VarMap {
		fileContent += fzvar.varDeclaration()
	}

	fileContent += "\n"
	for _, cnstr := range fzw.Constraints {
		fileContent += cnstr.constraintStatement()
	}

	fileContent += "\n" + fzw.SolveTarget.solveItemStatement()
	if _, err := file.WriteString(fileContent); err != nil {
		return err
	}
	return nil
}

// Parses a single variable assignment line in a FlatZinc solution file. Returns the variable name and its value(s)
func parseSolutionLine(line string, lineNum uint) (string, []string, error) {
	err := fmt.Errorf("parse error on line %d: %s", lineNum, line)
	lineParts := strings.Split(line, "=")
	if len(lineParts) != 2 {
		return "", nil, err
	}
	varName := lineParts[0]
	value := strings.TrimSuffix(lineParts[1], ";")
	values := []string{value}
	if strings.HasPrefix(value, "array") {
		leftBracketPos := strings.Index(value, "[")
		rightBracketPos := strings.Index(value, "]")
		if leftBracketPos == -1 || rightBracketPos == -1 || leftBracketPos > rightBracketPos {
			return "", nil, err
		}
		values = strings.Split(value[leftBracketPos+1:rightBracketPos], ",")
	}
	return varName, values, nil
}

// Represents a solution to the constraints problem - a map from variable names to their value(s) in the solution
type CPSolution map[string][]string

// Reading a FlatZinc solutions file and returning all solutions as a slice of CPSolution
func (fzw *FlatZincModel) ReadSolutions(fileName string) ([]CPSolution, error) {
	data, err := os.ReadFile(path.Clean(fileName))
	if err != nil {
		return nil, err
	}
	strContent := string(data)
	lines := strings.Split(strContent, "\n")
	res := []CPSolution{}
	currSolution := make(CPSolution)
	for lineNum, line := range lines {
		line = strings.Join(strings.Fields(line), "") // remove all whitespaces
		switch {
		case len(line) == 0:
			continue // empty line
		case strings.HasPrefix(line, "%%%"):
			continue // stat lines are ignored
		case line == strings.Repeat("=", 10):
			return res, nil // at least one  solution was found and the whole search space was covered
		case line == "=====UNSATISFIABLE=====":
			return []CPSolution{}, nil // no solution exists; returns an empty slice of solutions
		case strings.HasPrefix(line, "===="):
			err := fmt.Errorf("no solution found: %s", line) // no solution found (but not UNSAT either)
			return nil, err
		case line == strings.Repeat("-", 10):
			res = append(res, currSolution) // marks the end of current solution
			currSolution = make(CPSolution) // (and possible the beginning of a new one)
		default: // this should be a variable assignment line
			varName, values, err := parseSolutionLine(line, uint(lineNum))
			if err != nil {
				return nil, err
			}
			currSolution[varName] = values
		}
	}

	return res, nil
}

// Reading a FlatZinc solutions file and returning the best solution
// When a minimize/maximize goal is defined, best solution should be the last solution
func (fzw *FlatZincModel) ReadBestSolution(fileName string) (CPSolution, error) {
	solutions, error := fzw.ReadSolutions(fileName)
	if error != nil {
		return nil, error
	}
	return solutions[len(solutions)-1], nil
}
