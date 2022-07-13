// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	"fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/mockup"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/infrastructure"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/storage"
)

// Read utility
func readObjectFromFile(f string, obj interface{}) error {
	bytes, err := os.ReadFile(f)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, obj)
}

// create cluster-metadata config map
func createClusterMetadata() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-metadata",
			Namespace: utils.GetSystemNamespace(),
		},
		Data: map[string]string{
			"ClusterName":   "thegreendragon",
			"Zone":          "hobbiton",
			"Region":        "theshire",
			"VaultAuthPath": "kind",
		},
	}
}

// create FybrikApplication controller with mockup interfaces
func createTestFybrikApplicationController(cl client.Client, s *runtime.Scheme) *FybrikApplicationReconciler {
	log := logging.LogInit("test", "ConfigPolicyEvaluator")
	// environment: cluster-metadata configmap
	_ = cl.Create(context.Background(), createClusterMetadata())
	infrastructureManager, err := infrastructure.NewAttributeManager()
	if err != nil {
		log.Error().Err(err).Msg("unable to get infrastructure attributes")
		return nil
	}
	evaluator, err := adminconfig.NewRegoPolicyEvaluator()
	if err != nil {
		log.Error().Err(err).Msg("unable to compile policies")
		return nil
	}
	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	return &FybrikApplicationReconciler{
		Client:        cl,
		Name:          "TestReconciler",
		Log:           logging.LogInit(logging.CONTROLLER, "test-controller"),
		Scheme:        s,
		PolicyManager: &mockup.MockPolicyManager{},
		DataCatalog:   mockup.NewTestCatalog(),
		ResourceInterface: &PlotterInterface{
			Client: cl,
		},
		ClusterManager:  &mockup.ClusterLister{},
		Provision:       &storage.ProvisionTest{},
		ConfigEvaluator: evaluator,
		Infrastructure:  infrastructureManager,
	}
}

// TestFybrikApplicationController runs FybrikApplicationReconciler.Reconcile() against a
// fake client that tracks a FybrikApplication object.
// This test does not require a Kubernetes environment to run.
// This mechanism of testing can be used to test corner cases of the reconcile function.
func TestFybrikApplicationControllerCSVCopyAndRead(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	var (
		name      = "notebook"
		namespace = "default"
	)
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/fybrikcopyapp-csv.yaml", application)).To(gomega.BeNil(),
		"Cannot read fybrikapplication file for test")
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

	readModule := &v1alpha1.FybrikModule{}
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	copyModule.Namespace = utils.GetControllerNamespace()
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()

	// Create modules in fake K8s agent
	g.Expect(cl.Create(context.Background(), copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(cl.Create(context.Background(), readModule)).NotTo(gomega.HaveOccurred())

	// Create storage account
	dummySecret := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-theshire.yaml", dummySecret)).NotTo(gomega.HaveOccurred())
	dummySecret.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), dummySecret)).NotTo(gomega.HaveOccurred())
	account := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	account.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), account)).NotTo(gomega.HaveOccurred())

	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())
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
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(application.Status.Generated).NotTo(gomega.BeNil())

	controllerNamespace := utils.GetControllerNamespace()
	fmt.Printf("FybrikApplication unit test: controller namespace " + controllerNamespace)

	plotterObjectKey := types.NamespacedName{
		Namespace: controllerNamespace,
		Name:      "notebook-default",
	}
	plotter := &v1alpha1.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	// Check that plotter assets have the expected properties
	g.Expect(plotter.Spec.Assets).To(gomega.HaveLen(2)) // 2 assets should have been created. the original and the implicit from the copy
	g.Expect(plotter.Spec.Assets).To(gomega.HaveKey("s3-csv/redact-dataset"))
	g.Expect(plotter.Spec.Assets).To(gomega.HaveKey("s3-csv/redact-dataset-copy"))
	dataStore := plotter.Spec.Assets["s3-csv/redact-dataset"].DataStore
	dataStoreMap := dataStore.Connection.AdditionalProperties.Items
	g.Expect(dataStoreMap).To(gomega.HaveKey("s3"))
	s3Config := dataStoreMap["s3"].(map[string]interface{})
	g.Expect(s3Config["endpoint"]).To(gomega.Equal("s3.eu-gb.cloud-object-storage.appdomain.cloud"))
	g.Expect(s3Config["bucket"]).To(gomega.Equal("fybrik-test-bucket"))
	g.Expect(s3Config["object_key"]).To(gomega.Equal("small.csv"))

	// Check templates
	g.Expect(plotter.Spec.Templates).To(gomega.HaveLen(2))
	g.Expect(plotter.Spec.Flows).To(gomega.HaveLen(1))
	flow := plotter.Spec.Flows[0]
	g.Expect(flow.AssetID).To(gomega.Equal("s3-csv/redact-dataset"))
	g.Expect(flow.FlowType).To(gomega.Equal(taxonomy.ReadFlow))
	g.Expect(flow.SubFlows).To(gomega.HaveLen(2)) // Should have two subflows
	copyFlow := flow.SubFlows[0]                  // Assume flow 0 is copy
	g.Expect(copyFlow.FlowType).To(gomega.Equal(taxonomy.CopyFlow))
	g.Expect(copyFlow.Triggers).To(gomega.ContainElements(v1alpha1.InitTrigger))
	readFlow := flow.SubFlows[1]
	g.Expect(readFlow.FlowType).To(gomega.Equal(taxonomy.ReadFlow))
	g.Expect(readFlow.Triggers).To(gomega.ContainElements(v1alpha1.WorkloadTrigger))
	g.Expect(readFlow.Steps[0][0].Cluster).To(gomega.Equal("thegreendragon"))
	// Check statuses
	g.Expect(application.Status.Ready).To(gomega.Equal(false))
	endpoint := application.Status.AssetStates[application.Spec.Data[0].DataSetID].Endpoint
	g.Expect(endpoint).To(gomega.Not(gomega.BeNil()))
	connectionMap := endpoint.AdditionalProperties.Items
	g.Expect(connectionMap).To(gomega.HaveKey("fybrik-arrow-flight"))
	config := connectionMap["fybrik-arrow-flight"].(map[string]interface{})
	g.Expect(config["hostname"]).To(gomega.Equal("read-path.notebook-default-arrow-flight-module.notebook"))
	g.Expect(config["scheme"]).To(gomega.Equal("grpc"))
}

