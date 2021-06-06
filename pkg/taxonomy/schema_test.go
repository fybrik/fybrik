package taxonomy

import (
	"testing"

	. "github.com/onsi/gomega"
)

/*
func TestValidateSchema_embedded_module(t *testing.T) {
	taxonomy_file := "module.structs.schema.json"
	g := NewGomegaWithT(t)
	err := ValidateSchema("../../config/taxonomy/" + taxonomy_file)
	g.Expect(err).NotTo(HaveOccurred())
}
*/

/*
func TestValidateSchema_embedded_policymanager(t *testing.T) {
	taxonomy_file := "policymanager.structs.schema.json"
	g := NewGomegaWithT(t)
	err := ValidateSchema("../../config/taxonomy/" + taxonomy_file)
	g.Expect(err).NotTo(HaveOccurred())
}
*/

func TestValidateSchema_simple_success(t *testing.T) {
	taxonomy_file := "catalog.structs.schema.json"
	g := NewGomegaWithT(t)
	err := ValidateSchema("./" + taxonomy_file)
	g.Expect(err).NotTo(HaveOccurred())
}
func TestValidateSchema_simple_fail(t *testing.T) {
	taxonomy_file := "bad_catalog.structs.schema.json"
	g := NewGomegaWithT(t)
	err := ValidateSchema("./" + taxonomy_file)
	g.Expect(err).To(HaveOccurred())
}
