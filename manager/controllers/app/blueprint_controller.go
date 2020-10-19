// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"emperror.dev/errors"
	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"helm.sh/helm/v3/pkg/release"

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
	labels "k8s.io/apimachinery/pkg/labels"
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
		if !equality.Semantic.DeepEqual(&blueprint.Status, observedStatus) {
			if err := r.Client.Status().Update(ctx, &blueprint); err != nil {
				return ctrl.Result{}, errors.WrapWithDetails(err, "failed to update blueprint status", "status", blueprint.Status)
			}
		}
	}

	log.Info("blueprint reconcile cycle completed", "result", result)
	return result, nil
}

func (r *BlueprintReconciler) deleteChartResource(ctx context.Context, kubeNamespace string, releaseName string) (ctrl.Result, error) {
	rel, err := r.Helmer.Status(kubeNamespace, releaseName)
	if err == nil && rel != nil {
		_, _ = r.Helmer.Uninstall(kubeNamespace, releaseName)
	}

	return ctrl.Result{}, nil
}

func (r *BlueprintReconciler) applyChartResource(ctx context.Context, log logr.Logger, ref string, vals map[string]interface{}, kubeNamespace string, releaseName string) (ctrl.Result, error) {
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
	// reset blueprint state
	blueprint.Status.Ready = false
	blueprint.Status.Error = ""
	blueprint.Status.DataAccessInstructions = ""

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

		// TODO: current Copy modules do not expect the "copy" key.
		//       once updated, these lines could be dropped
		if step.Arguments.Flow == app.Copy {
			args = args["copy"].(map[string]interface{})
		}

		// Add metadata arguments (namespace, name, labels, etc.)
		hashedName := utils.Hash(step.Name, 20)
		args["metadata"] = map[string]interface{}{
			"name":      hashedName,
			"namespace": blueprint.Namespace,
			"labels":    blueprint.Labels,
		}
		// we add the "r" character at the beginning of the release name, since it must begin with an alphabetic character
		releaseName := "r" + hashedName
		log.V(0).Info("Release name: " + releaseName)
		// check if the blueprint is about to be deleted
		if blueprint.DeletionTimestamp != nil {
			_, _ = r.deleteChartResource(ctx, blueprint.Namespace, releaseName)
			continue
		}

		numReleases++
		//check the release status
		rel, err := r.Helmer.Status(blueprint.Namespace, releaseName)
		// unexisting release or a failed release - re-apply the chart
		if updateRequired || err != nil || rel == nil || rel.Info.Status == release.StatusFailed {
			// Process templates with arguments
			for _, resource := range templateSpec.Resources {
				if _, err := r.applyChartResource(ctx, log, resource, args, blueprint.Namespace, releaseName); err != nil {
					blueprint.Status.Error += errors.Wrap(err, "ChartDeploymentFailure: ").Error() + "\n"
				}
			}
		} else if rel.Info.Status == release.StatusDeployed {
			// TODO: add release notes of the read module to the status
			if args["flow"] == "read" {
				blueprint.Status.DataAccessInstructions += rel.Info.Notes
			}
			status, errMsg := r.checkReleaseStatus(releaseName, blueprint.Namespace)
			if status == corev1.ConditionFalse {
				blueprint.Status.Error += "ResourceAllocationFailure: " + errMsg + "\n"
			} else if status == corev1.ConditionTrue {
				numReady++
			}
		}
	}
	// check if all releases reached the ready state
	if numReady == numReleases {
		// all modules have been orhestrated successfully - the data is ready for use
		blueprint.Status.Ready = true
		return ctrl.Result{}, nil
	}

	// the status is unknown yet - continue polling
	if blueprint.Status.Error == "" {
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
		// TODO: use kstatus to compute the status of the resources for them the expected results have not been specified
		return corev1.ConditionTrue, ""
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
