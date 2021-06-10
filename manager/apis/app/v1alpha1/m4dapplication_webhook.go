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
	log.Printf("Validating m4dapplication %s for deletion", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *M4DApplication) validateM4DApplication() error {
	var allErrs field.ErrorList
	if err := r.validateM4DApplicationSpec(); err != nil {
		allErrs = append(allErrs, err...)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "app.m4d.ibm.com", Kind: "M4DApplication"},
		r.Name, allErrs)
}

func (r *M4DApplication) validateM4DApplicationSpec() []*field.Error {
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.

	var allErrs []*field.Error
	if len(r.Spec.Data) == 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("data"), r.Spec.Data, "'data' must include at least one element!"))
	}
	specField := field.NewPath("spec").Child("data")
	for i, dataSet := range r.Spec.Data {
		if err := r.validateDataContext(specField.Index(i), &dataSet); err != nil {
			allErrs = append(allErrs, err...)
		}
	}
	return allErrs
}

func (r *M4DApplication) validateDataContext(path *field.Path, dataSet *DataContext) []*field.Error {
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
	case "s3":
		return nil
	case "kafka":
		return nil
	case "jdbc-db2":
		return nil
	case "m4d-arrow-flight":
		return nil
	default:
		return errors.New("Value should be one of these: s3, kafka, jdbc-db2, m4d-arrow-flight")
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
