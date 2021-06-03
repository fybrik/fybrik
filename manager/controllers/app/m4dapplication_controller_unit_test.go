// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mesh-for-data/mesh-for-data/manager/controllers/mockup"
	"github.com/mesh-for-data/mesh-for-data/pkg/storage"

	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

// Read utility
func readObjectFromFile(f string, obj interface{}) error {
	bytes, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, obj)
}

// create M4DApplication controller with mockup interfaces
func createTestM4DApplicationController(cl client.Client, s *runtime.Scheme) *M4DApplicationReconciler {
	// Create a M4DApplicationReconciler object with the scheme and fake client.
	return &M4DApplicationReconciler{
		Client:         cl,
		Name:           "TestReconciler",
		Log:            ctrl.Log.WithName("test-controller"),
		Scheme:         s,
		PolicyCompiler: &mockup.MockPolicyCompiler{},
		ResourceInterface: &PlotterInterface{
			Client: cl,
		},
		ClusterManager: &mockup.ClusterLister{},
		Provision:      &storage.ProvisionTest{},
		DataCatalog:    mockup.NewTestCatalog(),
	}
}

// TestM4DApplicationController runs M4DApplicationReconciler.Reconcile() against a
// fake client that tracks a M4dApplication object.
// This test does not require a Kubernetes environment to run.
// This mechanism of testing can be used to test corner cases of the reconcile function.
func TestM4DApplicationControllerCSVCopyAndRead(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	var (
		name      = "notebook"
		namespace = "default"
	)
	application := &app.M4DApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/m4dcopyapp-csv.yaml", application)).To(gomega.BeNil(), "Cannot read m4dapplication file for test")

	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	readModule := &app.M4DModule{}
	copyModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())

	// Create modules in fake K8s agent
	g.Expect(cl.Create(context.Background(), copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), readModule)).NotTo(gomega.HaveOccurred())

	// Create storage account
	dummySecret := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-theshire.yaml", dummySecret)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), dummySecret)).NotTo(gomega.HaveOccurred())
	account := &app.M4DStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), account)).NotTo(gomega.HaveOccurred())

	r := createTestM4DApplicationController(cl, s)
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

	// Check if Application generated a plotter
	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	g.Expect(application.Status.Generated).NotTo(gomega.BeNil())

	plotterObjectKey := types.NamespacedName{
		Namespace: "m4d-system",
		Name:      "notebook-default",
	}
	plotter := &app.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	bpSpec := plotter.Spec.Blueprints["thegreendragon"]
	g.Expect(bpSpec.Flow.Steps[0].Template).To(gomega.Equal("implicit-copy-batch"))
	g.Expect(bpSpec.Flow.Steps[0].Arguments.Copy.Source.Format).To(gomega.Equal("csv"))
	g.Expect(bpSpec.Flow.Steps[0].Arguments.Copy.Destination.Format).To(gomega.Equal("csv"))
	g.Expect(bpSpec.Flow.Steps[0].Arguments.Copy.Destination.Format).To(gomega.Equal(bpSpec.Flow.Steps[1].Arguments.Read[0].Source.Format))
}

// This test checks proper reconciliation of M4DApplication finalizers
func TestM4DApplicationFinalizers(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	application := &app.M4DApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())

	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Create a M4DApplicationReconciler object with the scheme and fake client.
	r := createTestM4DApplicationController(cl, s)

	g.Expect(r.reconcileFinalizers(application)).To(gomega.BeNil())
	g.Expect(application.Finalizers).NotTo(gomega.BeEmpty(), "finalizers have not been created")
	// mark application as deleted
	application.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	g.Expect(r.reconcileFinalizers(application)).To(gomega.BeNil())
	g.Expect(application.Finalizers).To(gomega.BeEmpty(), "finalizers have not been removed")
}

