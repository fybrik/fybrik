// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"os"

	"github.com/rs/zerolog"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fapp "fybrik.io/fybrik/manager/apis/app/v12"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/taxonomy/validate"
)

// FybrikModuleReconciler reconciles a FybrikModule object
type FybrikModuleReconciler struct {
	client.Client
	Name   string
	Log    zerolog.Logger
	Scheme *runtime.Scheme
}

var ModuleTaxonomy = environment.GetDataDir() + "/taxonomy/fybrik_module.json"

const (
	ModuleValidationConditionIndex = 0
	FybrikModuleKind               = "FybrikModule"
)

// Reconcile validates FybrikModule CRD
func (r *FybrikModuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.With().Str(logging.CONTROLLER, FybrikModuleKind).Str("module", req.NamespacedName.String()).Logger()

	// obtain FybrikModule resource
	moduleContext := &fapp.FybrikModule{}
	if err := r.Get(ctx, req.NamespacedName, moduleContext); err != nil {
		log.Warn().Msg("The reconciled object was not found")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// If the object has a scheduled deletion time, update status and return
	if !moduleContext.DeletionTimestamp.IsZero() {
		// The object is being deleted
		return ctrl.Result{}, nil
	}

	observedStatus := moduleContext.Status.DeepCopy()
	moduleVersion := moduleContext.GetGeneration()
	if len(moduleContext.Status.Conditions) == 0 {
		moduleContext.Status.Conditions = []fapp.Condition{{Type: fapp.ValidCondition, Status: corev1.ConditionUnknown, ObservedGeneration: 0}}
	}

	// check if module has been validated before or if validated module is outdated
	condition := moduleContext.Status.Conditions[ModuleValidationConditionIndex]
	if condition.ObservedGeneration != moduleVersion || condition.Status == corev1.ConditionUnknown {
		// do validation on moduleContext
		var err error
		if os.Getenv("ENABLE_WEBHOOKS") != "true" {
			// validation was not done by the webhook
			err = ValidateFybrikModule(moduleContext, ModuleTaxonomy)
		}
		condition.ObservedGeneration = moduleVersion
		// if validation fails
		if err != nil {
			// set error message
			log.Error().Err(err).Msg("Fybrik module validation failed ")
			condition.Message = err.Error()
			condition.Status = corev1.ConditionFalse
		} else {
			condition.Status = corev1.ConditionTrue
			condition.Message = ""
		}
		moduleContext.Status.Conditions[ModuleValidationConditionIndex] = condition
	}

	// Update CRD status in case of change (other than deletion, which was handled separately)
	if moduleContext.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, utils.UpdateStatus(ctx, r.Client, moduleContext, observedStatus)
	}
	return ctrl.Result{}, nil
}

// NewFybrikModuleReconciler creates a new reconciler for FybrikModules
func NewFybrikModuleReconciler(mgr ctrl.Manager, name string) *FybrikModuleReconciler {
	return &FybrikModuleReconciler{
		Client: mgr.GetClient(),
		Name:   name,
		Log:    logging.LogInit(logging.CONTROLLER, name),
		Scheme: mgr.GetScheme(),
	}
}

func ValidateFybrikModule(module *fapp.FybrikModule, taxonomyFile string) error {
	var allErrs []*field.Error

	// Convert Fybrik module Go struct to JSON
	moduleJSON, err := json.Marshal(&module.Spec)
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
		schema.GroupKind{Group: "app.fybrik.io", Kind: FybrikModuleKind},
		module.Name, allErrs)
}

// SetupWithManager registers Module controller
func (r *FybrikModuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fapp.FybrikModule{}).
		Complete(r)
}
