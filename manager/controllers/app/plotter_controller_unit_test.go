// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"io/ioutil"

	"fmt"
	"testing"

	"fybrik.io/fybrik/manager/controllers/utils"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/multicluster/dummy"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

// TestBatchTransferController runs BatchTransferReconciler.Reconcile() against a
// fake client that tracks a BatchTransfer object.
// This test does not require a Kubernetes environment to run.
// This mechanism of testing can be used to test corner cases of the reconcile function.
func TestPlotterController(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	controllerNamespace := getControllerNamespace()
	blueprintNamespace := getBlueprintNamespace()
	fmt.Printf("Using controller namespace: " + controllerNamespace + " using blueprint namespace: " + blueprintNamespace)

	var (
		name      = "plotter"
		namespace = controllerNamespace
	)

	var err error
	plotterYAML, err := ioutil.ReadFile("../../testdata/plotter.yaml")
	g.Expect(err).To(gomega.BeNil(), "Cannot read plotter file for test")
	plotter := &app.Plotter{}
	err = yaml.Unmarshal(plotterYAML, plotter)
	g.Expect(err).To(gomega.BeNil(), "Cannot read plotter file for test")

	plotter.Namespace = controllerNamespace

	// Objects to track in the fake client.
	objs := []runtime.Object{
		plotter,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	dummyManager := &dummy.ClusterManager{
		DeployedBlueprints: make(map[string]*app.Blueprint),
	}

	// Create a BatchTransferReconciler object with the scheme and fake client.
	r := &PlotterReconciler{
		Client:         cl,
		Log:            ctrl.Log.WithName("test-controller"),
		Scheme:         s,
		ClusterManager: dummyManager,
	}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	// Check the result of reconciliation to make sure it has the desired state.
	g.Expect(res.Requeue).To(gomega.BeFalse(), "reconcile did not requeue request as expected")

	// Check if Job has been created and has the correct size.
	err = cl.Get(context.TODO(), req.NamespacedName, plotter)
	g.Expect(err).To(gomega.BeNil(), "Can fetch plotter")

	g.Expect(plotter.Status.Blueprints).To(gomega.HaveKey("thegreendragon"))
	blueprintMeta := plotter.Status.Blueprints["thegreendragon"]
	g.Expect(blueprintMeta.Name).To(gomega.Equal(plotter.Name))
	g.Expect(blueprintMeta.Namespace).To(gomega.Equal(blueprintNamespace))

	// Simulate that blueprint changes state to Ready=true
	dummyManager.DeployedBlueprints["thegreendragon"].Status.ObservedState.Ready = true
	dummyManager.DeployedBlueprints["thegreendragon"].Status.ObservedState.DataAccessInstructions = "nop"

	deployedBp := dummyManager.DeployedBlueprints["thegreendragon"]
	g.Expect(deployedBp.Labels[app.ApplicationNamespaceLabel]).To(gomega.Equal("default"))
	g.Expect(deployedBp.Labels[app.ApplicationNameLabel]).To(gomega.Equal("notebook"))
	res, err = r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	// Check the result of reconciliation to make sure it has the desired state.
	g.Expect(res.Requeue).To(gomega.BeFalse(), "reconcile did not requeue request as expected")

	// Check if Job has been created and has the correct size.
	err = cl.Get(context.TODO(), req.NamespacedName, plotter)
	g.Expect(err).To(gomega.BeNil(), "Can fetch plotter")

	g.Expect(plotter.Status.ObservedState.Ready).To(gomega.BeTrue(), "Plotter is ready")
	g.Expect(plotter.Status.ObservedState.DataAccessInstructions).To(gomega.Equal("nop\n"), "Plotter is ready")
}
