// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/scale/scheme"

	"fybrik.io/fybrik/pkg/slices"
	taxonomyio "fybrik.io/fybrik/pkg/taxonomy/io"
	"fybrik.io/fybrik/pkg/taxonomy/model"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

var (
	matcherTaxonomy   = regexp.MustCompile(`(taxonomy.json#/definitions/[a-zA-Z0-9]+)`)
	matcherImplements = regexp.MustCompile(`(Implements::[a-zA-Z0-9]+)`)
	definitionsStr    = "#/definitions"
)

// LoadCRDs loads all CustomResourceDefinition resources from a directory (glob)
func LoadCRDs(dirpath string) ([]*apiextensions.CustomResourceDefinition, error) {
	files, err := filepath.Glob(path.Join(dirpath, "*"))
	if err != nil {
		return nil, err
	}

	// Glob ignores file system errors, so check the supplied path when there
	// are no results. When it is a file, treat it like a single result from
	// Glob. When it does not exist, return an error.
	if len(files) == 0 {
		info, err := os.Stat(dirpath)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			files = append(files, dirpath)
		}
	}

	resources := []*apiextensions.CustomResourceDefinition{}

	for _, file := range files {
		// Read file
		filecontent, err := ioutil.ReadFile(filepath.Clean(file))
		if err != nil {
			return nil, err
		}

		// Split if multiple YAML documents are defined in the file
		fileDocuments, err := loadYAMLDocuments(filecontent)
		if err != nil {
			return nil, err
		}

		for _, document := range fileDocuments {
			crd, err := decodeCRD(document)
			if err != nil {
				return nil, err
			}
			if crd != nil {
				resources = append(resources, crd)
			}
		}
	}

	return resources, nil
}

func loadYAMLDocuments(filecontent []byte) ([][]byte, error) {
	dec := yaml.NewDecoder(bytes.NewReader(filecontent))

	var res [][]byte
	for {
		var value interface{}
		err := dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		valueBytes, err := yaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		res = append(res, valueBytes)
	}
	return res, nil
}

func decodeCRD(content []byte) (*apiextensions.CustomResourceDefinition, error) {
	sch := runtime.NewScheme()
	_ = scheme.AddToScheme(sch)
	_ = apiextensions.AddToScheme(sch)
	_ = apiextensionsv1.AddToScheme(sch)
	_ = apiextensionsv1.RegisterConversions(sch)
	_ = apiextensionsv1beta1.AddToScheme(sch)
	_ = apiextensionsv1beta1.RegisterConversions(sch)

	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode
	obj, _, err := decode(content, nil, nil)
	if err != nil {
		// Not a Kubernetes resource
		if strings.HasPrefix(err.Error(), "Object 'Kind' is missing in ") {
			return nil, nil
		}
		// Not a kind from the above registered schemas
		if strings.HasPrefix(err.Error(), "no kind ") {
			return nil, nil
		}
		return nil, err
	}

	if obj.GetObjectKind().GroupVersionKind().Kind != "CustomResourceDefinition" {
		return nil, nil
	}

	crd := &apiextensions.CustomResourceDefinition{}
	err = sch.Convert(obj, crd, nil)
	if err != nil {
		return nil, err
	}
	return crd, err
}

func GenerateValidationObjectFromCRDs(inputDirOrFile, outputFilepath, title string) error {
	crds, err := LoadCRDs(inputDirOrFile)
	if err != nil {
		return err
	}

	// create output directory if needed
	/*
		err = os.MkdirAll(filepath.Clean(outputDir), os.ModePerm)
		if err != nil {
			return err
		}
	*/

	for _, crd := range crds {
		err = generateFile(crd, outputFilepath, title)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateFile(crd *apiextensions.CustomResourceDefinition, outputFilepath, title string) error {
	document, err := processCRD(crd, title)
	if err != nil {
		return err
	}

	if document != nil {
		err = taxonomyio.WriteDocumentToFile(document, outputFilepath)
		if err != nil {
			return err
		}
	}

	return nil
}

func getOpenAPIV3Schema(crd *apiextensions.CustomResourceDefinition) (*apiextensions.JSONSchemaProps, error) {
	for _, version := range crd.Spec.Versions {
		if !version.Storage {
			continue
		}

		// Find schema
		validation := version.Schema
		if validation == nil {
			// Fallback to resource level schema
			validation = crd.Spec.Validation
		}

		if validation == nil {
			return nil, errors.New("missing validation field in input CRD")
		}
		schema := validation.OpenAPIV3Schema
		return schema, nil
	}
	return nil, errors.New("missing storage version in CRD")
}

func processCRD(crd *apiextensions.CustomResourceDefinition, title string) (*model.Schema, error) {
	schema, err := getOpenAPIV3Schema(crd)
	if err != nil {
		return nil, err
	}

	jsonSchema := buildJSONSchema(schema)
	if jsonSchema == nil {
		return nil, nil
	}

	document := &model.Schema{
		Version:     "http://json-schema.org/draft-04/schema#",
		Definitions: make(map[string]*model.SchemaRef),
		Title:       title,
		Type:        jsonSchema.Type,
	}

	// Move top level properties to definitions
	for k, v := range jsonSchema.Schema.Properties["spec"].Schema.Properties {
		document.Definitions[k] = v
		if groups := matcherImplements.FindStringSubmatch(v.Description); len(groups) > 1 {
			publicName := strings.Split(groups[1], "::")[1]
			if _, ok := document.Definitions[publicName]; !ok {
				document.Definitions[publicName] = &model.SchemaRef{
					Ref: definitionsStr + "/" + k,
				}
			}
		}
	}
	return document, nil
}

func buildJSONSchema(props *apiextensions.JSONSchemaProps) *model.SchemaRef {
	if props == nil {
		return nil
	}

	// Add a reference to the taxonomy object
	if groups := matcherTaxonomy.FindStringSubmatch(props.Description); len(groups) > 1 {
		return &model.SchemaRef{
			Ref: groups[1],
		}
	}

	out := &model.SchemaRef{
		Schema: model.Schema{
			Type:        props.Type,
			Description: props.Description,
			Required:    []string{},
		},
	}

	if props.Properties != nil {
		for k, v := range props.Properties {
			schema := buildJSONSchema(&v)
			if schema != nil || true {
				if out.Properties == nil {
					out.Properties = model.Schemas{}
				}
				out.Properties[k] = schema
				if slices.ContainsString(k, props.Required) {
					out.Required = append(out.Required, k)
				}
			}
		}
	}

	if props.Items != nil {
		schema := buildJSONSchema(props.Items.Schema)
		if schema != nil {
			out.Items = schema
		}
	}

	if props.AdditionalProperties != nil {
		schema := buildJSONSchema(props.AdditionalProperties.Schema)
		if schema != nil {
			out.AdditionalProperties = &model.AdditionalPropertiesType{
				Schema: schema,
			}
		}
	}
	return out
}