// This test checks proper reconciliation of FybrikApplication finalizers
func TestFybrikApplicationFinalizers(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)

	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())

	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())
	appContext := ApplicationContext{Application: application, Log: &r.Log}
	g.Expect(r.addFinalizers(context.TODO(), appContext)).To(gomega.BeNil())
	g.Expect(application.Finalizers).NotTo(gomega.BeEmpty(), "finalizers have not been created")
	// mark application as deleted
	application.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	g.Expect(r.removeFinalizers(context.TODO(), appContext)).To(gomega.BeNil())
	g.Expect(application.Finalizers).To(gomega.BeEmpty(), "finalizers have not been removed")
}

// Tests denial of the access to data
// Assumptions on response from connectors:
// Enforcement action for read operation: Deny
// Result: Deny condition, no reconcile
func TestDenyOnRead(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0] = v1alpha1.DataContext{
		DataSetID:    "s3/deny-dataset",
		Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.S3, DataFormat: v1alpha1.Parquet}},
	}
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

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	res, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	// Expect Deny condition
	cond := application.Status.AssetStates["s3/deny-dataset"].Conditions[DenyConditionIndex]
	g.Expect(cond.Status).To(gomega.BeIdenticalTo(corev1.ConditionTrue), "Deny condition is not set")
	g.Expect(cond.Message).To(gomega.ContainSubstring(v1alpha1.ReadAccessDenied))
	g.Expect(application.Status.Ready).To(gomega.BeTrue())
	g.Expect(res).To(gomega.BeEquivalentTo(ctrl.Result{}), "Requests another reconcile")
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
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0] = v1alpha1.DataContext{
		DataSetID:    "db2/allow-dataset",
		Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.JdbcDB2}},
	}
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

	// Read module
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	// Expect an error
	g.Expect(getErrorMessages(application)).NotTo(gomega.BeEmpty())
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
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data = []v1alpha1.DataContext{
		{
			DataSetID:    "s3/allow-dataset",
			Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
		},
		{
			DataSetID:    "kafka/allow-dataset",
			Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
		},
	}
	application.SetGeneration(1)
	application.SetUID("4")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/copy-db2-parquet.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	copyModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")
	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	// Expect an error
	g.Expect(getErrorMessages(application)).NotTo(gomega.BeEmpty())
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
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0] = v1alpha1.DataContext{
		DataSetID:    "db2/redact-dataset",
		Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
	}
	application.SetGeneration(1)
	application.SetUID("5")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/copy-db2-parquet-no-transforms.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	copyModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")
	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	// Expect an error
	g.Expect(getErrorMessages(application)).NotTo(gomega.BeEmpty())
}

// Assumptions on response from connectors:
// Datasets:
// S3 dataset, no access is granted.
// Db2 dataset, a copy is required.
// S3 dataset, no copy is needed
// Enforcement actions for the second dataset: redact
// Enforcement action for the third dataset: Allow
// Applied copy module db2->s3 supporting redact action
// Result: plotter with a single blueprint is created successfully, a read module is applied once

