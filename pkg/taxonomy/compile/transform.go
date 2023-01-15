// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package compile

import (
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"

	"fybrik.io/fybrik/pkg/slices"
	"fybrik.io/fybrik/pkg/taxonomy/model"
)

const nameKey = "name"

// transform applies transformations over an input document to make it structural.
// It requires a base document and the document to transform (mutable).
// The codegenTarget can be set to true to apply more transformations to make the document more suitable for code
// generation tools.
func transform(base, doc *model.Document, codegenTarget bool) (*model.Document, error) {
	t := &transformer{
		doc:     doc,
		visited: map[string]bool{},
	}

	for definitionKey := range base.Definitions {
		if value, ok := t.doc.Definitions[definitionKey]; ok {
			t.transformSchema(definitionKey, value)
		}
	}

	if codegenTarget {
		for _, value := range t.doc.Definitions {
			t.removeComplexValidation(value)
		}
	}

	return t.doc, nil
}

type transformer struct {
	doc     *model.Document
	visited map[string]bool
}

func (t *transformer) transformSchema(key string, schema *model.SchemaRef) *model.SchemaRef {
	// Only visit each schema once
	if _, exists := t.visited[key]; exists {
		return schema
	}
	t.visited[key] = true

	result := schema

	// Change a oneOf union to structural definition with a property per each union subtype
	result = t.oneOfRefsTransform(key, result)

	// Enrich enum elements based on anyOf/oneOf/allOf validation
	result = t.propogateEnum(result)

	// TODO: do we need to recurse over Properties, Items, AdditionalProperties? AllOf, OneOf, AnyOf, Not?

	return result
}

func (t *transformer) oneOfRefsTransform(key string, schema *model.SchemaRef) *model.SchemaRef {
	if key == "" || schema.OneOf == nil || len(schema.Properties) != 1 || !schema.AdditionalProperties.IsAllowed() {
		// this transform does not apply here
		return schema
	}

	if _, exists := schema.Properties[nameKey]; !exists {
		// this transform does not apply here because name property is missing
		return schema
	}

	for _, v := range schema.OneOf {
		if v.Title == "" && v.Ref == "" {
			// this transform does not apply here because identifier for subtype is missing
			return schema
		}
	}

	var options []apiextensions.JSON
	for _, v := range schema.OneOf {
		// Extract name
		name := v.Title
		k := ""
		if v.Ref != "" {
			name = v.RefName()
			k = name
		}

		// Add to name options
		options = append(options, name)

		// Add property
		schema.Properties[name] = t.transformSchema(k, v)
	}

	// Add name enum definition
	nameDefinitionKey := fmt.Sprintf("%sName", key)
	t.doc.Definitions[nameDefinitionKey] = &model.SchemaRef{
		Schema: model.Schema{
			Type: "string",
			Enum: options,
		},
	}

	schema.Properties[nameKey] = &model.SchemaRef{
		Ref: fmt.Sprintf("#/definitions/%s", nameDefinitionKey),
	}

	// Add oneOf validation
	schema.OneOf = model.SchemaRefs{}
	for _, option := range options {
		schema.OneOf = append(schema.OneOf, &model.SchemaRef{
			Schema: model.Schema{
				Properties: model.Schemas{
					nameKey: &model.SchemaRef{Schema: model.Schema{Enum: []apiextensions.JSON{option}}},
				},
				Required: []string{nameKey, option.(string)},
			},
		})
	}

	// Set additional properties to false
	falseValue := false
	schema.AdditionalProperties.Allowed = &falseValue

	return schema
}

func (t *transformer) propogateEnum(schema *model.SchemaRef) *model.SchemaRef {
	for propertyName, property := range schema.Properties {
		property = t.doc.Deref(property)
		t.propogateEnumFromValidationGroup(propertyName, property, schema.AllOf)
		t.propogateEnumFromValidationGroup(propertyName, property, schema.AnyOf)
		t.propogateEnumFromValidationGroup(propertyName, property, schema.OneOf)

		slices.UniqueJSONSlice(&property.Enum)
	}
	return schema
}

func (t *transformer) propogateEnumFromValidationGroup(propertyName string, property *model.SchemaRef, validationGroup model.SchemaRefs) {
	for _, validationItem := range validationGroup {
		if validationItem.Properties != nil {
			if validationProperty, exists := validationItem.Properties[propertyName]; exists {
				if validationProperty.Enum != nil {
					property.Enum = append(property.Enum, validationProperty.Enum...)
				}
			}
		}
		t.propogateEnumFromValidationGroup(propertyName, property, validationItem.AllOf)
		t.propogateEnumFromValidationGroup(propertyName, property, validationItem.AnyOf)
		t.propogateEnumFromValidationGroup(propertyName, property, validationItem.OneOf)
	}
}

func (t *transformer) removeComplexValidation(schema *model.SchemaRef) *model.SchemaRef {
	if schema != nil {
		schema.AllOf = nil
		schema.OneOf = nil
		schema.AnyOf = nil
		schema.Not = nil
		for _, property := range schema.Properties {
			property = t.doc.Deref(property)
			t.removeComplexValidation(property)
		}
		t.removeComplexValidation(schema.Items)
		if schema.AdditionalProperties != nil {
			t.removeComplexValidation(schema.AdditionalProperties.Schema)
		}
	}
	return schema
}
