// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"fybrik.io/fybrik/pkg/environment"
)

/*
This file implements FlatZincModel: a class to hold a constraint problem, based on the FlatZinc format.
The class can dump the constraint program to a file, using the FlatZinc specification.
The class can also read solver solutions, written using the FlatZinc specification.
The FlatZinc specification: https://www.minizinc.org/doc-latest/en/fzn-spec.html
*/

const (
	TrueValue  = "true"
	FalseValue = "false"

	BoolType = "bool"
	IntType  = "int"

	BoolLeConstraint     = "bool_le"
	BoolLinEqConstraint  = "bool_lin_eq"
	BoolLinLeConstraint  = "bool_lin_le"
	BoolNotEqConstraint  = "bool_not"
	ArrBoolOrConstraint  = "array_bool_or"
	IntEqConstraint      = "int_eq_reif"
	IntNotEqConstraint   = "int_ne_reif"
	IntLeConstraint      = "int_le_reif"
	SetInConstraint      = "set_in_reif"
	IntLinEqConstraint   = "int_lin_eq"
	ArrIntElemConstraint = "array_var_int_element"
	IntMaxConstraint     = "array_int_maximum"

	DefinedVarAnnotation = "is_defined_var"
	DefinesVarAnnotation = "defines_var(%s)"
	OutputVarAnnotation  = "output_var"
	OutputArrAnnotation  = "output_array([1..%d])"

	ElementSeparator = ", "
)

type Declares interface {
	Declaration() string
	GetSize() int
	GetType() string
	SetAssignment(string)
}

// Data for a single FlatZinc parameter
type FlatZincParam struct {
	Name       string
	Type       string
	Size       int
	IsArray    bool // (IsArray == false) implies (Size == 1)
	Assignment string
}

// formats a parameter declaration in FlatZinc format
func (fzp *FlatZincParam) Declaration() string {
	if fzp.IsArray {
		return fmt.Sprintf("array [1..%d] of %s: %s = %s;\n", fzp.Size, fzp.Type, fzp.Name, fzp.Assignment)
	}
	return fmt.Sprintf("%s: %s = %s;\n", fzp.Type, fzp.Name, fzp.Assignment)
}

func (fzp *FlatZincParam) GetSize() int {
	return fzp.Size
}

func (fzp *FlatZincParam) GetType() string {
	return fzp.Type
}

func (fzp *FlatZincParam) SetAssignment(assignment string) {
	fzp.Assignment = assignment
}

type Annotations []string

func (annotations Annotations) annotationString() string {
	// prepending an empty string, because annotations always start with "::"
	return strings.Join(append([]string{""}, annotations...), " :: ")
}

func GetDefinesVarAnnotation(variable string) string {
	return fmt.Sprintf(DefinesVarAnnotation, variable)
}

// Data for a single FlatZinc variable
type FlatZincVariable struct {
	Name       string
	Type       string
	Size       int
	IsArray    bool // (IsArray == false) implies (Size == 1)
	Assignment string
	Annotations
}

// formats a variable declaration in FlatZinc format
func (fzv *FlatZincVariable) Declaration() string {
	res := ""
	if fzv.IsArray {
		res = fmt.Sprintf("array [1..%d] of ", fzv.Size)
	}
	res += fmt.Sprintf("var %s: %s%s", fzv.Type, fzv.Name, fzv.annotationString())
	if len(fzv.Assignment) > 0 {
		res += fmt.Sprintf(" = %s", fzv.Assignment)
	}
	res += ";\n"
	return res
}

func (fzv *FlatZincVariable) GetSize() int {
	return fzv.Size
}

func (fzv *FlatZincVariable) GetType() string {
	return fzv.Type
}

func (fzv *FlatZincVariable) SetAssignment(assignment string) {
	fzv.Assignment = assignment
}

// Data for a single FlatZinc constraint
type FlatZincConstraint struct {
	Identifier  string
	Expressions []string
	Annotations
}

// formats a constraint statement in FlatZinc format
func (constraint *FlatZincConstraint) constraintStatement() string {
	exprs := strings.Join(constraint.Expressions, ElementSeparator)
	return fmt.Sprintf("constraint %s(%s)%s;\n", constraint.Identifier, exprs, constraint.annotationString())
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
	fzw.ParamMap = map[string]Declares{}
	fzw.VarMap = map[string]Declares{}
	return &fzw
}