func TestMultipleDatasets(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data = []v1alpha1.DataContext{
		{
			DataSetID:    "s3/deny-dataset",
			Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
		},
		{
			DataSetID:    "s3/allow-dataset",
			Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
		},
		{
			DataSetID:    "db2/redact-dataset",
			Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
		},
	}
	application.SetGeneration(1)
	application.SetUID("6")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/copy-db2-parquet.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	copyModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")
	// Create storage account
	dummySecret := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-theshire.yaml", dummySecret)).NotTo(gomega.HaveOccurred())
	dummySecret.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), dummySecret)).NotTo(gomega.HaveOccurred())
	account := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	account.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), account)).NotTo(gomega.HaveOccurred())

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	// check Deny for the first dataset
	g.Expect(application.Status.AssetStates["s3/deny-dataset"].Conditions[DenyConditionIndex].Status).
		To(gomega.BeIdenticalTo(corev1.ConditionTrue))
	// check provisioned storage
	g.Expect(application.Status.ProvisionedStorage["db2/redact-dataset"].DatasetRef).ToNot(gomega.BeEmpty(), "No storage provisioned")
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &v1alpha1.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(plotter.Spec.Assets).To(gomega.HaveLen(3))    // 3 assets. 2 original and one implicit copy asset
	g.Expect(plotter.Spec.Templates).To(gomega.HaveLen(2)) // expect two templates
	g.Expect(plotter.Spec.Flows).To(gomega.HaveLen(2))     // two flows. one for each valid asset
	g.Expect(plotter.Spec.Flows[0].AssetID).To(gomega.Equal("s3/allow-dataset"))
	g.Expect(plotter.Spec.Flows[1].AssetID).To(gomega.Equal("db2/redact-dataset"))
	g.Expect(plotter.Spec.Flows[0].SubFlows).To(gomega.HaveLen(1))
	g.Expect(plotter.Spec.Flows[1].SubFlows).To(gomega.HaveLen(2))
}

// Tests that the taxonomy is properly compiled
// with the FilterAction transformation
func TestFilterAsset(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data = []v1alpha1.DataContext{
		{
			DataSetID:    "s3/filter-dataset",
			Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
		},
	}
	application.SetGeneration(1)
	application.SetUID("23")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet-filter.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &v1alpha1.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(plotter.Spec.Assets).To(gomega.HaveLen(1))    // single asset
	g.Expect(plotter.Spec.Templates).To(gomega.HaveLen(1)) // expect one template
	g.Expect(plotter.Spec.Flows).To(gomega.HaveLen(1))     // Single flow
	g.Expect(plotter.Spec.Flows[0].AssetID).To(gomega.Equal("s3/filter-dataset"))
	g.Expect(plotter.Spec.Flows[0].SubFlows).To(gomega.HaveLen(1))
	g.Expect(plotter.Spec.Flows[0].SubFlows[0].Steps).To(gomega.HaveLen(1))
	g.Expect(plotter.Spec.Flows[0].SubFlows[0].Steps[0]).To(gomega.HaveLen(1))
	step := plotter.Spec.Flows[0].SubFlows[0].Steps[0][0]
	g.Expect(step.Parameters.Actions).To(gomega.HaveLen(1))
	filterAction, found := step.Parameters.Actions[0].AdditionalProperties.Items["FilterAction"]
	g.Expect(found).To(gomega.Equal(true))
	filterActionInterface := filterAction.(map[string]interface{})
	g.Expect(filterActionInterface["query"]).To(gomega.Equal("Country == 'UK'"))
}

// This test checks that a non-supported data store does not prevent a plotter from being created
func TestReadyAssetAfterUnsupported(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data = []v1alpha1.DataContext{
		{
			DataSetID:    "s3/deny-dataset",
			Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
		},
		{
			DataSetID:    "local/redact-dataset",
			Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
		},
		{
			DataSetID:    "s3/allow-dataset",
			Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
		},
	}
	application.SetGeneration(1)
	application.SetUID("7")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")

	// check Deny states
	g.Expect(application.Status.AssetStates["s3/deny-dataset"].Conditions[DenyConditionIndex].Status).
		To(gomega.BeIdenticalTo(corev1.ConditionTrue))
	g.Expect(application.Status.AssetStates["local/redact-dataset"].Conditions[DenyConditionIndex].Status).
		To(gomega.BeIdenticalTo(corev1.ConditionTrue))
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
}

