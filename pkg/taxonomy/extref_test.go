package taxonomy

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestExternalRef(t *testing.T) {
	loader := &openapi3.Loader{
		IsExternalRefsAllowed: true,
	}
	_, err := loader.LoadFromFile("spec.yaml")
	if err != nil {
		t.Error(err)
	}
}
