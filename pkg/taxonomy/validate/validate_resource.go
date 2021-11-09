// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	log "log"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateResource validates the given resource JSON against the taxonomy file provided
func TaxonomyCheck(resourceJSON []byte, taxonomy string, resourceName string) []*field.Error {
	var allErrs []*field.Error

	// Load taxonomy from file mounted in configMap
	path, err := filepath.Abs(taxonomy)
	if err != nil {
		log.Printf("Invalid taxonomy filepath %s\n", err)
	}

	// Validate resource against taxonomy
	taxonomyLoader := gojsonschema.NewReferenceLoader("file://" + path)
	documentLoader := gojsonschema.NewStringLoader(string(resourceJSON))
	result, err := gojsonschema.Validate(taxonomyLoader, documentLoader)
	if err != nil {
		log.Printf("Could not validate resource against taxonomy provided %s\n", err)
	}

	// Return validation errors
	if result.Valid() {
		validMessage := "This " + resourceName + " is valid\n"
		log.Printf(validMessage)
	} else {
		invalidMessage := "This " + resourceName + " is not valid. See errors :\n"
		log.Printf(invalidMessage)
		for _, desc := range result.Errors() {
			log.Printf("- %s\n", desc)
			allErrs = append(allErrs, field.Invalid(field.NewPath(desc.Field()), desc.Value(), desc.Description()))
		}
	}
	return allErrs
}