// This test checks the case where data comes from another region, and should be redacted.
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
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0] = v1alpha1.DataContext{
		DataSetID:    "s3-external/redact-dataset",
		Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
	}
	application.SetGeneration(1)
	application.SetUID("8")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/copy-csv-parquet.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	copyModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")
	// Create storage account
	dummySecret := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-theshire.yaml", dummySecret)).NotTo(gomega.HaveOccurred())
	dummySecret.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), dummySecret)).NotTo(gomega.HaveOccurred())
	account := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	account.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), account)).NotTo(gomega.HaveOccurred())

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	// check provisioned storage
	g.Expect(application.Status.ProvisionedStorage["s3-external/redact-dataset"].DatasetRef).ToNot(gomega.BeEmpty(), "No storage provisioned")
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &v1alpha1.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(plotter.Spec.Flows).To(gomega.HaveLen(1))
	subflow0 := plotter.Spec.Flows[0].SubFlows[0]
	subflow1 := plotter.Spec.Flows[0].SubFlows[1]
	g.Expect(subflow0.Steps).To(gomega.HaveLen(1))
	g.Expect(subflow0.Steps[0]).To(gomega.HaveLen(1))
	g.Expect(subflow0.Steps[0][0].Cluster).To(gomega.Equal("neverland-cluster"))
	g.Expect(subflow1.Steps).To(gomega.HaveLen(1))
	g.Expect(subflow1.Steps[0]).To(gomega.HaveLen(1))
	g.Expect(subflow1.Steps[0][0].Cluster).To(gomega.Equal("thegreendragon"))
}

// This test checks the ingest scenario - copy is required, no workload specified.
// Two storage accounts are created. Data cannot be stored in one of them according to governance policies.
func TestCopyData(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	assetName := "s3-external/allow-theshire"
	namespaced := types.NamespacedName{
		Name:      "ingest",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/ingest.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0].DataSetID = assetName
	application.Spec.Data[0].Flow = taxonomy.CopyFlow
	application.SetGeneration(1)
	application.SetUID("9")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	copyModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")
	// Create storage accounts
	secret1 := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-neverland.yaml", secret1)).NotTo(gomega.HaveOccurred())
	secret1.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), secret1)).NotTo(gomega.HaveOccurred())
	account1 := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-neverland.yaml", account1)).NotTo(gomega.HaveOccurred())
	account1.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), account1)).NotTo(gomega.HaveOccurred())
	secret2 := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-theshire.yaml", secret2)).NotTo(gomega.HaveOccurred())
	secret2.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), secret2)).NotTo(gomega.HaveOccurred())
	account2 := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account2)).NotTo(gomega.HaveOccurred())
	account2.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), account2)).NotTo(gomega.HaveOccurred())

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")

	// check provisioned storage
	g.Expect(application.Status.ProvisionedStorage[assetName].DatasetRef).ToNot(gomega.BeEmpty(), "No storage provisioned")
	g.Expect(application.Status.ProvisionedStorage[assetName].SecretRef.Name).To(gomega.Equal("credentials-theshire"),
		"Incorrect storage was selected")
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &v1alpha1.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// There should be a single copy module
	g.Expect(plotter.Spec.Assets).To(gomega.HaveLen(2)) // two assets. original + copy
	g.Expect(plotter.Spec.Flows).To(gomega.HaveLen(1))
	g.Expect(plotter.Spec.Flows[0].SubFlows).To(gomega.HaveLen(1))
	subflow := plotter.Spec.Flows[0].SubFlows[0]
	g.Expect(subflow.Triggers).To(gomega.ContainElements(v1alpha1.InitTrigger))
	g.Expect(subflow.FlowType).To(gomega.Equal(taxonomy.CopyFlow))
	g.Expect(subflow.Steps).To(gomega.HaveLen(1))
	g.Expect(subflow.Steps[0]).To(gomega.HaveLen(1))
	g.Expect(subflow.Steps[0][0].Parameters.Arguments[0].AssetID).To(gomega.Equal("s3-external/allow-theshire"))
	g.Expect(subflow.Steps[0][0].Parameters.Arguments[1].AssetID).To(gomega.Equal("s3-external/allow-theshire-copy"))
}

// This test checks the ingest scenario
// A storage account has been defined for the region where the dataset can not be written to according to governance policies.
// An error is received.
func TestCopyDataNotAllowed(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	assetName := "s3-external/deny-theshire"
	namespaced := types.NamespacedName{
		Name:      "ingest",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/ingest.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0].DataSetID = assetName
	application.Spec.Data[0].Flow = taxonomy.CopyFlow
	application.SetGeneration(1)
	application.SetUID("10")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	copyModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")

	// Create storage account
	dummySecret := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-theshire.yaml", dummySecret)).NotTo(gomega.HaveOccurred())
	dummySecret.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), dummySecret)).NotTo(gomega.HaveOccurred())
	account := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	account.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), account)).NotTo(gomega.HaveOccurred())

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	// check provisioned storage
	g.Expect(application.Status.ProvisionedStorage).To(gomega.BeEmpty())
	// Expect Deny condition
	cond := application.Status.AssetStates[assetName].Conditions[DenyConditionIndex]
	g.Expect(cond.Status).To(gomega.BeIdenticalTo(corev1.ConditionTrue), "Deny condition is not set")
	g.Expect(cond.Message).To(gomega.ContainSubstring(v1alpha1.WriteNotAllowed))
	g.Expect(application.Status.Ready).To(gomega.BeTrue())
}

