// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"os"
	"strconv"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"

	"emperror.dev/errors"

	"github.com/go-logr/logr"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	"github.com/ibm/the-mesh-for-data/pkg/helm"
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

func (r *BlueprintReconciler) deleteChartResource(ctx context.Context, log logr.Logger, ref string, vals map[string]interface{}) (ctrl.Result, error) {
	kubeNamespace := vals["metadata"].(map[string]interface{})["namespace"].(string)
	releaseName := vals["metadata"].(map[string]interface{})["name"].(string)

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
		rel, err = r.Helmer.Upgrade(chart, kubeNamespace, releaseName, vals, false)
		if err != nil {
			return ctrl.Result{}, errors.WithMessage(err, ref+": failed upgrade")
		}
	} else {
		rel, err = r.Helmer.Install(chart, kubeNamespace, releaseName, vals, false)
		if err != nil {
			return ctrl.Result{}, errors.WithMessage(err, ref+": failed install")
		}
	}
	log.Info(fmt.Sprintf("--- Release Status ---\n\n%s\n\n", rel.Info.Status))
	return ctrl.Result{}, nil
}

func (r *BlueprintReconciler) reconcile(ctx context.Context, log logr.Logger, blueprint *app.Blueprint) (ctrl.Result, error) {
	// Gather all templates and process them into a list of resources to apply
	// The log message below guarantees that 'make e2e' test instruction terminates successfully in case of blueprint errors
	// TODO - fix the e2e test and remove this message
	log.V(0).Info("PILOT_MAGIC_NUMBER=" + os.Getenv("PILOT_MAGIC_NUMBER"))
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
			"namespace": blueprint.Namespace, // TODO(roee.shlomo): should eventually be a dedicated namespace per M4DApplication
			"labels":    blueprint.Labels,
		}

		// Process templates with arguments
		for _, resource := range templateSpec.Resources {
			if blueprint.DeletionTimestamp != nil {
				_, _ = r.deleteChartResource(ctx, log, resource, args)
				continue
			}

			if res, err := r.applyChartResource(ctx, log, resource, args); err != nil {
				r.Log.V(0).Info("failed to apply chart: " + err.Error())
				return res, err
			}
		}
	}

	log.V(0).Info("PILOT_MAGIC_NUMBER=" + os.Getenv("PILOT_MAGIC_NUMBER"))
	blueprint.Status.Ready = false // Won't be ready until all components are running.  TODO - add status checks
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
