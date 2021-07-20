// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"testing"

	. "github.com/onsi/gomega"
)

// Validate a sample validation schema successfully
func TestValidateSchema_simple_success(t *testing.T) {
	taxonomyFile := "../../test/taxonomy/catalog.structs.schema.json"
	g := NewGomegaWithT(t)
	err := ValidateSchema(taxonomyFile)
	g.Expect(err).NotTo(HaveOccurred())
}

// Validate a sample badly formed validation schema
func TestValidateSchema_simple_fail(t *testing.T) {
	taxonomyFile := "../../test/taxonomy/bad_catalog.structs.schema.json"
	g := NewGomegaWithT(t)
	err := ValidateSchema(taxonomyFile)
	g.Expect(err).To(HaveOccurred())
}