// This test checks the ingest scenario
// A storage account has been defined for the region where the dataset can not be written to according to restrictions on cost
// An error is received.
func TestStorageCost(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	assetName := "s3-external/allow-dataset"
	namespaced := types.NamespacedName{
		Name:      "ingest",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/ingest.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0].DataSetID = assetName
	application.SetGeneration(1)
	application.SetUID("storage-cost")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	copyModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")

	// Create storage account
	dummySecret := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-neverland.yaml", dummySecret)).NotTo(gomega.HaveOccurred())
	dummySecret.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), dummySecret)).NotTo(gomega.HaveOccurred())
	account := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-neverland.yaml", account)).NotTo(gomega.HaveOccurred())
	account.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), account)).NotTo(gomega.HaveOccurred())

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	// check provisioned storage
	g.Expect(application.Status.ProvisionedStorage).To(gomega.BeEmpty())
	// check errors
	g.Expect(getErrorMessages(application)).NotTo(gomega.BeEmpty())
}

// This test checks that the plotter state propagates into the fybrikapp state
func TestPlotterUpdate(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0] = v1alpha1.DataContext{
		DataSetID:    "s3/allow-dataset",
		Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
	}
	application.SetGeneration(1)
	application.SetUID("11")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.Background(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	g.Expect(application.Status.Generated.AppVersion).To(gomega.Equal(application.Generation))
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &v1alpha1.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	// mark the plotter as in error state
	errorMsg := "failure to orchestrate modules"
	plotter.Status.ObservedState.Error = errorMsg
	g.Expect(cl.Update(context.Background(), plotter)).NotTo(gomega.HaveOccurred())

	// the new reconcile should update the application state
	newReq := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: req.Namespace, Name: "plotter_" + req.Name}}
	_, err = r.Reconcile(context.Background(), newReq)
	g.Expect(err).To(gomega.BeNil())
	err = cl.Get(context.Background(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(getErrorMessages(application)).To(gomega.ContainSubstring(errorMsg))

	// mark the plotter as ready
	plotter.Status.ObservedState.Error = ""
	plotter.Status.ObservedState.Ready = true
	g.Expect(cl.Update(context.Background(), plotter)).NotTo(gomega.HaveOccurred())

	// the new reconcile should update the application state
	_, err = r.Reconcile(context.Background(), newReq)
	g.Expect(err).To(gomega.BeNil())
	err = cl.Get(context.Background(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(application.Status.Ready).To(gomega.BeTrue())
}

// This test checks that the older plotter state does not propagate into the fybrikapp state
func TestSyncWithPlotter(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "notebook",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/fybrikcopyapp-csv.yaml", application)).NotTo(gomega.HaveOccurred())
	// imitate a ready phase for the earlier generation
	application.SetGeneration(2)
	application.SetUID("12")
	application.Finalizers = []string{"TestReconciler.finalizer"}
	controllerNamespace := utils.GetControllerNamespace()
	fmt.Printf("FybrikApplication unit test: controller namespace " + controllerNamespace)
	application.Status.Generated = &v1alpha1.ResourceReference{Name: "plotter", Namespace: controllerNamespace, Kind: "Plotter", AppVersion: 1}
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

	plotter := &v1alpha1.Plotter{}
	g.Expect(readObjectFromFile("../../testdata/plotter.yaml", plotter)).NotTo(gomega.HaveOccurred())
	plotter.Status.ObservedState.Ready = true
	plotter.Namespace = controllerNamespace
	g.Expect(cl.Create(context.Background(), plotter)).NotTo(gomega.HaveOccurred())

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	newApp := &v1alpha1.FybrikApplication{}
	err = cl.Get(context.Background(), req.NamespacedName, newApp)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(getErrorMessages(newApp)).NotTo(gomega.BeEmpty())
	g.Expect(newApp.Status.Ready).NotTo(gomega.BeTrue())
}

// This test checks that an empty fybrikapplication can be created and reconciled
func TestFybrikApplicationWithNoDatasets(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "notebook",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/fybrikcopyapp-csv.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data = []v1alpha1.DataContext{}
	application.SetGeneration(1)
	application.SetUID("13")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	res, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(res).To(gomega.BeEquivalentTo(ctrl.Result{}))
	// The application should be in Ready state
	newApp := &v1alpha1.FybrikApplication{}
	err = cl.Get(context.Background(), req.NamespacedName, newApp)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(getErrorMessages(newApp)).To(gomega.BeEmpty())
	g.Expect(newApp.Status.Ready).To(gomega.BeTrue())
}

func TestFybrikApplicationWithInvalidAppInfo(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)

	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "application-with-errors",
		Namespace: "default",
	}

	filename := "../../testdata/unittests/fybrikapplication-appInfoErrors.yaml"
	fybrikApp := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile(filename, fybrikApp)).NotTo(gomega.HaveOccurred())
	fybrikApp.SetGeneration(1)
	fybrikApp.SetUID("14")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		fybrikApp,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	newApp := &v1alpha1.FybrikApplication{}
	err = cl.Get(context.Background(), req.NamespacedName, newApp)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(getErrorMessages(newApp)).NotTo(gomega.BeEmpty())
	g.Expect(newApp.Status.Ready).NotTo(gomega.BeTrue())
}

