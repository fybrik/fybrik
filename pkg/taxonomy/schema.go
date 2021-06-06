package taxonomy

import (
	"fmt"
	"net/url"
	"path/filepath"

	"emperror.dev/errors"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mesh-for-data/openapi2crd/pkg/generator"
	"github.com/xeipuuv/gojsonschema"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidateSchema validates that the input schema adheres to the requirements defined in
// https://github.com/IBM/the-mesh-for-data/blob/master/config/taxonomy/HOWTO_SPEC.md
func ValidateSchema(path string) error {
	orig := path
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	path = filepath.Clean(path)

	err = ValidateDraft4(path)
	if err != nil {
		return errors.Wrap(err, "Schema is not a valid DRAFT 4 JSON schema")
	}

	err = ValidateStructural(orig)
	if err != nil {
		return errors.Wrap(err, "Schema is not a valid structural schema")
	}

	return err
}

func ValidateDraft4(path string) error {
	schemaLoader := gojsonschema.NewSchemaLoader()
	schemaLoader.Draft = gojsonschema.Draft4
	_, err := schemaLoader.Compile(gojsonschema.NewReferenceLoader("file://" + path))
	return err
}

func ValidateStructural(path string) error {
	crd := &apiextensions.CustomResourceDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name: "objects.group.io",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   "group.io",
			Version: "v1",
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Storage: true,
				},
			},
			Scope: apiextensions.NamespaceScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "Object",
				ListKind: "ObjectList",
				Plural:   "objects",
				Singular: "object",
			},
		},
	}

	loader := &openapi3.Loader{
		IsExternalRefsAllowed: true,
	}

	spec := fmt.Sprintf(`
openapi: 3.0.1
components:
  schemas:
    Object:
      type: object
      properties:
        root:
          $ref: '%s'
`, path)

	specDoc, err := loader.LoadFromDataWithPath([]byte(spec), &url.URL{Path: "."})
	if err != nil {
		return err
	}

	_, err = generator.New().Generate(crd, specDoc.Components.Schemas)
	if err != nil {
		return err
	}

	return nil
}
