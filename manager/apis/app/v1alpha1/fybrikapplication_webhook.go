// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"errors"
	log "log"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *FybrikApplication) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,admissionReviewVersions=v1;v1beta1,sideEffects=None,path=/validate-app-fybrik-io-v1alpha1-fybrikapplication,mutating=false,failurePolicy=fail,groups=app.fybrik.io,resources=fybrikapplications,versions=v1alpha1,name=vfybrikapplication.kb.io

var _ webhook.Validator = &FybrikApplication{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *FybrikApplication) ValidateCreate() error {
	log.Printf("Validating fybrikapplication %s for creation", r.Name)
	return r.validateFybrikApplication()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *FybrikApplication) ValidateUpdate(old runtime.Object) error {
	log.Printf("Validating fybrikapplication %s for update", r.Name)

	return r.validateFybrikApplication()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *FybrikApplication) ValidateDelete() error {
	log.Printf("Validating fybrikapplication %s for deletion", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *FybrikApplication) validateFybrikApplication() error {
	var allErrs field.ErrorList
	if err := r.validateFybrikApplicationSpec(); err != nil {
		allErrs = append(allErrs, err...)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "app.fybrik.io", Kind: "FybrikApplication"},
		r.Name, allErrs)
}

func (r *FybrikApplication) validateFybrikApplicationSpec() []*field.Error {
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.

	var allErrs []*field.Error
	specField := field.NewPath("spec").Child("data")
	for i, dataSet := range r.Spec.Data {
		// To avoid aliasing issue due to the fact the address of dataSet variable is passed
		// to validateDataContext function in each loop iteration, a new temporary
		// variable is created and used in each iteration.
		dataSetTemp := dataSet
		if err := r.validateDataContext(specField.Index(i), &dataSetTemp); err != nil {
			allErrs = append(allErrs, err...)
		}
	}
	return allErrs
}

func (r *FybrikApplication) validateDataContext(path *field.Path, dataSet *DataContext) []*field.Error {
	var allErrs []*field.Error
	interfacePath := path.Child("Requirements", "Interface")
	if err := validateProtocol(dataSet.Requirements.Interface.Protocol); err != nil {
		allErrs = append(allErrs, field.Invalid(interfacePath.Child("Protocol"), &dataSet.Requirements.Interface.Protocol, err.Error()))
	}
	if err := validateDataFormat(dataSet.Requirements.Interface.DataFormat); err != nil {
		allErrs = append(allErrs, field.Invalid(interfacePath.Child("DataFormat"), &dataSet.Requirements.Interface.DataFormat, err.Error()))
	}
	return allErrs
}

func validateProtocol(protocol string) error {
	switch protocol {
	case "s3", "kafka", "jdbc-db2", "fybrik-arrow-flight":
		return nil
	default:
		return errors.New("Value should be one of these: s3, kafka, jdbc-db2, fybrik-arrow-flight")
	}
}

func validateDataFormat(format string) error {
	switch format {
	case "parquet", "table", "csv", "json", "avro", "orc", "binary", "arrow":
		return nil
	default:
		return errors.New("Value should be one of these: parquet, table, csv, json, avro, orc, binary, arrow")
	}
}
