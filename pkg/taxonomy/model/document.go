// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package model

// Document represents a taxonomy schema document.
type Document struct {
	Schema        `json:",inline"`
	SchemaVersion string                `json:"$schema,omitempty"`
	Definitions   map[string]*SchemaRef `json:"definitions,omitempty"`
}

// Deref dereferences a $ref to the schema model that it points to
func (d *Document) Deref(in *SchemaRef) *SchemaRef {
	// TODO: support references across documents
	if in != nil && in.Ref != "" {
		return d.Definitions[in.RefName()]
	}
	return in
}