func (fzw *FlatZincModel) AddHeaderComment(commentLine string) {
	fzw.HeaderComments = fzw.HeaderComments + "% " + commentLine + "\n"
}

func (fzw *FlatZincModel) AddParam(name, varType, assignment string) {
	fzw.ParamMap[name] = &FlatZincParam{Name: name, Type: varType, Size: 1, IsArray: false, Assignment: assignment}
}

func (fzw *FlatZincModel) AddParamArray(name, varType string, size int, assignment string) {
	fzw.ParamMap[name] = &FlatZincParam{Name: name, Type: varType, Size: size, IsArray: true, Assignment: assignment}
}

func (fzw *FlatZincModel) AddVariable(name, varType string, isDefined, isOutput bool) {
	annotations := []string{}
	if isDefined {
		annotations = append(annotations, DefinedVarAnnotation)
	}
	if isOutput {
		annotations = append(annotations, OutputVarAnnotation)
	}
	fzw.VarMap[name] = &FlatZincVariable{Name: name, Type: varType, Size: 1, IsArray: false, Annotations: annotations}
}

func (fzw *FlatZincModel) AddVariableArray(name, varType string, size int, isDefined, isOutput bool) {
	annotations := []string{}
	if isDefined {
		annotations = append(annotations, DefinedVarAnnotation)
	}
	if isOutput {
		annotations = append(annotations, fmt.Sprintf(OutputArrAnnotation, size))
	}
	fzw.VarMap[name] = &FlatZincVariable{Name: name, Type: varType, Size: size, IsArray: true, Annotations: annotations}
}

func (fzw *FlatZincModel) SetVariableAssignment(name, assignment string) {
	if _, found := fzw.ParamMap[name]; found {
		fzw.ParamMap[name].SetAssignment(assignment)
	}
	if _, found := fzw.VarMap[name]; found {
		fzw.VarMap[name].SetAssignment(assignment)
	}
}

func (fzw *FlatZincModel) GetVariableSize(name string) int {
	if param, found := fzw.ParamMap[name]; found {
		return param.GetSize()
	}

	if variable, found := fzw.VarMap[name]; found {
		return variable.GetSize()
	}

	return 0
}

func (fzw *FlatZincModel) GetVariableType(name string) string {
	if param, found := fzw.ParamMap[name]; found {
		return param.GetType()
	}

	if variable, found := fzw.VarMap[name]; found {
		return variable.GetType()
	}

	return ""
}

func (fzw *FlatZincModel) varArrayElements(varName string) []string {
	varSize := fzw.GetVariableSize(varName)
	if varSize == 0 {
		return []string{}
	}
	elements := []string{}
	for i := 1; i <= varSize; i++ {
		elements = append(elements, varAtPos(varName, i))
	}
	return elements
}

func (fzw *FlatZincModel) AddConstraint(identifier string, exprs []string, annotations ...string) {
	fzw.Constraints = append(fzw.Constraints, FlatZincConstraint{identifier, exprs, annotations})
}

func (fzw *FlatZincModel) SetSolveTarget(goal SolveGoal, expr string, annotations ...string) {
	fzw.SolveTarget = FlatZincSolveItem{goal, expr, annotations}
}

func (fzw *FlatZincModel) Clear() {
	fzw.ParamMap = map[string]Declares{}
	fzw.VarMap = map[string]Declares{}
	fzw.Constraints = []FlatZincConstraint{}
}

// dumps a FlatZinc model to a temp file using the FlatZinc syntax, returning the file name
// It is the caller responsibility to delete the file
func (fzw *FlatZincModel) Dump() (string, error) {
	file, err := os.CreateTemp(environment.GetDataDir(), "DataPathModel.*.fzn")
	if err != nil {
		return "", fmt.Errorf("failed creating temp file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file %s: %s\n", file.Name(), err)
		}
	}()

	fileContent := fzw.HeaderComments + "\n"
	for _, fzParam := range mapValuesSortedByKey(fzw.ParamMap) {
		fileContent += fzParam.Declaration()
	}

	fileContent += "\n"
	for _, fzVar := range mapValuesSortedByKey(fzw.VarMap) {
		fileContent += fzVar.Declaration()
	}

	fileContent += "\n"
	for _, constraint := range fzw.Constraints {
		fileContent += constraint.constraintStatement()
	}

	fileContent += "\n" + fzw.SolveTarget.solveItemStatement()
	if _, err := file.WriteString(fileContent); err != nil {
		return file.Name(), err
	}
	return file.Name(), nil
}

