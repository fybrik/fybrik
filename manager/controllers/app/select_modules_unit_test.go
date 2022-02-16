// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestCheckDependencies(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	application := &app.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/fybrikcopyapp-csv.yaml", application)).To(gomega.BeNil(), "Cannot read fybrikapplication file for test")
	application.SetGeneration(1)
	application.SetUID("1")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	readModule := &app.FybrikModule{}
	copyModule := &app.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	copyModule.Namespace = utils.GetControllerNamespace()
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()

	// Create modules in fake K8s agent
	g.Expect(cl.Create(context.Background(), copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), readModule)).NotTo(gomega.HaveOccurred())

	found, missing := CheckDependencies(readModule, map[string]*app.FybrikModule{"copyModule": copyModule})

	g.Expect(len(found)).To(gomega.Equal(0))
	g.Expect(len(missing)).To(gomega.Equal(0))
}

func TestSupportsDependencies(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	application := &app.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/fybrikcopyapp-csv.yaml", application)).To(gomega.BeNil(), "Cannot read fybrikapplication file for test")
	application.SetGeneration(1)
	application.SetUID("2")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	readModule := &app.FybrikModule{}
	copyModule := &app.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	copyModule.Namespace = utils.GetControllerNamespace()
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()

	// Create modules in fake K8s agent
	g.Expect(cl.Create(context.Background(), copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), readModule)).NotTo(gomega.HaveOccurred())

	support := SupportsDependencies(readModule, map[string]*app.FybrikModule{})
	g.Expect(support).To(gomega.Equal(true))
}

func TestGetDependencies(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	application := &app.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/fybrikcopyapp-csv.yaml", application)).To(gomega.BeNil(), "Cannot read fybrikapplication file for test")
	application.SetGeneration(1)
	application.SetUID("3")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	readModule := &app.FybrikModule{}
	copyModule := &app.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	copyModule.Namespace = utils.GetControllerNamespace()
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()

	// Create modules in fake K8s agent
	g.Expect(cl.Create(context.Background(), copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), readModule)).NotTo(gomega.HaveOccurred())

	dependencies, err := GetDependencies(readModule, map[string]*app.FybrikModule{})
	g.Expect(len(dependencies)).To(gomega.Equal(0))
	g.Expect(err).To(gomega.BeNil())
}
