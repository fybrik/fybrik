// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package model

import "strings"

// SchemaRef is either a schema or a reference to a schema
type SchemaRef struct {
	Schema
	Ref string `json:"$ref,omitempty"`
}

// SchemaRefs is a list of schemas/references
type SchemaRefs []*SchemaRef

// Schemas is a map of schemas/references
type Schemas map[string]*SchemaRef

// RefName returns the name from a reference.
// For example given a reference `$ref: "#/definitions/MyObject"`
//  the returned value is "MyObject".
func (s *SchemaRef) RefName() string {
	if s != nil && s.Ref != "" {
		return s.Ref[strings.LastIndex(s.Ref, "/")+1:]
	}
	return ""
}
