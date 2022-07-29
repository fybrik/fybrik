// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	app "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/logging"
)

// create FybrikModule controller with mockup interfaces
func createTestFybrikModuleController(cl client.Client, s *runtime.Scheme) *FybrikModuleReconciler {
	// Create a FybrikModuleReconciler object with the scheme and fake client.
	return &FybrikModuleReconciler{
		Client: cl,
		Name:   "TestModuleReconciler",
		Log:    logging.LogInit(logging.CONTROLLER, "test-module-controller"),
		Scheme: s,
	}
}

func TestFybrikModuleWithInvalidInterface(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)

	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "module-with-interface-errors",
		Namespace: "fybrik-system",
	}
	filename := "../../testdata/unittests/fybrikmodule-interfaceErrors.yaml"
	fybrikModule := &app.FybrikModule{}
	g.Expect(readObjectFromFile(filename, fybrikModule)).NotTo(gomega.HaveOccurred())
	fybrikModule.SetGeneration(1)

	// Objects to track in the fake client.
	objs := []runtime.Object{
		fybrikModule,
	}
	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Create a FybrikModuleReconciler object with the scheme and fake client.
	r := createTestFybrikModuleController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}
	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	newModule := &app.FybrikModule{}
	err = cl.Get(context.Background(), req.NamespacedName, newModule)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrik module")
	g.Expect(newModule.Status.Conditions[ModuleValidationConditionIndex].Status).To(gomega.BeIdenticalTo(corev1.ConditionFalse))
}

func TestFybrikModuleWithInvalidActions(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)

	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "module-with-actions-errors",
		Namespace: "fybrik-system",
	}
	filename := "../../testdata/unittests/fybrikmodule-actionsErrors.yaml"
	fybrikModule := &app.FybrikModule{}
	g.Expect(readObjectFromFile(filename, fybrikModule)).NotTo(gomega.HaveOccurred())
	fybrikModule.SetGeneration(1)

	// Objects to track in the fake client.
	objs := []runtime.Object{
		fybrikModule,
	}
	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Create a FybrikModuleReconciler object with the scheme and fake client.
	r := createTestFybrikModuleController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}
	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	newModule := &app.FybrikModule{}
	err = cl.Get(context.Background(), req.NamespacedName, newModule)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrik module")
	g.Expect(newModule.Status.Conditions[ModuleValidationConditionIndex].Status).To(gomega.BeIdenticalTo(corev1.ConditionFalse))
}

func TestFybrikModuleWithValidFields(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)

	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "valid-module",
		Namespace: "fybrik-system",
	}
	filename := "../../testdata/unittests/fybrikmodule-validActions.yaml"
	fybrikModule := &app.FybrikModule{}
	g.Expect(readObjectFromFile(filename, fybrikModule)).NotTo(gomega.HaveOccurred())
	fybrikModule.SetGeneration(1)

	// Objects to track in the fake client.
	objs := []runtime.Object{
		fybrikModule,
	}
	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Create a FybrikModuleReconciler object with the scheme and fake client.
	r := createTestFybrikModuleController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}
	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	newModule := &app.FybrikModule{}
	err = cl.Get(context.Background(), req.NamespacedName, newModule)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrik module")
	g.Expect(newModule.Status.Conditions[ModuleValidationConditionIndex].Status).To(gomega.BeIdenticalTo(corev1.ConditionTrue))
}
