// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"emperror.dev/errors"
	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	"helm.sh/helm/v3/pkg/release"

	"github.com/go-logr/logr"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	"github.com/mesh-for-data/mesh-for-data/pkg/helm"
	corev1 "k8s.io/api/core/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
)

// BlueprintReconciler reconciles a Blueprint object
type BlueprintReconciler struct {
	client.Client
	Name   string
	Log    logr.Logger
	Scheme *runtime.Scheme
	Helmer helm.Interface
}

// Reconcile receives a Blueprint CRD
//nolint:dupl
func (r *BlueprintReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("blueprint", req.NamespacedName)
	var err error

	blueprint := app.Blueprint{}
	if err := r.Get(ctx, req.NamespacedName, &blueprint); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if res, err := r.reconcileFinalizers(&blueprint); err != nil {
		log.V(0).Info("Could not reconcile finalizers: " + err.Error())
		return res, err
	}

	// If the object has a scheduled deletion time, update status and return
	if !blueprint.DeletionTimestamp.IsZero() {
		// The object is being deleted
		log.V(0).Info("Reconcile: Deleting Blueprint " + blueprint.GetName())
		return ctrl.Result{}, nil
	}

	observedStatus := blueprint.Status.DeepCopy()
	log.V(0).Info("Reconcile: Installing/Updating Blueprint " + blueprint.GetName())

	result, err := r.reconcile(ctx, log, &blueprint)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to reconcile blueprint")
	}

	if !equality.Semantic.DeepEqual(&blueprint.Status, observedStatus) {
		if err := r.Client.Status().Update(ctx, &blueprint); err != nil {
			return ctrl.Result{}, errors.WrapWithDetails(err, "failed to update blueprint status", "status", blueprint.Status)
		}
	}

	log.Info("blueprint reconcile cycle completed", "result", result)
	return result, nil
}