func TestFybrikApplicationWithInvalidInterface(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)

	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "application-with-errors-2",
		Namespace: "default",
	}
	filename := "../../testdata/unittests/fybrikapplication-interfaceErrors.yaml"
	fybrikApp := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile(filename, fybrikApp)).NotTo(gomega.HaveOccurred())
	fybrikApp.SetGeneration(1)
	fybrikApp.SetUID("15")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		fybrikApp,
	}
	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}
	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	newApp := &v1alpha1.FybrikApplication{}
	err = cl.Get(context.Background(), req.NamespacedName, newApp)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(getErrorMessages(newApp)).NotTo(gomega.BeEmpty())
	g.Expect(newApp.Status.Ready).NotTo(gomega.BeTrue())
}

// This test checks the ingest scenario - copy is required, no workload specified.
// Two copy modules exist. One of them has an incorrect structure (sinks and sources in different capabilities)
func TestCopyModule(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	assetName := "s3-external/allow-theshire"
	namespaced := types.NamespacedName{
		Name:      "ingest",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/ingest.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0].DataSetID = assetName
	application.SetGeneration(1)
	application.SetUID("16")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	invalidModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", invalidModule)).NotTo(gomega.HaveOccurred())
	invalidModule.Namespace = utils.GetControllerNamespace()
	invalidModule.Name = "copy-module-with-invalid-structure"
	capability := invalidModule.Spec.Capabilities[0]
	sink := capability.SupportedInterfaces[0].Sink
	source := capability.SupportedInterfaces[0].Source
	capability.SupportedInterfaces = []v1alpha1.ModuleInOut{{Source: source}, {Sink: sink}}
	invalidModule.Spec.Capabilities[0] = capability
	g.Expect(cl.Create(context.TODO(), invalidModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")

	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	copyModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), copyModule)).NotTo(gomega.HaveOccurred(), "the copy module could not be created")

	// Create storage account
	secret := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-theshire.yaml", secret)).NotTo(gomega.HaveOccurred())
	secret.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), secret)).NotTo(gomega.HaveOccurred())
	account := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	account.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), account)).NotTo(gomega.HaveOccurred())

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")

	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &v1alpha1.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

func TestReadAndTransform(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0] = v1alpha1.DataContext{
		DataSetID:    "s3/redact-dataset",
		Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
	}
	application.SetGeneration(1)
	application.SetUID("17")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	transformModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-transform.yaml", transformModule)).NotTo(gomega.HaveOccurred())
	transformModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), transformModule)).NotTo(gomega.HaveOccurred(), "the transform module could not be created")

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(getErrorMessages(application)).To(gomega.BeEmpty())
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	endpoint := application.Status.AssetStates[application.Spec.Data[0].DataSetID].Endpoint
	g.Expect(endpoint).To(gomega.Not(gomega.BeNil()))
	connectionMap := endpoint.AdditionalProperties.Items
	g.Expect(connectionMap).To(gomega.HaveKey("fybrik-arrow-flight"))
	config := connectionMap["fybrik-arrow-flight"].(map[string]interface{})
	g.Expect(config["hostname"]).To(gomega.Equal("arrow-flight-transform"))
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &v1alpha1.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(plotter.Spec.Assets).To(gomega.HaveLen(1))
	g.Expect(plotter.Spec.Templates).To(gomega.HaveLen(2)) // expect two templates
	g.Expect(plotter.Spec.Flows).To(gomega.HaveLen(1))
	g.Expect(plotter.Spec.Flows[0].SubFlows).To(gomega.HaveLen(1))
	g.Expect(plotter.Spec.Flows[0].SubFlows[0].Steps[0]).To(gomega.HaveLen(2))
}

