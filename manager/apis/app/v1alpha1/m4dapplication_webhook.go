// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"encoding/json"
	log "log"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *M4DApplication) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,admissionReviewVersions=v1;v1beta1,sideEffects=None,path=/validate-app-m4d-ibm-com-v1alpha1-m4dapplication,mutating=false,failurePolicy=fail,groups=app.m4d.ibm.com,resources=m4dapplications,versions=v1alpha1,name=vm4dapplication.kb.io

var _ webhook.Validator = &M4DApplication{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *M4DApplication) ValidateCreate() error {
	log.Printf("Validating m4dapplication %s for creation", r.Name)
	return r.validateM4DApplication()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *M4DApplication) ValidateUpdate(old runtime.Object) error {
	log.Printf("Validating m4dapplication %s for update", r.Name)

	return r.validateM4DApplication()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *M4DApplication) ValidateDelete() error {
	return nil
}

func (r *M4DApplication) validateM4DApplication() error {
	var allErrs []*field.Error

	// Convert M4D application Go struct to JSON
	applicationJSON, err := json.Marshal(r)
	if err != nil {
		return err
	}

	// Validate against strict taxonomy
	path, err := filepath.Abs("/tmp/taxonomy/application.values.schema.json")
	if err != nil {
		return err
	}

	taxonomyLoader := gojsonschema.NewReferenceLoader("file://" + path)
	documentLoader := gojsonschema.NewStringLoader(string(applicationJSON))
	result, err := gojsonschema.Validate(taxonomyLoader, documentLoader)
	if err != nil {
		return err
	}

	if result.Valid() {
		log.Printf("This M4D application is valid\n")
	} else {
		log.Printf("This M4D application is not valid. see errors :\n")
		for _, desc := range result.Errors() {
			log.Printf("- %s\n", desc)
			allErrs = append(allErrs, field.Invalid(field.NewPath(desc.Field()), desc.Value(), desc.Description()))
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "app.m4d.ibm.com", Kind: "M4DApplication"},
		r.Name, allErrs)
}