// reconcileFinalizers reconciles finalizers for Blueprint
func (r *BlueprintReconciler) reconcileFinalizers(blueprint *app.Blueprint) (ctrl.Result, error) {
	// finalizer
	finalizerName := r.Name + ".finalizer"
	hasFinalizer := ctrlutil.ContainsFinalizer(blueprint, finalizerName)

	// If the object has a scheduled deletion time, delete it and its associated resources
	if !blueprint.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if hasFinalizer { // Finalizer was created when the object was created
			// the finalizer is present - delete the allocated resources
			if err := r.deleteExternalResources(blueprint); err != nil {
				r.Log.V(0).Info("Error while deleting owned resources: " + err.Error())
				return ctrl.Result{}, err
			}
			if r.hasExternalResources(blueprint) {
				return ctrl.Result{RequeueAfter: 2 * time.Second}, errors.NewPlain("helm release uninstall is still in progress")
			}
			// remove the finalizer from the list and update it, because it needs to be deleted together with the object
			ctrlutil.RemoveFinalizer(blueprint, finalizerName)

			if err := r.Update(context.Background(), blueprint); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
	// Make sure this CRD instance has a finalizer
	if !hasFinalizer {
		ctrlutil.AddFinalizer(blueprint, finalizerName)
		if err := r.Update(context.Background(), blueprint); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *BlueprintReconciler) deleteExternalResources(blueprint *app.Blueprint) error {
	errs := make([]string, 0)
	for _, step := range blueprint.Spec.Flow.Steps {
		releaseName := utils.GetReleaseName(blueprint.Labels[app.ApplicationNameLabel], blueprint.Labels[app.ApplicationNamespaceLabel], step)
		if rel, errStatus := r.Helmer.Status(blueprint.Namespace, releaseName); errStatus != nil || rel == nil {
			continue
		}
		if _, err := r.Helmer.Uninstall(blueprint.Namespace, releaseName); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "; "))
}

func (r *BlueprintReconciler) hasExternalResources(blueprint *app.Blueprint) bool {
	for _, step := range blueprint.Spec.Flow.Steps {
		releaseName := utils.GetReleaseName(blueprint.Labels[app.ApplicationNameLabel], blueprint.Labels[app.ApplicationNamespaceLabel], step)
		if rel, errStatus := r.Helmer.Status(blueprint.Namespace, releaseName); errStatus == nil && rel != nil {
			return true
		}
	}
	return false
}

func (r *BlueprintReconciler) applyChartResource(log logr.Logger, chartSpec app.ChartSpec, args map[string]interface{}, blueprint *app.Blueprint, releaseName string) (ctrl.Result, error) {
	log.Info(fmt.Sprintf("--- Chart Ref ---\n\n%v\n\n", chartSpec.Name))
	kubeNamespace := blueprint.Namespace

	args = CopyMap(args)
	for k, v := range chartSpec.Values {
		SetMapField(args, k, v)
	}
	labels := map[string]string{
		app.ApplicationNamespaceLabel: blueprint.Labels[app.ApplicationNamespaceLabel],
		app.ApplicationNameLabel:      blueprint.Labels[app.ApplicationNameLabel],
		app.BlueprintNamespaceLabel:   kubeNamespace,
		app.BlueprintNameLabel:        blueprint.Name,
	}
	SetMapField(args, "labels", labels)
	nbytes, _ := yaml.Marshal(args)
	log.Info(fmt.Sprintf("--- Values.yaml ---\n\n%s\n\n", nbytes))

	err := r.Helmer.ChartPull(chartSpec.Name)
	if err != nil {
		return ctrl.Result{}, errors.WithMessage(err, chartSpec.Name+": failed chart pull")
	}
	chart, err := r.Helmer.ChartLoad(chartSpec.Name)
	if err != nil {
		return ctrl.Result{}, errors.WithMessage(err, chartSpec.Name+": failed chart load")
	}

	rel, err := r.Helmer.Status(kubeNamespace, releaseName)
	if err == nil && rel != nil {
		rel, err = r.Helmer.Upgrade(chart, kubeNamespace, releaseName, args)
		if err != nil {
			return ctrl.Result{}, errors.WithMessage(err, chartSpec.Name+": failed upgrade")
		}
	} else {
		rel, err = r.Helmer.Install(chart, kubeNamespace, releaseName, args)
		if err != nil {
			return ctrl.Result{}, errors.WithMessage(err, chartSpec.Name+": failed install")
		}
	}
	log.Info(fmt.Sprintf("--- Release Status ---\n\n%s\n\n", rel.Info.Status))
	return ctrl.Result{}, nil
}

// CopyMap copies a map
func CopyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = CopyMap(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}

// SetMapField updates a map
func SetMapField(obj map[string]interface{}, k string, v interface{}) bool {
	components := strings.Split(k, ".")
	for n, component := range components {
		if n == len(components)-1 {
			obj[component] = v
		} else {
			m, ok := obj[component]
			if !ok {
				m := make(map[string]interface{})
				obj[component] = m
				obj = m
			} else if obj, ok = m.(map[string]interface{}); !ok {
				return false
			}
		}
	}
	return true
}

func (r *BlueprintReconciler) reconcile(ctx context.Context, log logr.Logger, blueprint *app.Blueprint) (ctrl.Result, error) {
	// Gather all templates and process them into a list of resources to apply
	// force-update if the blueprint spec is different
	updateRequired := blueprint.Status.ObservedGeneration != blueprint.GetGeneration()
	blueprint.Status.ObservedGeneration = blueprint.GetGeneration()
	// reset blueprint state
	blueprint.Status.ObservedState.Ready = false
	blueprint.Status.ObservedState.Error = ""
	blueprint.Status.ObservedState.DataAccessInstructions = ""
	if blueprint.Status.Releases == nil {
		blueprint.Status.Releases = map[string]int64{}
	}

	// count the overall number of Helm releases and how many of them are ready
	numReleases, numReady := 0, 0

	for _, step := range blueprint.Spec.Flow.Steps {
		templateName := step.Template
		templateSpec, err := findComponentTemplateByName(blueprint.Spec.Templates, templateName)
		if err != nil {
			return ctrl.Result{}, errors.WithMessage(err, "Blueprint step uses non-existing template")
		}

		// We only orchestrate M4DModule templates
		if templateSpec.Kind != "M4DModule" {
			continue
		}

		// Get arguments by type
		var args map[string]interface{}
		args, err = utils.StructToMap(step.Arguments)
		if err != nil {
			return ctrl.Result{}, errors.WithMessage(err, "Blueprint step arguments are invalid")
		}
		releaseName := utils.GetReleaseName(blueprint.Labels[app.ApplicationNameLabel], blueprint.Labels[app.ApplicationNamespaceLabel], step)
		log.V(0).Info("Release name: " + releaseName)
		numReleases++
		// check the release status
		rel, err := r.Helmer.Status(blueprint.Namespace, releaseName)
		// unexisting release or a failed release - re-apply the chart
		if updateRequired || err != nil || rel == nil || rel.Info.Status == release.StatusFailed {
			// Process templates with arguments
			chart := templateSpec.Chart
			if _, err := r.applyChartResource(log, chart, args, blueprint, releaseName); err != nil {
				blueprint.Status.ObservedState.Error += errors.Wrap(err, "ChartDeploymentFailure: ").Error() + "\n"
			}
		} else if rel.Info.Status == release.StatusDeployed {
			if len(step.Arguments.Read) > 0 {
				blueprint.Status.ObservedState.DataAccessInstructions += rel.Info.Notes
			}
			status, errMsg := r.checkReleaseStatus(releaseName, blueprint.Namespace)
			if status == corev1.ConditionFalse {
				blueprint.Status.ObservedState.Error += "ResourceAllocationFailure: " + errMsg + "\n"
			} else if status == corev1.ConditionTrue {
				numReady++
			}
		}
		blueprint.Status.Releases[releaseName] = blueprint.Status.ObservedGeneration
	}
	// clean-up
	for release, version := range blueprint.Status.Releases {
		if version != blueprint.Status.ObservedGeneration {
			_, err := r.Helmer.Uninstall(blueprint.Namespace, release)
			if err != nil {
				log.V(0).Info("Error uninstalling release " + release + " : " + err.Error())
			} else {
				delete(blueprint.Status.Releases, release)
			}
		}
	}
	// check if all releases reached the ready state
	if numReady == numReleases {
		// all modules have been orhestrated successfully - the data is ready for use
		blueprint.Status.ObservedState.Ready = true
		return ctrl.Result{}, nil
	}

	// the status is unknown yet - continue polling
	if blueprint.Status.ObservedState.Error == "" {
		return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
	}
	return ctrl.Result{}, nil
}

func findComponentTemplateByName(templates []app.ComponentTemplate, name string) (*app.ComponentTemplate, error) {
	// TODO(roee.shlomo): BlueprintSpec#Templates should probably be a map from name to the module spec. Then we can remove this function.
	for _, template := range templates {
		if template.Name == name {
			return &template, nil
		}
	}
	return nil, fmt.Errorf("template %s not found", name)
}

// NewBlueprintReconciler creates a new reconciler for Blueprint resources
func NewBlueprintReconciler(mgr ctrl.Manager, name string, helmer helm.Interface) *BlueprintReconciler {
	return &BlueprintReconciler{
		Client: mgr.GetClient(),
		Name:   name,
		Log:    ctrl.Log.WithName("controllers").WithName(name),
		Scheme: mgr.GetScheme(),
		Helmer: helmer,
	}
}

// SetupWithManager registers Blueprint controller
func (r *BlueprintReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&app.Blueprint{}).
		Complete(r)
}

