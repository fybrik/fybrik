// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"os"
	"testing"

	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/multicluster/dummy"
)

// TestPlotterController runs PlotterReconciler.Reconcile() against a
// fake client that tracks a Plotter object.
// This test does not require a Kubernetes environment to run.
// This mechanism of testing can be used to test corner cases of the reconcile function.
func TestPlotterController(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	var (
		name      = "plotter"
		namespace = utils.GetSystemNamespace()
	)

	var err error
	plotterYAML, err := os.ReadFile("../../testdata/plotter.yaml")
	g.Expect(err).To(gomega.BeNil(), "Cannot read plotter file for test")
	plotter := &app.Plotter{}
	err = yaml.Unmarshal(plotterYAML, plotter)
	g.Expect(err).To(gomega.BeNil(), "Cannot read plotter file for test")

	plotter.Namespace = namespace

	// Objects to track in the fake client.
	objs := []runtime.Object{
		plotter,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	dummyManager := dummy.NewDummyClusterManager(
		make(map[string]*app.Blueprint),
		[]multicluster.Cluster{{
			Name: "thegreendragon",
			Metadata: multicluster.ClusterMetadata{
				Region:        "theshire",
				Zone:          "hobbiton",
				VaultAuthPath: "kubernetes",
			}}})

	// Create a PlotterReconciler object with the scheme and fake client.
	r := &PlotterReconciler{
		Client:         cl,
		Log:            logging.LogInit(logging.CONTROLLER, "test-controller"),
		Scheme:         s,
		ClusterManager: &dummyManager,
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
	g.Expect(blueprintMeta.Namespace).To(gomega.Equal(plotter.Namespace))

	// Simulate that blueprint changes state to Ready=true
	blueprint := dummyManager.DeployedBlueprints["thegreendragon"]
	blueprint.Status.ObservedState.Ready = true
	for instanceName := range blueprint.Spec.Modules {
		if blueprint.Status.ModulesState == nil {
			blueprint.Status.ModulesState = map[string]app.ObservedState{}
		}
		blueprint.Status.ModulesState[instanceName] = app.ObservedState{
			Ready: true,
		}
	}

	deployedBp := dummyManager.DeployedBlueprints["thegreendragon"]
	g.Expect(deployedBp.Labels[app.ApplicationNamespaceLabel]).To(gomega.Equal("default"))
	g.Expect(deployedBp.Labels[app.ApplicationNameLabel]).To(gomega.Equal("notebook"))
	res, err = r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(deployedBp.Spec.Modules["implicit-copy-batch-latest-6575548090"].Chart.Name).
		To(gomega.Equal("ghcr.io/mesh-for-data/m4d-implicit-copy-batch:0.1.0"))
	// Check that the auth path of the credentials is set
	g.Expect(deployedBp.Spec.Modules["implicit-copy-batch-latest-6575548090"].Arguments.Assets[0].Arguments[0].
		Vault[string(taxonomy.ReadFlow)].AuthPath).To(gomega.Equal("/v1/auth/kubernetes/login"))
	g.Expect(deployedBp.Spec.Modules["implicit-copy-batch-latest-6575548090"].Arguments.Assets[0].Arguments[1].
		Vault[string(taxonomy.WriteFlow)].AuthPath).To(gomega.Equal("/v1/auth/kubernetes/login"))
	g.Expect(deployedBp.Spec.Modules["arrow-flight-read"].Arguments.Assets[0].Arguments[0].
		Vault[string(taxonomy.ReadFlow)].AuthPath).To(gomega.Equal("/v1/auth/kubernetes/login"))

	// Check the result of reconciliation to make sure it has the desired state.
	g.Expect(res.Requeue).To(gomega.BeFalse(), "reconcile did not requeue request as expected")

	// Check if Job has been created and has the correct size.
	err = cl.Get(context.TODO(), req.NamespacedName, plotter)
	g.Expect(err).To(gomega.BeNil(), "Can fetch plotter")

	g.Expect(plotter.Status.ObservedState.Ready).To(gomega.BeTrue(), "Plotter is ready")
	for _, assetState := range plotter.Status.Assets {
		g.Expect(assetState.Ready).To(gomega.BeTrue(), "Asset is ready")
	}
	g.Expect(plotter.Status.Assets).To(gomega.HaveLen(1), "Plotter Asset status list contains one element")
}

func TestPlotterWithWriteFlow(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	var (
		name      = "plotter"
		namespace = utils.GetSystemNamespace()
	)

	var err error
	plotterYAML, err := os.ReadFile("../../testdata/plotter-read-write.yaml")
	g.Expect(err).To(gomega.BeNil(), "Cannot read plotter file for test")
	plotter := &app.Plotter{}
	err = yaml.Unmarshal(plotterYAML, plotter)
	g.Expect(err).To(gomega.BeNil(), "Cannot read plotter file for test")

	plotter.Namespace = namespace

	// Objects to track in the fake client.
	objs := []runtime.Object{
		plotter,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	dummyManager := dummy.NewDummyClusterManager(
		make(map[string]*app.Blueprint),
		[]multicluster.Cluster{{
			Name: "thegreendragon",
			Metadata: multicluster.ClusterMetadata{
				Region:        "theshire",
				Zone:          "hobbiton",
				VaultAuthPath: "kubernetes",
			}}})

	// Create a PlotterReconciler object with the scheme and fake client.
	r := &PlotterReconciler{
		Client:         cl,
		Log:            logging.LogInit(logging.CONTROLLER, "test-controller"),
		Scheme:         s,
		ClusterManager: &dummyManager,
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
	g.Expect(blueprintMeta.Namespace).To(gomega.Equal(plotter.Namespace))

	// Simulate that blueprint changes state to Ready=true
	blueprint := dummyManager.DeployedBlueprints["thegreendragon"]
	blueprint.Status.ObservedState.Ready = true
	for instanceName := range blueprint.Spec.Modules {
		if blueprint.Status.ModulesState == nil {
			blueprint.Status.ModulesState = map[string]app.ObservedState{}
		}
		blueprint.Status.ModulesState[instanceName] = app.ObservedState{
			Ready: true,
		}
	}

	deployedBp := dummyManager.DeployedBlueprints["thegreendragon"]
	g.Expect(deployedBp.Spec.ModulesNamespace).To(gomega.Equal("test-modules-namespace"))
	g.Expect(deployedBp.Labels[utils.FybrikModuleNamespace]).To(gomega.Equal("test-modules-namespace"))
	g.Expect(deployedBp.Labels[app.ApplicationNamespaceLabel]).To(gomega.Equal("default"))
	g.Expect(deployedBp.Labels[app.ApplicationNameLabel]).To(gomega.Equal("notebook"))
	res, err = r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(deployedBp.Spec.Modules) == 1).To(gomega.BeTrue(), "Blueprint should have one module")
	g.Expect(deployedBp.Spec.Modules["arrow-flight-module"].Chart.Name).
		To(gomega.Equal("ghcr.io/fybrik/arrow-flight-module-chart:latest"))
	g.Expect(len(deployedBp.Spec.Modules["arrow-flight-module"].AssetIDs) == 3).To(gomega.BeTrue(), "Blueprint AssetIDs should be 3")
	g.Expect(len(deployedBp.Spec.Modules["arrow-flight-module"].Arguments.Assets) == 3).To(gomega.BeTrue(), "Blueprint should have 3 Assets")
	// Check that the auth path of the credentials is set
	g.Expect(deployedBp.Spec.Modules["arrow-flight-module"].Arguments.Assets[0].Arguments[0].
		Vault[string(taxonomy.ReadFlow)].AuthPath).To(gomega.Equal("/v1/auth/kubernetes/login"))
	g.Expect(deployedBp.Spec.Modules["arrow-flight-module"].Arguments.Assets[1].Arguments[0].
		Vault[string(taxonomy.WriteFlow)].AuthPath).To(gomega.Equal("/v1/auth/kubernetes/login"))
	g.Expect(deployedBp.Spec.Modules["arrow-flight-module"].Arguments.Assets[2].Arguments[0].
		Vault[string(taxonomy.ReadFlow)].AuthPath).To(gomega.Equal("/v1/auth/kubernetes/login"))

	// Check the result of reconciliation to make sure it has the desired state.
	g.Expect(res.Requeue).To(gomega.BeFalse(), "reconcile did not requeue request as expected")

	// Check if Plotter has been created
	err = cl.Get(context.TODO(), req.NamespacedName, plotter)
	g.Expect(err).To(gomega.BeNil(), "Can fetch plotter")

	g.Expect(plotter.Status.ObservedState.Ready).To(gomega.BeTrue(), "Plotter is ready")
	for _, assetState := range plotter.Status.Assets {
		g.Expect(assetState.Ready).To(gomega.BeTrue(), "Asset is ready")
	}
	g.Expect(plotter.Status.Assets).To(gomega.HaveLen(2), "Plotter Asset status list contains two elements")
}
