// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package motion

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/onsi/gomega/gexec"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	motionv1 "github.com/mesh-for-data/mesh-for-data/manager/apis/motion/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var mgr ctrl.Manager
var testEnv *envtest.Environment
var noSimulatedProgress bool

// This is the entry method for the ginko testing framework suite.
// The actual test are in the following files:
// - batchtransfer_controller_test.go
// - streamtransfer_controller_test.go
func TestMotionAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

// Before a test suite is run a K8s test environment can be set up.
// There are multiple flags that can be set to run the simulation test
// - USE_EXISTING_CLUSTER: (true/false)
//   This variable controls if an existing K8s cluster should be used or not.
//   If not testEnv will spin up an artificial environment that includes a local etcd setup.
// - NO_SIMULATED_PROGRESS: (true/false)
//   This variable can be used by tests that can manually simulate progress of e.g. jobs or pods.
//   e.g. the simulated test environment from testEnv does not progress pods etc while when testing against
//   an external Kubernetes cluster this will actually run pods.
// - USE_EXISTING_CONTROLLER: (true/false)
//   This setting controls if a controller should be set up and run by this test suite or if an external one
//   should be used. E.g. in integration tests running against an existing setup a controller is already existing
//   in the Kubernetes cluster and should not be started by the test as two controllers competing may influence the test.
var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	By("bootstrapping test environment")
	if os.Getenv("NO_SIMULATED_PROGRESS") == "true" {
		noSimulatedProgress = true
	}
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "charts", "m4d-crd", "templates")},
		ErrorIfCRDPathMissing: true,
		//AttachControlPlaneOutput: true,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = motionv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	if os.Getenv("USE_EXISTING_CONTROLLER") == "true" {
		logf.Log.Info("Using existing controller in existing cluster...")
		k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	} else {
		workerNamespaceSelector := fields.SelectorFromSet(fields.Set{"metadata.namespace": "m4d-blueprints"})
		selectorsByObject := cache.SelectorsByObject{
			&motionv1.BatchTransfer{}:       {Field: workerNamespaceSelector},
			&motionv1.StreamTransfer{}:      {Field: workerNamespaceSelector},
			&kbatch.Job{}:                   {Field: workerNamespaceSelector},
			&corev1.Secret{}:                {Field: workerNamespaceSelector},
			&corev1.Pod{}:                   {Field: workerNamespaceSelector},
			&corev1.PersistentVolumeClaim{}: {Field: workerNamespaceSelector},
		}

		mgr, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme:             scheme.Scheme,
			MetricsBindAddress: "localhost:8085",
			NewCache:           cache.BuilderWithOptions(cache.Options{SelectorsByObject: selectorsByObject}),
		})
		Expect(err).ToNot(HaveOccurred())

		err = NewBatchTransferReconciler(mgr, "BatchTransfer").SetupWithManager(mgr)
		Expect(err).ToNot(HaveOccurred())

		err = NewStreamTransferReconciler(mgr, "StreamTransfer").SetupWithManager(mgr)
		Expect(err).ToNot(HaveOccurred())

		go func() {
			err = mgr.Start(ctrl.SetupSignalHandler())
			Expect(err).ToNot(HaveOccurred())
		}()

		k8sClient = mgr.GetClient()
		err = k8sClient.Create(context.Background(), &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "m4d-blueprints",
			},
		})
		Expect(err).ToNot(HaveOccurred())
	}

	Expect(k8sClient).ToNot(BeNil())

	// k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	// Expect(err).ToNot(HaveOccurred())
	// Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	gexec.KillAndWait(5 * time.Second)
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