func (r *BlueprintReconciler) getExpectedResults(kind string) (*app.ResourceStatusIndicator, error) {
	// Assumption: specification for each resource kind is done in one place.
	ctx := context.Background()

	var moduleList app.M4DModuleList
	if err := r.List(ctx, &moduleList); err != nil {
		return nil, err
	}
	for _, module := range moduleList.Items {
		for _, res := range module.Spec.StatusIndicators {
			if res.Kind == kind {
				return res.DeepCopy(), nil
			}
		}
	}
	return nil, nil
}

// checkResourceStatus returns the computed state and an error message if exists
func (r *BlueprintReconciler) checkResourceStatus(res *unstructured.Unstructured) (corev1.ConditionStatus, string) {
	// get indications how to compute the resource status based on a module spec
	expected, err := r.getExpectedResults(res.GetKind())
	if err != nil {
		// Could not retrieve the list of modules, will retry later
		return corev1.ConditionUnknown, ""
	}
	if expected == nil {
		// use kstatus to compute the status of the resources for them the expected results have not been specified
		// Current status of a deployed release indicates that the resource has been successfully reconciled
		// Failed status indicates a failure
		computedResult, err := kstatus.Compute(res)
		if err != nil {
			r.Log.V(0).Info("Error computing the status of " + res.GetKind() + " : " + err.Error())
			return corev1.ConditionUnknown, ""
		}
		switch computedResult.Status {
		case kstatus.FailedStatus:
			return corev1.ConditionFalse, computedResult.Message
		case kstatus.CurrentStatus:
			return corev1.ConditionTrue, ""
		default:
			return corev1.ConditionUnknown, ""
		}
	}
	// use expected values to compute the status
	if r.matchesCondition(res, expected.SuccessCondition) {
		return corev1.ConditionTrue, ""
	}
	if r.matchesCondition(res, expected.FailureCondition) {
		return corev1.ConditionFalse, getErrorMessage(res, expected.ErrorMessage)
	}
	return corev1.ConditionUnknown, ""
}

