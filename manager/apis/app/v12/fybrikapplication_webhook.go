// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v12

import (
	"encoding/json"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"fybrik.io/fybrik/pkg/environment"
	validate "fybrik.io/fybrik/pkg/taxonomy/validate"
)

var taxonomyFilePath = environment.GetDataDir() + "/taxonomy/fybrik_application.json"

func (r *FybrikApplication) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,admissionReviewVersions=v1;v1beta1,sideEffects=None,path=/validate-app-fybrik-io-v12-fybrikapplication,mutating=false,failurePolicy=fail,groups=app.fybrik.io,resources=fybrikapplications,versions=v12,name=vfybrikapplication.kb.io

var _ webhook.Validator = &FybrikApplication{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *FybrikApplication) ValidateCreate() error {
	taxonomyFile := taxonomyFilePath
	return r.ValidateFybrikApplication(taxonomyFile)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *FybrikApplication) ValidateUpdate(old runtime.Object) error {
	taxonomyFile := taxonomyFilePath
	return r.ValidateFybrikApplication(taxonomyFile)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *FybrikApplication) ValidateDelete() error {
	return nil
}

func (r *FybrikApplication) ValidateFybrikApplication(taxonomyFile string) error {
	var allErrs []*field.Error

	// Convert Fybrik application Go struct to JSON
	applicationJSON, err := json.Marshal(&r.Spec)
	if err != nil {
		return err
	}
	// Validate Fybrik application against taxonomy
	allErrs, err = validate.TaxonomyCheck(applicationJSON, taxonomyFile)
	if err != nil {
		return err
	}

	// Return any error
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "app.fybrik.io", Kind: "FybrikApplication"},
		r.Name, allErrs)
}