// Tests denial of the access to data
// Assumptions on response from connectors:
// Enforcement action for read operation: Deny
// Result: an error
func TestDenyOnRead(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &app.M4DApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0] = app.DataContext{
		DataSetID:    "{\"asset_id\": \"deny-dataset\", \"catalog_id\": \"s3\"}",
		Requirements: app.DataRequirements{Interface: app.InterfaceDetails{Protocol: app.S3, DataFormat: app.Parquet}},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Create a M4DApplicationReconciler object with the scheme and fake client.
	r := createTestM4DApplicationController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	// Expect an error
	g.Expect(getErrorMessages(application)).To(gomega.ContainSubstring(app.ReadAccessDenied))
}

// Tests selection of read-path module
// Read module does not have api for s3/parquet
// Result: an error
func TestNoReadPath(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &app.M4DApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0] = app.DataContext{
		DataSetID:    "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"db2\"}",
		Requirements: app.DataRequirements{Interface: app.InterfaceDetails{Protocol: app.JdbcDb2, DataFormat: app.Table}},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	// Create a M4DApplicationReconciler object with the scheme and fake client.
	r := createTestM4DApplicationController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	// Expect an error
	g.Expect(getErrorMessages(application)).To(gomega.ContainSubstring(app.ModuleNotFound))
	g.Expect(getErrorMessages(application)).To(gomega.ContainSubstring("read"))
}

// Tests finding a module for copy
// Assumptions on response from connectors:
// Two datasets:
// Kafka dataset, a copy is required.
// S3 dataset, no copy is needed
// Enforcement action for both datasets: Allow
// No copy module (kafka->s3)
// Result: an error
func TestWrongCopyModule(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &app.M4DApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data = []app.DataContext{
		{
			DataSetID:    "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"s3\"}",
			Requirements: app.DataRequirements{Interface: app.InterfaceDetails{Protocol: app.ArrowFlight, DataFormat: app.Arrow}},
		},
		{
			DataSetID:    "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"kafka\"}",
			Requirements: app.DataRequirements{Interface: app.InterfaceDetails{Protocol: app.ArrowFlight, DataFormat: app.Arrow}},
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	copyModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/copy-db2-parquet.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")
	// Create a M4DApplicationReconciler object with the scheme and fake client.
	r := createTestM4DApplicationController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	// Expect an error
	g.Expect(getErrorMessages(application)).To(gomega.ContainSubstring(app.ModuleNotFound))
	g.Expect(getErrorMessages(application)).To(gomega.ContainSubstring("copy"))
}

// Tests finding a module for copy supporting actions
// Assumptions on response from connectors:
// db2 dataset
// Enforcement action: Redact
// copy (db2->s3) and read modules do not support redact action
// Result: an error
func TestActionSupport(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &app.M4DApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0] = app.DataContext{
		DataSetID:    "{\"asset_id\": \"redact-dataset\", \"catalog_id\": \"db2\"}",
		Requirements: app.DataRequirements{Interface: app.InterfaceDetails{Protocol: app.ArrowFlight, DataFormat: app.Arrow}},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	copyModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/copy-db2-parquet-no-transforms.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")
	// Create a M4DApplicationReconciler object with the scheme and fake client.
	r := createTestM4DApplicationController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	// Expect an error
	g.Expect(getErrorMessages(application)).To(gomega.ContainSubstring(app.ModuleNotFound))
}

// Assumptions on response from connectors:
// Two datasets:
// Db2 dataset, a copy is required.
// S3 dataset, no copy is needed
// Enforcement actions for the first dataset: redact
// Enforcement action for the second dataset: Allow
// Applied copy module db2->s3 supporting redact action
// Result: plotter with a single blueprint is created successfully, a read module is applied once for both datasets

func TestMultipleDatasets(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &app.M4DApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data = []app.DataContext{
		{
			DataSetID:    "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"s3\"}",
			Requirements: app.DataRequirements{Interface: app.InterfaceDetails{Protocol: app.ArrowFlight, DataFormat: app.Arrow}},
		},
		{
			DataSetID:    "{\"asset_id\": \"redact-dataset\", \"catalog_id\": \"db2\"}",
			Requirements: app.DataRequirements{Interface: app.InterfaceDetails{Protocol: app.ArrowFlight, DataFormat: app.Arrow}},
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	copyModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/copy-db2-parquet.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")
	// Create storage account
	dummySecret := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-theshire.yaml", dummySecret)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), dummySecret)).NotTo(gomega.HaveOccurred())
	account := &app.M4DStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), account)).NotTo(gomega.HaveOccurred())

	// Create a M4DApplicationReconciler object with the scheme and fake client.
	r := createTestM4DApplicationController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	// check provisioned storage
	g.Expect(application.Status.ProvisionedStorage["{\"asset_id\": \"redact-dataset\", \"catalog_id\": \"db2\"}"].DatasetRef).ToNot(gomega.BeEmpty(), "No storage provisioned")
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &app.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(len(plotter.Spec.Blueprints)).To(gomega.Equal(1), "A single blueprint should be created")
	// Check the generated blueprint
	// There should be a single read module with two datasets
	blueprint := plotter.Spec.Blueprints["thegreendragon"]
	g.Expect(blueprint).NotTo(gomega.BeNil())
	numReads := 0
	for _, step := range blueprint.Flow.Steps {
		if len(step.Arguments.Read) > 0 {
			numReads++
			g.Expect(len(step.Arguments.Read)).To(gomega.Equal(2), "A read module should support both datasets")
		}
	}
	g.Expect(numReads).To(gomega.Equal(1), "A single read module should be instantiated")
}

// This test checks the case where data comes from another regions, and should be redacted.
// In this case a read module will be deployed close to the compute, while a copy module - close to the data.
func TestMultipleRegions(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &app.M4DApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0] = app.DataContext{
		DataSetID:    "{\"asset_id\": \"redact-dataset\", \"catalog_id\": \"s3-external\"}",
		Requirements: app.DataRequirements{Interface: app.InterfaceDetails{Protocol: app.ArrowFlight, DataFormat: app.Arrow}},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	copyModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/copy-csv-parquet.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")
	// Create storage account
	dummySecret := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-theshire.yaml", dummySecret)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), dummySecret)).NotTo(gomega.HaveOccurred())
	account := &app.M4DStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), account)).NotTo(gomega.HaveOccurred())

	// Create a M4DApplicationReconciler object with the scheme and fake client.
	r := createTestM4DApplicationController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	// check provisioned storage
	g.Expect(application.Status.ProvisionedStorage["{\"asset_id\": \"redact-dataset\", \"catalog_id\": \"s3-external\"}"].DatasetRef).ToNot(gomega.BeEmpty(), "No storage provisioned")
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &app.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(len(plotter.Spec.Blueprints)).To(gomega.Equal(2))
}