func getErrorMessage(res *unstructured.Unstructured, fieldPath string) string {
	// convert the unstructured data to Labels interface to use Has and Get methods for retrieving data
	labelsImpl := utils.UnstructuredAsLabels{Data: res}
	if !labelsImpl.Has(fieldPath) {
		return ""
	}
	return labelsImpl.Get(fieldPath)
}

func (r *BlueprintReconciler) matchesCondition(res *unstructured.Unstructured, condition string) bool {
	selector, err := labels.Parse(condition)
	if err != nil {
		r.Log.V(0).Info("condition " + condition + "failed to parse: " + err.Error())
		return false
	}
	// get selector requirements, 'selectable' property is ignored
	requirements, _ := selector.Requirements()
	// convert the unstructured data to Labels interface to leverage the package capability of parsing and evaluating conditions
	labelsImpl := utils.UnstructuredAsLabels{Data: res}
	for _, req := range requirements {
		if !req.Matches(labelsImpl) {
			return false
		}
	}
	return true
}

func (r *BlueprintReconciler) checkReleaseStatus(releaseName string, namespace string) (corev1.ConditionStatus, string) {
	// get all resources for the given helm release in their current state
	resources, err := r.Helmer.GetResources(namespace, releaseName)
	if err != nil {
		r.Log.V(0).Info("Error getting resources: " + err.Error())
		return corev1.ConditionUnknown, ""
	}
	// return True if all resources are ready, False - if any resource failed, Unknown - otherwise
	numReady := 0
	for _, res := range resources {
		state, errMsg := r.checkResourceStatus(res)
		r.Log.V(0).Info("Status of " + res.GetKind() + " " + res.GetName() + " is " + string(state))
		if state == corev1.ConditionFalse {
			return state, errMsg
		}
		if state == corev1.ConditionTrue {
			numReady++
		}
	}
	if numReady == len(resources) {
		return corev1.ConditionTrue, ""
	}
	return corev1.ConditionUnknown, ""
}
