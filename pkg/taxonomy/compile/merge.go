// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package compile

import (
	"github.com/mesh-for-data/mesh-for-data/pkg/slices"
	"github.com/mesh-for-data/mesh-for-data/pkg/taxonomy/model"
	"github.com/mpvl/unique"
)

// mergeDefinitions merges the definitions section of multiple documents.
// Any definition that appears only in one document is added as is.
// Multiple definitions with the same name are merged recursively.
func mergeDefinitions(documents ...*model.Document) (*model.Document, error) {
	result := &model.Document{
		Definitions: make(map[string]*model.SchemaRef),
	}

	for _, doc := range documents {
		merge(result, doc)
	}

	return result, nil
}

func merge(dst, src *model.Document) {
	for definitionKey, srcSchemaRef := range src.Definitions {
		dst.Definitions[definitionKey] = mergeSchemaRef(dst.Definitions[definitionKey], srcSchemaRef)
	}
}

func maybeOverrideString(dst *string, src string) {
	if src != "" {
		*dst = src
	}
}

func mergeSchemaRef(dst, src *model.SchemaRef) *model.SchemaRef {
	if dst == nil {
		return src
	}

	if src == nil {
		return dst
	}

	// maybeOverrideString(&dst.Ref, src.Ref)
	// maybeOverrideString(&dst.Type, src.Type)
	maybeOverrideString(&dst.Title, src.Title)
	maybeOverrideString(&dst.Description, src.Description)

	// Merge properties
	if src.Properties != nil && dst.Properties == nil {
		dst.Properties = model.Schemas{}
	}
	for propertyName, srcProperty := range src.Properties {
		dst.Properties[propertyName] = mergeSchemaRef(dst.Properties[propertyName], srcProperty)
	}

	// Merge additional properties
	if src.AdditionalProperties != nil {
		if src.AdditionalProperties.Allowed != nil {
			// handle override to `additionalProperties: true` or `additionalProperties: false`
			dst.AdditionalProperties = &model.AdditionalPropertiesType{
				Allowed: src.AdditionalProperties.Allowed,
				Schema:  nil,
			}
		} else if src.AdditionalProperties.Schema != nil {
			// handle override to `additionalProperties: [object]`
			var maybeSchema *model.SchemaRef = nil
			if dst.AdditionalProperties != nil {
				maybeSchema = dst.AdditionalProperties.Schema
			}
			dst.AdditionalProperties = &model.AdditionalPropertiesType{
				Allowed: nil,
				Schema:  mergeSchemaRef(maybeSchema, src.AdditionalProperties.Schema),
			}
		}
	}

	dst.Items = mergeSchemaRef(dst.Items, src.Items)

	if src.Default != nil {
		dst.Default = src.Default
	}

	dst.OneOf = append(dst.OneOf, src.OneOf...)
	dst.AllOf = append(dst.AllOf, src.AllOf...)
	dst.AnyOf = append(dst.AnyOf, src.AnyOf...)
	dst.Not = mergeSchemaRef(dst.Not, src.Not)

	dst.Required = append(dst.Required, src.Required...)
	unique.Strings(&dst.Required)

	if src.MinProps > 0 {
		dst.MinProps = src.MinProps
	}
	if src.MaxProps != nil {
		dst.MaxProps = src.MaxProps
	}

	if src.Format != "" {
		dst.Format = src.Format
	}

	dst.Enum = append(dst.Enum, src.Enum...)
	slices.UniqueInterfaceSlice(&dst.Enum)

	if src.MinLength > 0 {
		dst.MinLength = src.MinLength
	}
	if src.MaxLength != nil {
		dst.MaxLength = src.MaxLength
	}
	if src.Pattern != "" {
		dst.Pattern = src.Pattern
	}

	if src.Min != nil {
		dst.Min = src.Min
	}
	if src.Max != nil {
		dst.Max = src.Max
	}
	if src.MultipleOf != nil {
		dst.MultipleOf = src.MultipleOf
	}
	if src.ExclusiveMin {
		dst.ExclusiveMin = src.ExclusiveMin
	}
	if src.ExclusiveMax {
		dst.ExclusiveMax = src.ExclusiveMax
	}
	if src.MinItems > 0 {
		dst.MinItems = src.MinItems
	}
	if src.MaxItems != nil {
		dst.MaxItems = src.MaxItems
	}

	if src.UniqueItems {
		dst.UniqueItems = src.UniqueItems
	}

	return dst
}
