// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

// This file contains functions to create high-level constructs on top of the low-level FlatZinc interface

package optimizer

import (
	"fmt"
	"strconv"
	"strings"
)

// ***********
// Indicators are Boolean variables that are set to true on some condition
// ***********

// Adds an indicator array whose elements are true iff a given integer variable EQUALS a given value in a given pos
// Setting 'equality' to false will make the indicators true iff the integer DOES NOT EQUAL the given value
func equalityIndicator(fzModel *FlatZincModel, variable string, value, pathLength int, equality bool) string {
	constraint := IntEqConstraint
	if !equality {
		constraint = IntNotEqConstraint
	}
	indicator := fmt.Sprintf("ind_%s_%s_%d", variable, constraint, value)
	if _, defined := fzModel.VarMap[indicator]; defined {
		return indicator
	}

	fzModel.AddVariableArray(indicator, BoolType, pathLength, true, false)
	strVal := strconv.Itoa(value)
	for pathPos := 1; pathPos <= pathLength; pathPos++ {
		variableAtPos := varAtPos(variable, pathPos)
		indicatorAtPos := varAtPos(indicator, pathPos)
		annotation := GetDefinesVarAnnotation(indicatorAtPos)
		fzModel.AddConstraint(constraint, []string{variableAtPos, strVal, indicatorAtPos}, annotation)
	}
	return indicator
}

// Given a Boolean variable, returns indicator variable array which is true iff the variable is false at each pos
func boolNotIndicator(fzModel *FlatZincModel, variable string, pathLength int) string {
	indicator := fmt.Sprintf("ind_not_%s", variable)
	if _, defined := fzModel.VarMap[indicator]; defined {
		return indicator
	}
	fzModel.AddVariableArray(indicator, BoolType, pathLength, true, false)
	for pathPos := 1; pathPos <= pathLength; pathPos++ {
		indicatorAtPos := varAtPos(indicator, pathPos)
		annotation := GetDefinesVarAnnotation(indicatorAtPos)
		fzModel.AddConstraint(BoolNotEqConstraint, []string{varAtPos(variable, pathPos), indicatorAtPos}, annotation)
	}

	return indicator
}

// Adds a Boolean indicator variable that is implied by either the given var or its negation
func impliedByIndicator(fzModel *FlatZincModel, variable string, pathLen int, impliedByNegatedVar bool) string {
	notStr := ""
	if impliedByNegatedVar {
		notStr = "not_"
	}
	indicator := fmt.Sprintf("ind_implied_by_%s%s", notStr, variable)
	if _, defined := fzModel.VarMap[indicator]; defined {
		return indicator
	}
	fzModel.AddVariableArray(indicator, BoolType, pathLen, true, false)
	for pathPos := 1; pathPos <= pathLen; pathPos++ {
		variableAtPos := varAtPos(variable, pathPos)
		indicatorAtPos := varAtPos(indicator, pathPos)
		annotations := GetDefinesVarAnnotation(indicatorAtPos)
		if impliedByNegatedVar {
			arrayToOr := fznCompoundLiteral([]string{variableAtPos, indicatorAtPos}, false)
			fzModel.AddConstraint(ArrBoolOrConstraint, []string{arrayToOr, TrueValue}, annotations)
		} else {
			fzModel.AddConstraint(BoolLeConstraint, []string{variableAtPos, indicatorAtPos}, annotations)
		}
	}

	return indicator
}

// Adds an indicator per path location to check if the value of "variable" in this location is in the given set of values
func setInIndicator(fzModel *FlatZincModel, variable string, valueSet []string, pathLength int) string {
	indicator := fmt.Sprintf("ind_%s_in_%s", variable, strings.Join(valueSet, "_"))
	if _, defined := fzModel.VarMap[indicator]; defined {
		return indicator
	}

	fzModel.AddVariableArray(indicator, BoolType, pathLength, true, false)
	if len(valueSet) > 0 {
		for pathPos := 1; pathPos <= pathLength; pathPos++ {
			variableAtPos := varAtPos(variable, pathPos)
			indicatorAtPos := varAtPos(indicator, pathPos)
			setLiteral := fznCompoundLiteral(valueSet, true)
			annotation := GetDefinesVarAnnotation(indicatorAtPos)
			fzModel.AddConstraint(SetInConstraint, []string{variableAtPos, setLiteral, indicatorAtPos}, annotation)
		}
	} else { // value set is empty - indicators should always be false as variable value is never in the given set
		fzModel.AddConstraint(ArrBoolOrConstraint, []string{indicator, FalseValue})
	}
	return indicator
}

// Returns a variable which holds the OR of all boolean variables in indicatorArray
func orOfIndicators(fzModel *FlatZincModel, indicatorArray string) string {
	bigOrVarname := indicatorArray + "_OR"
	fzModel.AddVariable(bigOrVarname, BoolType, true, false)
	annotation := GetDefinesVarAnnotation(bigOrVarname)
	fzModel.AddConstraint(ArrBoolOrConstraint, []string{indicatorArray, bigOrVarname}, annotation)
	return bigOrVarname
}

