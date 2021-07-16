// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package slices

import (
	"bytes"
	"encoding/json"

	"github.com/mpvl/unique"
)

// UniqueInterfaceSlice removes duplicate entries from an interface slice
// where items are considered duplicate if their JSON representation
// is the same.
func UniqueInterfaceSlice(items *[]interface{}) {
	unique.Sort(interfaceSlice{items})
}

// interfaceSlice attaches the methods of unique.Interface to []interface{}.
type interfaceSlice struct{ P *[]interface{} }

func (p interfaceSlice) Len() int       { return len(*p.P) }
func (p interfaceSlice) Swap(i, j int)  { (*p.P)[i], (*p.P)[j] = (*p.P)[j], (*p.P)[i] }
func (p interfaceSlice) Truncate(n int) { *p.P = (*p.P)[:n] }
func (p interfaceSlice) Less(i, j int) bool {
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