func TestWriteUnregisteredAsset(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-write-test",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/fybrikapplication-write-AssetNotExist.yaml",
		application)).NotTo(gomega.HaveOccurred())
	application.SetGeneration(1)
	application.SetUID("18")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readWriteModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-write.yaml", readWriteModule)).NotTo(gomega.HaveOccurred())
	readWriteModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), readWriteModule)).NotTo(gomega.HaveOccurred(), "the write module could not be created")

	// Create storage account
	secret := &corev1.Secret{}
	g.Expect(readObjectFromFile("../../testdata/unittests/credentials-theshire.yaml", secret)).NotTo(gomega.HaveOccurred())
	secret.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), secret)).NotTo(gomega.HaveOccurred())
	account := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	account.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.Background(), account)).NotTo(gomega.HaveOccurred())

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(getErrorMessages(application)).To(gomega.BeEmpty())
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	writeEndpoint := application.Status.AssetStates[application.Spec.Data[0].DataSetID].Endpoint
	g.Expect(writeEndpoint).To(gomega.Not(gomega.BeNil()))
	writeConnectionMap := writeEndpoint.AdditionalProperties.Items
	g.Expect(writeConnectionMap).To(gomega.HaveKey("fybrik-arrow-flight"))
	writeDataConfig := writeConnectionMap["fybrik-arrow-flight"].(map[string]interface{})
	g.Expect(writeDataConfig["hostname"]).To(gomega.Equal("read-write-module"))

	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())

	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &v1alpha1.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(plotter.Spec.Assets).To(gomega.HaveLen(1))
	g.Expect(plotter.Spec.Assets["s3-not-exists/new-dataset"].DataStore.Connection.Name).To(gomega.Equal(v1alpha1.S3))
	g.Expect(plotter.Spec.Assets["s3-not-exists/new-dataset"].DataStore.Format).ToNot(gomega.BeEmpty())
	g.Expect(plotter.Spec.Templates).To(gomega.HaveLen(1))
}

func TestWriteRegisteredAsset(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-write-test",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/fybrikapplication-write-AssetExists.yaml", application)).NotTo(gomega.HaveOccurred())
	application.SetGeneration(1)
	application.SetUID("18")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readWriteModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-write.yaml", readWriteModule)).NotTo(gomega.HaveOccurred())
	readWriteModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), readWriteModule)).NotTo(gomega.HaveOccurred(), "the write module could not be created")

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(getErrorMessages(application)).To(gomega.BeEmpty())
	// check plotter creation
	g.Expect(application.Status.AssetStates).To(gomega.HaveLen(2))
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	readOriginalDatalEndpoint := application.Status.AssetStates[application.Spec.Data[0].DataSetID].Endpoint
	g.Expect(readOriginalDatalEndpoint).To(gomega.Not(gomega.BeNil()))
	readOriginalConnectionMap := readOriginalDatalEndpoint.AdditionalProperties.Items
	g.Expect(readOriginalConnectionMap).To(gomega.HaveKey("fybrik-arrow-flight"))
	readOriginalDataConfig := readOriginalConnectionMap["fybrik-arrow-flight"].(map[string]interface{})
	g.Expect(readOriginalDataConfig["hostname"]).To(gomega.Equal("read-write-module"))

	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	readNewDataEndpoint := application.Status.AssetStates[application.Spec.Data[2].DataSetID].Endpoint
	g.Expect(readNewDataEndpoint).To(gomega.Not(gomega.BeNil()))
	readNewConnectionMap := readNewDataEndpoint.AdditionalProperties.Items
	g.Expect(readNewConnectionMap).To(gomega.HaveKey("fybrik-arrow-flight"))
	readNewDataConfig := readNewConnectionMap["fybrik-arrow-flight"].(map[string]interface{})
	g.Expect(readNewDataConfig["hostname"]).To(gomega.Equal("read-write-module"))

	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &v1alpha1.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(plotter.Spec.Assets).To(gomega.HaveLen(2))
	g.Expect(plotter.Spec.Templates).To(gomega.HaveLen(2)) // expect two templates: one for read and one for write
}

func TestWriteAndTransform(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "write-test",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/write_asset.yaml", application)).NotTo(gomega.HaveOccurred())
	application.SetGeneration(1)
	application.SetUID("19")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Write module
	writeModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-write.yaml", writeModule)).NotTo(gomega.HaveOccurred())
	writeModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), writeModule)).NotTo(gomega.HaveOccurred(), "the write module could not be created")
	transformModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-transform.yaml", transformModule)).NotTo(gomega.HaveOccurred())
	transformModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), transformModule)).NotTo(gomega.HaveOccurred(), "the transform module could not be created")

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(getErrorMessages(application)).To(gomega.BeEmpty())
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	endpoint := application.Status.AssetStates[application.Spec.Data[0].DataSetID].Endpoint
	g.Expect(endpoint).To(gomega.Not(gomega.BeNil()))
	connectionMap := endpoint.AdditionalProperties.Items
	g.Expect(connectionMap).To(gomega.HaveKey("fybrik-arrow-flight"))
	config := connectionMap["fybrik-arrow-flight"].(map[string]interface{})
	g.Expect(config["hostname"]).To(gomega.Equal("arrow-flight-transform"))
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &v1alpha1.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	bytes, _ := yaml.Marshal(&plotter)
	fmt.Println("WriteAndTransform:\n " + string(bytes))
	g.Expect(plotter.Spec.Assets).To(gomega.HaveLen(1))
	g.Expect(plotter.Spec.Templates).To(gomega.HaveLen(2)) // expect two templates
	g.Expect(plotter.Spec.Flows).To(gomega.HaveLen(1))
	g.Expect(plotter.Spec.Flows[0].SubFlows).To(gomega.HaveLen(1))
	g.Expect(plotter.Spec.Flows[0].SubFlows[0].Steps[0]).To(gomega.HaveLen(2))
}

