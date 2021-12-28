// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	taxonomyio "fybrik.io/fybrik/pkg/taxonomy/io"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsStructuralSchema returns an error if the input file is not a valid structural schema
func IsStructuralSchema(path string) error {
	document, err := taxonomyio.ReadDocumentFromFile(path)
	if err != nil {
		return err
	}

	crd := &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
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
			Validation: &apiextensions.CustomResourceValidation{
				OpenAPIV3Schema: document.ToFlatJSONSchemaProps(),
			},
			Scope: apiextensions.NamespaceScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:     "Object",
				ListKind: "ObjectList",
				Plural:   "objects",
				Singular: "object",
			},
		},
		Status: apiextensions.CustomResourceDefinitionStatus{
			StoredVersions: []string{"v1"},
		},
	}

	errorList := validation.ValidateCustomResourceDefinition(crd, apiextensionsv1.SchemeGroupVersion)
	if len(errorList) > 0 {
		return errorList.ToAggregate()
	}

	return nil
}
