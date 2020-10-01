// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"helm.sh/helm/v3/pkg/release"

	"emperror.dev/errors"

	"github.com/go-logr/logr"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	"github.com/ibm/the-mesh-for-data/pkg/helm"
	corev1 "k8s.io/api/core/v1"
	kstatus "sigs.k8s.io/kustomize/kstatus/status"
)

// BlueprintReconciler reconciles a Blueprint object
type BlueprintReconciler struct {
	client.Client
	Name   string
	Log    logr.Logger
	Scheme *runtime.Scheme
	Helmer helm.Interface
}

// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=blueprints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=blueprints/status,verbs=get;update;patch

// Reconcile receives a Blueprint CRD
func (r *BlueprintReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("blueprint", req.NamespacedName)
	var err error

	blueprint := app.Blueprint{}
	if err := r.Get(ctx, req.NamespacedName, &blueprint); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	observedStatus := blueprint.Status.DeepCopy()

	if blueprint.DeletionTimestamp != nil {
		log.V(0).Info("Reconcile: Deleting Blueprint " + blueprint.GetName())
	} else {
		log.V(0).Info("Reconcile: Installing/Updating Blueprint " + blueprint.GetName())
	}

	result, err := r.reconcile(ctx, log, &blueprint)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to reconcile blueprint")
	}

	if blueprint.DeletionTimestamp == nil {
		if !equality.Semantic.DeepEqual(blueprint.Status, observedStatus) {
			if err := r.Client.Status().Update(ctx, &blueprint); err != nil {
				return ctrl.Result{}, errors.WrapWithDetails(err, "failed to update blueprint status", "status", blueprint.Status)
			}
		}
	}

	log.Info("blueprint reconcile cycle completed", "result", result)
	return result, nil
}

func (r *BlueprintReconciler) deleteChartResource(ctx context.Context, log logr.Logger, kubeNamespace string, releaseName string) (ctrl.Result, error) {
	rel, err := r.Helmer.Status(kubeNamespace, releaseName)
	if err == nil && rel != nil {
		_, _ = r.Helmer.Uninstall(kubeNamespace, releaseName)
	}

	return ctrl.Result{}, nil
}

func (r *BlueprintReconciler) applyChartResource(ctx context.Context, log logr.Logger, ref string, vals map[string]interface{}) (ctrl.Result, error) {
	log.Info(fmt.Sprintf("--- Chart Ref ---\n\n%v\n\n", ref))

	nbytes, _ := yaml.Marshal(vals)
	log.Info(fmt.Sprintf("--- Values.yaml ---\n\n%s\n\n", nbytes))

	// TODO: should change to use an ImagePullSecret referenced from the M4DModule resource
	hostname := os.Getenv("DOCKER_HOSTNAME")
	username := os.Getenv("DOCKER_USERNAME")
	password := os.Getenv("DOCKER_PASSWORD")
	insecure, _ := strconv.ParseBool(os.Getenv("DOCKER_INSECURE"))
	if username != "" && password != "" {
		err := r.Helmer.RegistryLogin(hostname, username, password, insecure)
		if err != nil {
			return ctrl.Result{}, errors.WithMessage(err, ref+": failed chart pull")
		}
	}

	err := r.Helmer.ChartPull(ref)
	if err != nil {
		return ctrl.Result{}, errors.WithMessage(err, ref+": failed chart pull")
	}
	chart, err := r.Helmer.ChartLoad(ref)
	if err != nil {
		return ctrl.Result{}, errors.WithMessage(err, ref+": failed chart load")
	}

	kubeNamespace := vals["metadata"].(map[string]interface{})["namespace"].(string)
	// we add the "r" character at the beginning of the release name, since it must begin with an alphabetic character
	releaseName := "r" + vals["metadata"].(map[string]interface{})["name"].(string)

	rel, err := r.Helmer.Status(kubeNamespace, releaseName)
	if err == nil && rel != nil {
		rel, err = r.Helmer.Upgrade(chart, kubeNamespace, releaseName, vals)
		if err != nil {
			return ctrl.Result{}, errors.WithMessage(err, ref+": failed upgrade")
		}
	} else {
		rel, err = r.Helmer.Install(chart, kubeNamespace, releaseName, vals)
		if err != nil {
			return ctrl.Result{}, errors.WithMessage(err, ref+": failed install")
		}
	}
	log.Info(fmt.Sprintf("--- Release Status ---\n\n%s\n\n", rel.Info.Status))
	return ctrl.Result{}, nil
}

