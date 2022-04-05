// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strings"
)

/*
This file implements FlatZincModel: a class to hold a constraint problem, based on the FlatZinc format.
The class can dump the constraint program to a file, using the FlatZinc specification.
The class can also read solver solutions, written using the FlatZinc specification.
The FlatZinc specification: https://www.minizinc.org/doc-latest/en/fzn-spec.html
*/

type Declares interface {
	Declaration() string
}

// Data for a single FlatZinc parameter
type FlatZincParam struct {
	Name       string
	Type       string
	Size       uint
	IsArray    bool // (IsArray == false) implies (Size == 1)
	Assignment string
}

// formats a paramater declaraion in FlatZinc format
func (fzp FlatZincParam) Declaration() string {
	if fzp.IsArray {
		return fmt.Sprintf("array [1..%d] of %s: %s = %s;\n", fzp.Size, fzp.Type, fzp.Name, fzp.Assignment)
	}
	return fmt.Sprintf("%s: %s = %s;\n", fzp.Type, fzp.Name, fzp.Assignment)
}

type Annotations []string

func (annots Annotations) annotationString() string {
	// prepending an empty string, because annotations always start with "::"
	return strings.Join(append([]string{""}, annots...), " :: ")
}

// Data for a single FlatZinc variable
type FlatZincVariable struct {
	Name       string
	Type       string
	Size       uint
	IsArray    bool // (IsArray == false) implies (Size == 1)
	Assignment string
	Annotations
}

// formats a variable declaraion in FlatZinc format
func (fzv FlatZincVariable) Declaration() string {
	if fzv.IsArray {
		return fmt.Sprintf("array [1..%d] of var %s: %s%s;\n", fzv.Size, fzv.Type, fzv.Name, fzv.annotationString())
	}
	return fmt.Sprintf("var %s: %s%s;\n", fzv.Type, fzv.Name, fzv.annotationString())
}

// Data for a single FlatZinc constraint
type FlatZincConstraint struct {
	Identifier  string
	Expressions []string
	Annotations
}

// formats a constraint statement in FlatZinc format
func (cnstr *FlatZincConstraint) constraintStatement() string {
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
func (slv *FlatZincSolveItem) solveItemStatement() string {
	return fmt.Sprintf("solve%s %s %s;\n", slv.annotationString(), slv.goal, slv.expr)
}

// The main class for holding a FlatZinc constraint problem
type FlatZincModel struct {
	HeaderComments string
	ParamMap       map[string]Declares
	VarMap         map[string]Declares
	Constraints    []FlatZincConstraint
	SolveTarget    FlatZincSolveItem
}

func NewFlatZincModel() *FlatZincModel {
	var fzw FlatZincModel
	fzw.ParamMap = make(map[string]Declares)
	fzw.VarMap = make(map[string]Declares)
	return &fzw
}

func (fzw *FlatZincModel) AddHeaderComment(commentLine string) {
	fzw.HeaderComments = fzw.HeaderComments + "% " + commentLine + "\n"
}

func (fzw *FlatZincModel) AddParam(name, vartype, assignment string) {
	fzw.ParamMap[name] = FlatZincParam{Name: name, Type: vartype, Size: 1, IsArray: false, Assignment: assignment}
}

func (fzw *FlatZincModel) AddParamArray(name, vartype string, size uint, assignment string) {
	fzw.ParamMap[name] = FlatZincParam{Name: name, Type: vartype, Size: size, IsArray: true, Assignment: assignment}
}

func (fzw *FlatZincModel) AddVariable(name, vartype, assignment string, isDefined, isOutput bool) {
	annotations := []string{}
	if isDefined {
		annotations = append(annotations, "is_defined_var")
	}
	if isOutput {
		annotations = append(annotations, "output_var")
	}
	fzw.VarMap[name] = FlatZincVariable{
		Name: name, Type: vartype, Size: 1, IsArray: false,
		Assignment: assignment, Annotations: annotations,
	}
}

func (fzw *FlatZincModel) AddVariableArray(name, vartype string, size uint, assignment string, isDefined, isOutput bool) {
	annotations := []string{}
	if isDefined {
		annotations = append(annotations, "is_defined_var")
	}
	if isOutput {
		annotations = append(annotations, fmt.Sprintf("output_array([1..%d])", size))
	}
	fzw.VarMap[name] = FlatZincVariable{
		Name: name, Type: vartype, Size: size, IsArray: true,
		Assignment: assignment, Annotations: annotations,
	}
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
		return fmt.Errorf("failed opening file %s for writing: %w", fileName, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file %s: %s\n", fileName, err)
		}
	}()

	fileContent := fzw.HeaderComments + "\n"
	for _, fzparam := range mapValuesSortedByKey(fzw.ParamMap) {
		fileContent += fzparam.Declaration()
	}

	fileContent += "\n"
	for _, fzvar := range mapValuesSortedByKey(fzw.VarMap) {
		fileContent += fzvar.Declaration()
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
// If there can be no solution to the constraint problem (UNSAT), returns a slice with a single empty solution
// Otherwise, must return at least one solution, or return an error
func (fzw *FlatZincModel) ReadSolutions(fileName string) ([]CPSolution, error) {
	data, err := os.ReadFile(path.Clean(fileName))
	if err != nil {
		return nil, fmt.Errorf("failed opening file %s for reading: %w", fileName, err)
	}

	strContent := string(data)
	lines := strings.Split(strContent, "\n")
	res := []CPSolution{}
	currSolution := make(CPSolution)
	for lineNum, line := range lines {
		line = strings.Join(strings.Fields(line), "") // remove all whitespaces
		switch {
		case line == "":
			continue // empty line
		case strings.HasPrefix(line, "%%%"):
			continue // stat lines are ignored
		case line == "==========":
			if len(res) == 0 {
				return nil, errors.New("no solution was found, though solver says it did find solution(s)")
			}
			return res, nil // at least one solution was found and the whole search space was covered
		case line == "=====UNSATISFIABLE=====":
			return []CPSolution{make(CPSolution)}, nil // no solution exists; returns a single empty solution
		case strings.HasPrefix(line, "===="):
			err := fmt.Errorf("no solution found. Solver says %s", line) // no solution found (but not UNSAT either)
			return nil, err
		case line == "----------":
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

	if len(res) == 0 {
		return nil, errors.New("no solution was found; no solver status was found either")
	}
	return res, nil
}

// Reading a FlatZinc solutions file and returning the best solution
// When a minimize/maximize goal is defined, best solution should be the last solution
func (fzw *FlatZincModel) ReadBestSolution(fileName string) (CPSolution, error) {
	solutions, err := fzw.ReadSolutions(fileName)
	if err != nil {
		return nil, err
	}
	if len(solutions) < 1 {
		return nil, errors.New("no solution found")
	}
	return solutions[len(solutions)-1], nil
}

// helper functions

func mapValuesSortedByKey(mapToSort map[string]Declares) []Declares {
	keys := make([]string, 0, len(mapToSort))
	for k := range mapToSort {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	res := []Declares{}
	for _, k := range keys {
		res = append(res, mapToSort[k])
	}
	return res
}
