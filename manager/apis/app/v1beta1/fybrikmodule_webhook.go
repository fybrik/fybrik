// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1beta1

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

func (r *FybrikModule) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,admissionReviewVersions=v1;v1beta1,sideEffects=None,path=/validate-app-fybrik-io-v1beta1-fybrikmodule,mutating=false,failurePolicy=fail,groups=app.fybrik.io,resources=fybrikmodules,versions=v1beta1,name=vfybrikmodule.kb.io

var _ webhook.Validator = &FybrikModule{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *FybrikModule) ValidateCreate() error {
	taxonomyFile := environment.GetDataDir() + "/taxonomy/fybrik_module.json"
	return r.ValidateFybrikModule(taxonomyFile)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *FybrikModule) ValidateUpdate(old runtime.Object) error {
	return r.ValidateCreate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *FybrikModule) ValidateDelete() error {
	return nil
}

func (r *FybrikModule) ValidateFybrikModule(taxonomyFile string) error {
	var allErrs []*field.Error

	// Convert FybrikModule Go struct to JSON
	moduleJSON, err := json.Marshal(&r.Spec)
	if err != nil {
		return err
	}

	// Validate Fybrik module against taxonomy
	allErrs, err = validate.TaxonomyCheck(moduleJSON, taxonomyFile)
	if err != nil {
		return err
	}

	// Return any error
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "app.fybrik.io", Kind: "FybrikModule"},
		r.Name, allErrs)
}
