// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package compile

import (
	"fmt"

	"github.com/mesh-for-data/mesh-for-data/pkg/slices"
	"github.com/mesh-for-data/mesh-for-data/pkg/taxonomy/model"
	strcase "github.com/stoewer/go-strcase"
)

// transform applies transformations over an input document to make it structural.
// It requires a base document and the document to transform (mutable).
// The codegenTarget can be set to true to apply more transformations to make the document
//  more suitable for code generation tools.
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
			// t.adjustAdditionalProperties(value)
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
	result = t.propogateEnum(key, result)

	// TODO: do we need to recurse over Properties, Items, AdditionalProperties? AllOf, OneOf, AnyOf, Not?

	return result
}

func (t *transformer) oneOfRefsTransform(key string, schema *model.SchemaRef) *model.SchemaRef {

	if key == "" || schema.OneOf == nil || len(schema.Properties) != 1 || !schema.AdditionalProperties.IsAllowed() {
		// this transform does not apply here
		return schema
	}

	if nameProp, ok := schema.Properties["name"]; !ok || nameProp.Type != "string" {
		// this transform does not apply here because name property is missing
		return schema
	}

	for _, v := range schema.OneOf {
		if v.Title == "" && v.Ref == "" {
			// this transform does not apply here because identifier for subtype is missing
			return schema
		}
	}

	var options []interface{}
	for _, v := range schema.OneOf {
		// Extract name
		name := v.Title
		key := ""
		if v.Ref != "" {
			name = v.RefName()
			key = name
		}
		name = strcase.LowerCamelCase(name)

		// Add to name options
		options = append(options, name)

		// Add property
		schema.Properties[name] = t.transformSchema(key, v)
	}

	// Add name enum definition
	nameDefinitionKey := fmt.Sprintf("%sName", key)
	t.doc.Definitions[nameDefinitionKey] = &model.SchemaRef{
		Schema: model.Schema{
			Type: "string",
			Enum: options,
		},
	}

	schema.Properties["name"] = &model.SchemaRef{
		Ref: fmt.Sprintf("#/definitions/%s", nameDefinitionKey),
	}

	// Add oneOf validation
	schema.OneOf = model.SchemaRefs{}
	for _, option := range options {
		schema.OneOf = append(schema.OneOf, &model.SchemaRef{
			Schema: model.Schema{
				Properties: model.Schemas{
					"name": &model.SchemaRef{Schema: model.Schema{Enum: []interface{}{option}}},
				},
				Required: []string{"name", option.(string)},
			},
		})
	}

	// Set additional properties to false
	falseValue := false
	schema.AdditionalProperties.Allowed = &falseValue

	return schema
}

func (t *transformer) propogateEnum(key string, schema *model.SchemaRef) *model.SchemaRef {
	for propertyName, property := range schema.Properties {
		property := t.doc.Deref(property)
		t.propogateEnumFromValidationGroup(propertyName, property, schema.AllOf)
		t.propogateEnumFromValidationGroup(propertyName, property, schema.AnyOf)
		t.propogateEnumFromValidationGroup(propertyName, property, schema.OneOf)

		slices.UniqueInterfaceSlice(&property.Enum)
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
			property := t.doc.Deref(property)
			t.removeComplexValidation(property)
		}
		t.removeComplexValidation(schema.Items)
		if schema.AdditionalProperties != nil {
			t.removeComplexValidation(schema.AdditionalProperties.Schema)
		}
	}
	return schema
}

// func (t *transformer) adjustAdditionalProperties(schema *model.SchemaRef) *model.SchemaRef {
// 	if schema != nil {
// 		if schema.AdditionalProperties.IsAllowed() {
// 			schema.AdditionalProperties.Allowed = nil
// 			schema.AdditionalProperties.Schema = &model.SchemaRef{Schema: model.Schema{}}
// 		}
// 		for _, property := range schema.Properties {
// 			property := t.doc.Deref(property)
// 			t.adjustAdditionalProperties(property)
// 		}
// 		t.adjustAdditionalProperties(schema.Items)
// 		if schema.AdditionalProperties != nil {
// 			t.adjustAdditionalProperties(schema.AdditionalProperties.Schema)
// 		}
// 	}
// 	return schema
// }