// This test checks the ingest scenario - copy is required, no workload specified.
// Two storage accounts are created. Data cannot be stored in one of them according to governance policies.
func TestCopyData(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	assetName := "{\"asset_id\": \"allow-theshire\", \"catalog_id\": \"s3-external\"}"
	namespaced := types.NamespacedName{
		Name:      "ingest",
		Namespace: "default",
	}
	application := &app.M4DApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/ingest.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0].DataSetID = assetName
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	copyModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")
	// Create storage accounts
	secret1 := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-neverland.yaml", secret1)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), secret1)).NotTo(gomega.HaveOccurred())
	account1 := &app.M4DStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-neverland.yaml", account1)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), account1)).NotTo(gomega.HaveOccurred())
	secret2 := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-theshire.yaml", secret2)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), secret2)).NotTo(gomega.HaveOccurred())
	account2 := &app.M4DStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account2)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), account2)).NotTo(gomega.HaveOccurred())

	// Create a M4DApplicationReconciler object with the scheme and fake client.
	r := createTestM4DApplicationController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	// check provisioned storage
	g.Expect(application.Status.ProvisionedStorage[assetName].DatasetRef).ToNot(gomega.BeEmpty(), "No storage provisioned")
	g.Expect(application.Status.ProvisionedStorage[assetName].SecretRef).To(gomega.Equal("credentials-theshire"), "Incorrect storage was selected")
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &app.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	// There should be a single copy module
	g.Expect(len(plotter.Spec.Blueprints)).To(gomega.Equal(1))
	blueprint := plotter.Spec.Blueprints["thegreendragon"]
	g.Expect(blueprint).NotTo(gomega.BeNil())
	g.Expect(len(blueprint.Flow.Steps)).To(gomega.Equal(1))
}