func (r *BlueprintReconciler) reconcile(ctx context.Context, log logr.Logger, blueprint *app.Blueprint) (ctrl.Result, error) {
	// Gather all templates and process them into a list of resources to apply
	// force-update if the blueprint spec is different
	updateRequired := blueprint.Status.ObservedGeneration != blueprint.GetGeneration()
	blueprint.Status.ObservedGeneration = blueprint.GetGeneration()
	// reset conditions
	blueprint.Status.Conditions = make([]app.Condition, 0)
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

		// TODO: current Copy modules do not expect the "copy" key.
		//       once updated, these lines could be dropped
		if step.Arguments.Flow == app.Copy {
			args = args["copy"].(map[string]interface{})
		}

		// Add metadata arguments (namespace, name, labels, etc.)
		args["metadata"] = map[string]interface{}{
			"name":      utils.Hash(step.Name, 20),
			"namespace": blueprint.Namespace,
			"labels":    blueprint.Labels,
		}
		releaseName := "r" + args["metadata"].(map[string]interface{})["name"].(string)
		log.V(0).Info("Release name: " + releaseName)
		// check if the blueprint is about to be deleted
		if blueprint.DeletionTimestamp != nil {
			_, _ = r.deleteChartResource(ctx, log, blueprint.Namespace, releaseName)
			continue
		}

		numReleases++
		//check the release status
		rel, err := r.Helmer.Status(blueprint.Namespace, releaseName)
		// unexisting release or a failed release - re-apply the chart
		if updateRequired || err != nil || rel == nil || rel.Info.Status == release.StatusFailed {
			// Process templates with arguments
			for _, resource := range templateSpec.Resources {
				if _, err := r.applyChartResource(ctx, log, resource, args); err != nil {
					utils.UpdateCondition(&blueprint.Status.Conditions, app.FailureCondition, "ChartDeploymentFailure", err.Error())
				}
			}
		} else if rel.Info.Status == release.StatusDeployed {
			// TODO: add release notes of the read module to the status
			if args["flow"] == "read" {
				log.V(0).Info(rel.Info.Notes)
			}
			status, errMsg := r.checkStatus(releaseName, blueprint.Namespace)
			if status == corev1.ConditionFalse {
				utils.UpdateCondition(&blueprint.Status.Conditions, app.FailureCondition, "ResourceAllocationFailure", errMsg)
			} else if status == corev1.ConditionTrue {
				numReady++
			}
		}
	}
	// check if all releases reached the ready state
	if numReady == numReleases {
		// raise a ready condition
		utils.UpdateCondition(&blueprint.Status.Conditions, app.ReadyCondition, "", "")
		return ctrl.Result{}, nil
	}

	// the status is unknown yet - continue polling
	if !utils.HasCondition(blueprint.Status.Conditions, app.FailureCondition) {
		return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
	}
	return ctrl.Result{}, nil
}

// checkResourceStatus returns the computed state and an error message if exists
func (r *BlueprintReconciler) checkResourceStatus(res *unstructured.Unstructured) (corev1.ConditionStatus, string) {
	obj := res.UnstructuredContent()
	kind, ok := obj["kind"].(string)
	if !ok {
		// invalid resource
		return corev1.ConditionUnknown, ""
	}

	// get indications how to compute the resource status based on a module spec
	expected, _ := r.getExpectedResults(kind)
	if len(expected) == 0 {
		// no information available - use kstatus instead
		computedResult, _ := kstatus.Compute(res)
		switch computedResult.Status {
		case kstatus.FailedStatus:
			return corev1.ConditionFalse, computedResult.Message
		case kstatus.CurrentStatus: // Current status of a deployed release in addition to waiting for complete deployment should be enough
			return corev1.ConditionTrue, ""
		default:
			return corev1.ConditionUnknown, ""
		}
	}
	// use expected values to compute the status
	errorMsg := ""
	state := corev1.ConditionUnknown
	for _, expectedResult := range expected {
		// TODO: handle an hierarchical path
		actualValue, found, err := unstructured.NestedString(obj, "status", expectedResult.Path)
		if !found || err != nil {
			continue
		}
		// find error message
		if expectedResult.State == string(app.ErrorState) {
			errorMsg = actualValue
			continue
		}
		if expectedResult.Value != "" && actualValue == expectedResult.Value {
			// set state: ready or failed
			if expectedResult.State == string(app.ReadyState) {
				state = corev1.ConditionTrue
			} else if expectedResult.State == string(app.FailedState) {
				state = corev1.ConditionFalse
			}
		}
	}
	return state, errorMsg
}

func (r *BlueprintReconciler) checkStatus(releaseName string, namespace string) (corev1.ConditionStatus, string) {
	resources, err := r.Helmer.GetResources(namespace, releaseName)
	if err != nil {
		r.Log.V(0).Info("Error getting resources: " + err.Error())
		return corev1.ConditionUnknown, ""
	}
	// return True if all resources are ready, False - if any resource failed, Unknown - otherwise
	numReady := 0
	for _, res := range resources {
		//utils.PrintStructure(res, r.Log, "Resource")
		state, errMsg := r.checkResourceStatus(res)
		r.Log.V(0).Info("Status of " + releaseName + " is " + string(state))
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

func findComponentTemplateByName(templates []app.ComponentTemplate, name string) (*app.ComponentTemplate, error) {
	// TODO(roee.shlomo): BlueprintSpec#Templates should probably be a map from name to the module spec. Then we can remove this function.
	for _, template := range templates {
		if template.Name == name {
			return &template, nil
		}
	}
	return nil, fmt.Errorf("template %s not found", name)
}

func (r *BlueprintReconciler) getExpectedResults(kind string) ([]app.ExpectedResourceStatus, error) {
	ctx := context.Background()

	expected := make([]app.ExpectedResourceStatus, 0)
	var moduleList app.M4DModuleList
	if err := r.List(ctx, &moduleList); err != nil {
		return expected, err
	}
	for _, module := range moduleList.Items {
		for _, res := range module.Spec.Expected {
			if res.Kind == kind {
				expected = append(expected, res)
			}
		}
	}
	return expected, nil
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