// Parses a single variable assignment line in a FlatZinc solution file. Returns the variable name and its value(s)
func parseSolutionLine(line string, lineNum int) (string, []string, error) {
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

// Reading FlatZinc-solver solutions and returning them as a slice of CPSolution
// If there can be no solution to the constraint problem (UNSAT), returns a slice with a single empty solution
// Otherwise, must return at least one solution, or return an error
func (fzw *FlatZincModel) ReadSolutions(solverOutput string) ([]CPSolution, error) {
	solverOutput = strings.ReplaceAll(solverOutput, "\r", "") // in case we run on Windows
	lines := strings.Split(solverOutput, "\n")
	res := []CPSolution{}
	currentSolution := make(CPSolution)
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
			return []CPSolution{{}}, nil // no solution exists; returns a single empty solution
		case strings.HasPrefix(line, "===="):
			err := fmt.Errorf("no solution found. Solver says %s", line) // no solution found (but not UNSAT either)
			return nil, err
		case line == "----------":
			res = append(res, currentSolution) // marks the end of current solution
			currentSolution = CPSolution{}     // (and possible the beginning of a new one)
		default: // this should be a variable assignment line
			varName, values, err := parseSolutionLine(line, lineNum)
			if err != nil {
				return nil, err
			}
			currentSolution[varName] = values
		}
	}

	if len(res) == 0 {
		return nil, errors.New("no solution was found; no solver status was found either")
	}
	return res, nil
}

// Reading FlatZinc-solver solutions and returning the best one
// When a minimize/maximize goal is defined, best solution should be the last solution
func (fzw *FlatZincModel) ReadBestSolution(solverOutput string) (CPSolution, error) {
	solutions, err := fzw.ReadSolutions(solverOutput)
	if err != nil {
		return nil, err
	}
	if len(solutions) < 1 {
		return nil, errors.New("no solution found")
	}
	return solutions[len(solutions)-1], nil
}

// Just like ReadSolutions() but reading the solutions from a file
func (fzw *FlatZincModel) ReadSolutionsFromFile(fileName string) ([]CPSolution, error) {
	fileContent, err := getFileContent(fileName)
	if err != nil {
		return nil, err
	}

	return fzw.ReadSolutions(fileContent)
}

// Just like ReadBestSolution() but reading the solutions from a file
func (fzw *FlatZincModel) ReadBestSolutionFromFile(fileName string) (CPSolution, error) {
	fileContent, err := getFileContent(fileName)
	if err != nil {
		return nil, err
	}

	return fzw.ReadBestSolution(fileContent)
}

// helper functions

func getFileContent(fileName string) (string, error) {
	data, err := os.ReadFile(path.Clean(fileName))
	if err != nil {
		return "", fmt.Errorf("failed opening file %s for reading: %w", fileName, err)
	}
	return string(data), nil
}

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

// Given a slice of identifiers/constants, returns either a FlatZinc array "[a1, a2, ..., an]"
// or a FlatZinc set "{a1, a2, ..., an}" made from these identifiers/constants
func fznCompoundLiteral(values []string, isSet bool) string {
	jointValues := strings.Join(values, ElementSeparator)
	if isSet {
		return fmt.Sprintf("{%s}", jointValues)
	}
	return fmt.Sprintf("[%s]", jointValues)
}

func fznRangeVarType(rangeStart, rangeEnd int) string {
	if rangeEnd < rangeStart {
		rangeEnd = rangeStart
	}
	return fmt.Sprintf("%d..%d", rangeStart, rangeEnd)
}

// given an arbitrary string, returns a legal FlatZinc identifier, based on this string
// with illegal characters replaced by their ASCII value
func sanitizeFznIdentifier(identifier string) string {
	if identifier == "" {
		return "empty"
	}

	match, _ := regexp.MatchString("^[A-Za-z][A-Za-z0-9_]+$", identifier)
	if match {
		return identifier
	}

	var resStr strings.Builder
	for i := 0; i < len(identifier); i++ {
		ch := identifier[i]
		if ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') {
			resStr.WriteByte(ch)
		} else if (('0' <= ch && ch <= '9') || ch == '_') && i > 0 { // An identifier must not BEGIN with numerics or '_'
			resStr.WriteByte(ch)
		} else {
			resStr.WriteString("A" + strconv.Itoa(int(ch)))
		}
	}

	return resStr.String()
}

func varAtPos(variable string, pos int) string {
	return fmt.Sprintf("%s[%d]", variable, pos)
}
