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

type FlatZincParam struct {
	NumInstances uint // NumInstances > 1 means an array of params
	Name         string
	Type         string
	Assignment   string
}

func (fzp FlatZincParam) ParamDeclaration() string {
	if fzp.NumInstances == 1 {
		return fmt.Sprintf("%s: %s = %s;\n", fzp.Type, fzp.Name, fzp.Assignment)
	}
	return fmt.Sprintf("array [1..%d] of %s: %s = %s;\n", fzp.NumInstances, fzp.Type, fzp.Name, fzp.Assignment)
}

type Annotations []string

func (annots Annotations) AnnotationString() string {
	annotsStr := strings.Join(annots, " :: ")
	if len(annotsStr) > 0 {
		annotsStr = " :: " + annotsStr
	}
	return annotsStr
}

type FlatZincVariable struct {
	NumInstances uint // NumInstances > 1 means an array of vars
	Name         string
	Type         string
	Assignment   string
	Annotations
}

func (fzv FlatZincVariable) VarDeclaration() string {
	if fzv.NumInstances == 1 {
		return fmt.Sprintf("var %s: %s%s;\n", fzv.Type, fzv.Name, fzv.AnnotationString())
	}
	return fmt.Sprintf("array [1..%d] of var %s: %s%s;\n", fzv.NumInstances, fzv.Type, fzv.Name, fzv.AnnotationString())
}

type FlatZincConstraint struct {
	Identifier  string
	Expressions []string
	Annotations
}

func (cnstr FlatZincConstraint) ConstraintStatement() string {
	exprs := strings.Join(cnstr.Expressions, ", ")
	return fmt.Sprintf("constraint %s(%s)%s;\n", cnstr.Identifier, exprs, cnstr.AnnotationString())
}

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

type FlatZincSolveItem struct {
	goal SolveGoal
	expr string
	Annotations
}

func (slv FlatZincSolveItem) SolveItemStatement() string {
	return fmt.Sprintf("solve%s %s %s;\n", slv.AnnotationString(), slv.goal, slv.expr)
}

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
		fileContent += fzparam.ParamDeclaration()
	}

	fileContent += "\n"
	for _, fzvar := range fzw.VarMap {
		fileContent += fzvar.VarDeclaration()
	}

	fileContent += "\n"
	for _, cnstr := range fzw.Constraints {
		fileContent += cnstr.ConstraintStatement()
	}

	fileContent += "\n" + fzw.SolveTarget.SolveItemStatement()
	if _, err := file.WriteString(fileContent); err != nil {
		return err
	}
	return nil
}

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

func (fzw *FlatZincModel) ReadSolutions(fileName string) (map[string][]string, error) {
	data, err := os.ReadFile(path.Clean(fileName))
	if err != nil {
		return nil, err
	}
	strContent := string(data)
	lines := strings.Split(strContent, "\n")
	newSolution := false
	res := make(map[string][]string)
	for lineNum, line := range lines {
		line = strings.Join(strings.Fields(line), "") // remove all whitespaces
		switch {
		case len(line) == 0:
			continue // empty line
		case strings.HasPrefix(line, "%%%"):
			continue // stat lines are ignored
		case line == strings.Repeat("=", 10):
			return res, nil // a solution was found and the whole search space was covered
		case strings.HasPrefix(line, "===="):
			err := fmt.Errorf("solution not found: %s", line)
			return nil, err
		case line == strings.Repeat("-", 10):
			newSolution = true // marks the end of current solution (and possible the beginning of a new one)
		default: // this should be a variable assignment line
			if newSolution {
				for k := range res {
					delete(res, k)
				}
				newSolution = false
			}
			varName, values, err := parseSolutionLine(line, uint(lineNum))
			if err != nil {
				return nil, err
			}
			res[varName] = values
		}
	}

	return res, nil
}
