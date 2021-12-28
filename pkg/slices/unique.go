// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package slices

import (
	"bytes"
	"encoding/json"

	"github.com/mpvl/unique"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

// UniqueJSONSlice removes duplicate entries from an interface slice
// where items are considered duplicate if their JSON representation
// is the same.
func UniqueJSONSlice(items *[]apiextensions.JSON) {
	unique.Sort(jsonSlice{items})
}

// interfaceSlice attaches the methods of unique.Interface to []apiextensions.JSON.
type jsonSlice struct{ P *[]apiextensions.JSON }

func (p jsonSlice) Len() int       { return len(*p.P) }
func (p jsonSlice) Swap(i, j int)  { (*p.P)[i], (*p.P)[j] = (*p.P)[j], (*p.P)[i] }
func (p jsonSlice) Truncate(n int) { *p.P = (*p.P)[:n] }
func (p jsonSlice) Less(i, j int) bool {
	left := (*p.P)[i]
	leftBytes, err := json.Marshal(left)
	if err != nil {
		panic(err)
	}

	right := (*p.P)[j]
	rightBytes, err := json.Marshal(right)
	if err != nil {
		panic(err)
	}

	return bytes.Compare(leftBytes, rightBytes) < 0
}