// ***********
// Functions to add some high-level constraints
// ***********

// Replicates a constraint to block a specific combination of assignments for each position in the path
func preventAssignments(fzModel *FlatZincModel, variables []string, values []int, pathLength int) {
	// Prepare an indicator for each variable which is true iff the variable is NOT assigned its given value
	indicators := []string{}
	for idx, variable := range variables {
		if fzModel.GetVariableType(variable) == BoolType {
			if values[idx] == 0 { // if the variable is Boolean, we assume that "false" is 0 and "true" is anything else
				indicators = append(indicators, variable) // the var is an indicator for itself not being "false"
			} else {
				indicators = append(indicators, boolNotIndicator(fzModel, variable, pathLength))
			}
		} else {
			indicators = append(indicators, equalityIndicator(fzModel, variable, values[idx], pathLength, false))
		}
	}

	for pos := 1; pos <= pathLength; pos++ {
		indexedIndicators := []string{}
		for _, v := range indicators {
			indexedIndicators = append(indexedIndicators, varAtPos(v, pos))
		}
		indicatorsArray := fznCompoundLiteral(indexedIndicators, false)
		fzModel.AddConstraint(ArrBoolOrConstraint, []string{indicatorsArray, TrueValue})
	}
}

// "varToAssign" gets assigned with "valIfTrue" if "selectorVar" is true, and with "valIfFalse" otherwise
func assignWithSelector(fzModel *FlatZincModel, varToAssign, selectorVar string, valIfTrue, valIfFalse []string, pathLen int) {
	impliedBySelector := impliedByIndicator(fzModel, selectorVar, pathLen, false)
	impliedByNotSelector := impliedByIndicator(fzModel, selectorVar, pathLen, true)
	for pos := 1; pos <= pathLen; pos++ {
		varToAssignAtPos := varAtPos(varToAssign, pos)
		impliedBySelectorAtPos := varAtPos(impliedBySelector, pos)
		impliedByNotSelectorAtPos := varAtPos(impliedByNotSelector, pos)
		fzModel.AddConstraint(IntEqConstraint, []string{varToAssignAtPos, valIfTrue[pos-1], impliedBySelectorAtPos})
		fzModel.AddConstraint(IntEqConstraint, []string{varToAssignAtPos, valIfFalse[pos-1], impliedByNotSelectorAtPos})
	}
}

// Sets the CSP int variable sumVarname to be the weighted sum of int elements in arrayToSum.
// The integer weight of each element is given in the array "weights".
// FlatZinc doesn't give us a "weighted sum" constraint (and not even sum constraint).
// The trick is to use the dot-product constraint, add the summing var with weight -1 and force the result to be 0
func setVarAsWeightedSum(fzModel *FlatZincModel, sumVarname string, arrayToSum, weights []string) {
	arrayToSum = append(arrayToSum, sumVarname)
	weights = append(weights, minusOneStr)
	fzModel.AddConstraint(
		IntLinEqConstraint,
		[]string{fznCompoundLiteral(weights, false), fznCompoundLiteral(arrayToSum, false), strconv.Itoa(0)},
		GetDefinesVarAnnotation(sumVarname),
	)
}

// Sets the CSP int variable sumVarname to be the weighted sum of the elements in the variable array varArrayToSum.
func setVarAsWeightedSumOfVarArray(fzModel *FlatZincModel, sumVarname, varArrayToSum string, weightsArray []string) {
	arrayToSum := arrayOfVarPositions(varArrayToSum, len(weightsArray))
	setVarAsWeightedSum(fzModel, sumVarname, arrayToSum, weightsArray)
}

// Sets the CSP int variable sumVarname to be the sum of the elements in the variable array varArrayToSum.
func setVarAsSimpleSumOfVarArray(fzModel *FlatZincModel, sumVarname, varArrayToSum string) {
	arrayLen := fzModel.GetVariableSize(varArrayToSum)
	setVarAsWeightedSumOfVarArray(fzModel, sumVarname, varArrayToSum, arrayOfSameInt(1, arrayLen))
}

// ***********
// Utility functions
// ***********

func arrayOfSameStr(str string, arrayLen int) []string {
	array := make([]string, arrayLen)
	for i := 0; i < arrayLen; i++ {
		array[i] = str
	}
	return array
}

func arrayOfSameInt(num, arrayLen int) []string {
	return arrayOfSameStr(strconv.Itoa(num), arrayLen)
}

func arrayOfVarPositions(variableArray string, arrayLen int) []string {
	array := make([]string, arrayLen)
	for i := 1; i <= arrayLen; i++ {
		array[i-1] = varAtPos(variableArray, i)
	}
	return array
}
