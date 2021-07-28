// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"encoding/json"
	log "log"

	validate "github.com/mesh-for-data/mesh-for-data/pkg/taxonomy/validate"
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
	taxonomyFile := "/tmp/taxonomy/application.values.schema.json"
	return r.validateM4DApplication(taxonomyFile)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *M4DApplication) ValidateUpdate(old runtime.Object) error {
	log.Printf("Validating m4dapplication %s for update", r.Name)
	taxonomyFile := "/tmp/taxonomy/application.values.schema.json"
	return r.validateM4DApplication(taxonomyFile)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *M4DApplication) ValidateDelete() error {
	return nil
}

func (r *M4DApplication) validateM4DApplication(taxonomyFile string) error {
	var allErrs []*field.Error

	// Convert M4D application Go struct to JSON
	applicationJSON, err := json.Marshal(r)
	if err != nil {
		return err
	}

	// Validate M4D application against taxonomy
	allErrs = validate.TaxonomyCheck(applicationJSON, taxonomyFile, "M4D application")

	// Return any error
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "app.m4d.ibm.com", Kind: "M4DApplication"},
		r.Name, allErrs)
}
