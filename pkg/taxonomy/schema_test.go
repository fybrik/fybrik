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
	taxonomyFile := "policymanager.structs.schema.json"
	g := NewGomegaWithT(t)
	err := ValidateSchema("../../config/taxonomy/" + taxonomyFile)
	g.Expect(err).NotTo(HaveOccurred())
}
*/
func TestValidateSchema_simple_success(t *testing.T) {
	taxonomyFile := "catalog.structs.schema.json"
	g := NewGomegaWithT(t)
	err := ValidateSchema("./" + taxonomyFile)
	g.Expect(err).NotTo(HaveOccurred())
}
func TestValidateSchema_simple_fail(t *testing.T) {
	taxonomyFile := "bad_catalog.structs.schema.json"
	g := NewGomegaWithT(t)
	err := ValidateSchema("./" + taxonomyFile)
	g.Expect(err).To(HaveOccurred())
}