func TestWriteWithoutPermissions(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "write-test",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/write_asset.yaml", application)).NotTo(gomega.HaveOccurred())
	application.SetGeneration(1)
	application.SetUID("20")
	application.Spec.Data[0] = v1alpha1.DataContext{
		DataSetID:    "s3/deny-dataset",
		Flow:         taxonomy.WriteFlow,
		Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Write module
	writeModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-write.yaml", writeModule)).NotTo(gomega.HaveOccurred())
	writeModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), writeModule)).NotTo(gomega.HaveOccurred(), "the write module could not be created")

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	// Expect Deny condition
	cond := application.Status.AssetStates["s3/deny-dataset"].Conditions[DenyConditionIndex]
	g.Expect(cond.Status).To(gomega.BeIdenticalTo(corev1.ConditionTrue), "Deny condition is not set")
	g.Expect(cond.Message).To(gomega.ContainSubstring(v1alpha1.WriteNotAllowed))
	g.Expect(application.Status.Ready).To(gomega.BeTrue())
}

func TestReadChain(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	namespaced := types.NamespacedName{
		Name:      "read-test",
		Namespace: "default",
	}
	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/data-usage.yaml", application)).NotTo(gomega.HaveOccurred())
	application.Spec.Data[0] = v1alpha1.DataContext{
		DataSetID:    "s3/redact-dataset",
		Requirements: v1alpha1.DataRequirements{Interface: &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight}},
	}
	application.SetGeneration(1)
	application.SetUID("21")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// Read module
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	readModule.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), readModule)).NotTo(gomega.HaveOccurred(), "the read module could not be created")
	transformModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-transform.yaml", transformModule)).NotTo(gomega.HaveOccurred())
	transformModule.Namespace = utils.GetControllerNamespace()
	transformModule.Spec.Capabilities[0].Capability = "read"
	transformModule.Spec.Capabilities[0].Scope = "workload"
	g.Expect(cl.Create(context.TODO(), transformModule)).NotTo(gomega.HaveOccurred(), "the transform module could not be created")

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(getErrorMessages(application)).To(gomega.BeEmpty())
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
	endpoint := application.Status.AssetStates[application.Spec.Data[0].DataSetID].Endpoint
	g.Expect(endpoint).To(gomega.Not(gomega.BeNil()))
	connectionMap := endpoint.AdditionalProperties.Items
	g.Expect(connectionMap).To(gomega.HaveKey("fybrik-arrow-flight"))
	config := connectionMap["fybrik-arrow-flight"].(map[string]interface{})
	g.Expect(config["hostname"]).To(gomega.Equal("arrow-flight-transform"))
	plotterObjectKey := types.NamespacedName{
		Namespace: application.Status.Generated.Namespace,
		Name:      application.Status.Generated.Name,
	}
	plotter := &v1alpha1.Plotter{}
	err = cl.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(plotter.Spec.Templates).To(gomega.HaveLen(2)) // expect two templates
}

// TestEmptyInterface checks requests that do not expect a response
// Delete flow, as an example
func TestEmptyInterface(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	application := &v1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/unittests/delete-flow.yaml", application)).NotTo(gomega.HaveOccurred())
	application.SetGeneration(1)
	application.SetUID("22")
	// Objects to track in the fake client.
	objs := []runtime.Object{
		application,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)

	// module
	module := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-delete.yaml", module)).NotTo(gomega.HaveOccurred())
	module.Namespace = utils.GetControllerNamespace()
	g.Expect(cl.Create(context.TODO(), module)).NotTo(gomega.HaveOccurred(), "the read module could not be created")

	// Create a FybrikApplicationReconciler object with the scheme and fake client.
	r := createTestFybrikApplicationController(cl, s)
	g.Expect(r).NotTo(gomega.BeNil())
	namespaced := types.NamespacedName{
		Name:      application.Name,
		Namespace: application.Namespace,
	}

	req := reconcile.Request{
		NamespacedName: namespaced,
	}

	_, err := r.Reconcile(context.Background(), req)
	g.Expect(err).To(gomega.BeNil())

	err = cl.Get(context.TODO(), req.NamespacedName, application)
	g.Expect(err).To(gomega.BeNil(), "Cannot fetch fybrikapplication")
	g.Expect(getErrorMessages(application)).To(gomega.BeEmpty())
	// check plotter creation
	g.Expect(application.Status.Generated).ToNot(gomega.BeNil())
}
