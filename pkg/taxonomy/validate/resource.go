// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"path/filepath"

	"emperror.dev/errors"
	"github.com/xeipuuv/gojsonschema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// TaxonomyCheck validates the given resource JSON against a  schema file
func TaxonomyCheck(resourceJSON []byte, schemaPath string) ([]*field.Error, error) {
	schemaPath, err := filepath.Abs(schemaPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not get absolute path for the schema")
	}

	// Validate resource against taxonomy
	taxonomyLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
	documentLoader := gojsonschema.NewBytesLoader(resourceJSON)
	result, err := gojsonschema.Validate(taxonomyLoader, documentLoader)
	if err != nil {
		return nil, errors.Wrap(err, "could not validate resource against the provided schema, check files at "+filepath.Dir(schemaPath))
	}

	// Return validation errors
	var allErrs []*field.Error

	if !result.Valid() {
		for _, desc := range result.Errors() {
			allErrs = append(allErrs, field.Invalid(field.NewPath(desc.Field()), desc.Value(), desc.Description()))
		}
	}

	return allErrs, nil
}