// This test checks the ingest scenario
// A storage account has been defined for the region where the dataset can not be written to according to governance policies.
// An error is received.
func TestCopyDataNotAllowed(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	assetName := "{\"asset_id\": \"deny-theshire\", \"catalog_id\": \"s3-external\"}"
	namespaced := types.NamespacedName{
		Name:      "ingest",
		Namespace: "default",
	}
	application := &app.M4DApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/ingest.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0].DataSetID = assetName
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	copyModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")

	// Create storage account
	dummySecret := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-theshire.yaml", dummySecret)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), dummySecret)).NotTo(gomega.HaveOccurred())
	account := &app.M4DStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.TODO(), account)).NotTo(gomega.HaveOccurred())

	// Create a M4DApplicationReconciler object with the scheme and fake client.
	r := createTestM4DApplicationController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	// check provisioned storage
	g.Expect(application.Status.ProvisionedStorage).To(gomega.BeEmpty())
	// check errors
	g.Expect(getErrorMessages(application)).NotTo(gomega.BeEmpty())
}

// This test checks that the plotter state propagates into the m4dapp state
func TestPlotterUpdate(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &app.M4DApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0] = app.DataContext{
		DataSetID:    "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"s3\"}",
		Requirements: app.DataRequirements{Interface: app.InterfaceDetails{Protocol: app.ArrowFlight, DataFormat: app.Arrow}},
	}
	application.SetGeneration(1)

	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &app.M4DModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")

	// Create a M4DApplicationReconciler object with the scheme and fake client.
	r := createTestM4DApplicationController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.Background(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	g.Expect(application.Status.Generated.AppVersion).To(gomega.Equal(application.Generation))
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &app.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	// mark the plotter as in error state
	errorMsg := "failure to orchestrate modules"
	plotter.Status.ObservedState.Error = errorMsg
	g.Expect(cl.Update(context.Background(), plotter)).NotTo(gomega.HaveOccurred())

	// the new reconcile should update the application state
	_, err = r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())
	err = cl.Get(context.Background(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	g.Expect(getErrorMessages(application)).To(gomega.ContainSubstring(errorMsg))

	// mark the plotter as ready
	plotter.Status.ObservedState.Error = ""
	plotter.Status.ObservedState.Ready = true
	g.Expect(cl.Update(context.Background(), plotter)).NotTo(gomega.HaveOccurred())

	// the new reconcile should update the application state
	_, err = r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())
	err = cl.Get(context.Background(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	g.Expect(application.Status.Ready).To(gomega.BeTrue())
}

// This test checks that the older plotter state does not propagate into the m4dapp state
func TestSyncWithPlotter(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "notebook",
		Namespace: "default",
	}
	application := &app.M4DApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/m4dcopyapp-csv.yaml", application)).NotTo(gomega.HaveOccurred())
	// imitate a ready phase for the earlier generation
	application.SetGeneration(2)
	application.Finalizers = []string{"TestReconciler.finalizer"}
	application.Status.Generated = &app.ResourceReference{Name: "plotter", Namespace: "m4d-system", Kind: "Plotter", AppVersion: 1}
	application.Status.Ready = true
	application.Status.ObservedGeneration = 1

	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	plotter := &app.Plotter{}
	g.Expect(readObjectFromFile("../../testdata/plotter.yaml", plotter)).NotTo(gomega.HaveOccurred())
	plotter.Status.ObservedState.Ready = true
	g.Expect(cl.Create(context.Background(), plotter)).NotTo(gomega.HaveOccurred())

	// Create a M4DApplicationReconciler object with the scheme and fake client.
	r := createTestM4DApplicationController(cl, s)
	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	newApp := &app.M4DApplication{}
	err = cl.Get(context.Background(), req.NamespacedName, newApp)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch m4dapplication")
	g.Expect(getErrorMessages(newApp)).NotTo(gomega.BeEmpty())
	g.Expect(newApp.Status.Ready).NotTo(gomega.BeTrue())
}
