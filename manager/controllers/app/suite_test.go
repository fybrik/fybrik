// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/helm"
	local "fybrik.io/fybrik/pkg/multicluster/local"
	sa "fybrik.io/fybrik/pkg/storage/apis/storageaccount/v1alpha1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client
	mgr       ctrl.Manager
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t,
		"Controller Suite")
}

var _ = BeforeSuite(func() {
	done := make(chan interface{})
	ctx, cancel = context.WithCancel(context.TODO())
	go func() {
		logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

		path, pathErr := os.Getwd()
		if pathErr != nil {
			logf.Log.Info(pathErr.Error())
		}
		By("bootstrapping test environment")
		if os.Getenv("USE_EXISTING_CONTROLLER") == "true" {
			fmt.Printf("Using existing environment; don't load CRDs. \n")
			useexistingcluster := true
			testEnv = &envtest.Environment{
				UseExistingCluster: &useexistingcluster,
			}
		} else {
			fmt.Printf("Using fake environment; so set path to CRDs so they are installed. \n")
			testEnv = &envtest.Environment{
				CRDDirectoryPaths: []string{
					filepath.Join(path, "..", "..", "..", "charts", "fybrik-crd", "templates"),
				},
				ErrorIfCRDPathMissing: true,
			}
		}

		DefaultTestConfiguration(GinkgoT())

		var err error
		cfg, err = testEnv.Start()
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg).ToNot(BeNil())

		err = fapp.AddToScheme(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

		// +kubebuilder:scaffold:scheme

		if os.Getenv("USE_EXISTING_CONTROLLER") == "true" {
			logf.Log.Info("Using existing controller in existing cluster...")
			fmt.Printf("Using existing controller in existing cluster... \n")
			k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
			Expect(err).ToNot(HaveOccurred())
		} else {
			fmt.Printf("Setup fake environment... \n")
			controllerNamespace := environment.GetControllerNamespace()
			modulesNamespace := environment.GetDefaultModulesNamespace()
			fmt.Printf("Suite test: Using controller namespace: %s; using data access module namespace %s\n: ",
				controllerNamespace, modulesNamespace)

			systemNamespaceSelector := fields.SelectorFromSet(fields.Set{"metadata.namespace": environment.GetSystemNamespace()})
			workerNamespaceSelector := fields.SelectorFromSet(fields.Set{"metadata.namespace": environment.GetDefaultModulesNamespace()})
			// the testing environment will restrict access to secrets, modules and storage accounts
			mgr, err = ctrl.NewManager(cfg, ctrl.Options{
				Scheme:             scheme.Scheme,
				MetricsBindAddress: "localhost:8086",
				NewCache: cache.BuilderWithOptions(cache.Options{SelectorsByObject: cache.SelectorsByObject{
					&fapp.FybrikModule{}:       {Field: systemNamespaceSelector},
					&sa.FybrikStorageAccount{}: {Field: systemNamespaceSelector},
					&corev1.Secret{}:           {Field: workerNamespaceSelector},
				}}),
			})
			Expect(err).ToNot(HaveOccurred())

			// Setup application controller
			reconciler := createTestFybrikApplicationController(mgr.GetClient(), mgr.GetScheme())
			Expect(reconciler).NotTo(BeNil())

			err = reconciler.SetupWithManager(mgr)
			Expect(err).ToNot(HaveOccurred())

			// Setup blueprint controller
			fakeHelm := helm.NewFake(
				&release.Release{
					Name: "ra8afad067a6a96084dcb", // Release name is from arrow-flight module
					Info: &release.Info{Status: release.StatusDeployed},
				}, []*unstructured.Unstructured{},
			)
			err = NewBlueprintReconciler(mgr, "Blueprint", fakeHelm).SetupWithManager(mgr)
			Expect(err).ToNot(HaveOccurred())

			// Setup plotter controller
			clusterMgr, err := local.NewClusterManager(mgr.GetClient(), controllerNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterMgr).NotTo(BeNil())
			err = NewPlotterReconciler(mgr, "Plotter", clusterMgr).SetupWithManager(mgr)
			Expect(err).ToNot(HaveOccurred())

			go func() {
				err = mgr.Start(ctx)
				Expect(err).ToNot(HaveOccurred())
			}()

			k8sClient = mgr.GetClient()
			Expect(k8sClient.Create(context.Background(), &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: environment.GetSystemNamespace(),
				},
			}))

			Expect(k8sClient.Create(context.Background(), &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: modulesNamespace,
				},
			}))
		}
		Expect(k8sClient).ToNot(BeNil())
		close(done)
	}()
	Eventually(done, 60).Should(BeClosed())
})

var _ = AfterSuite(func() {
	// see https://github.com/kubernetes-sigs/controller-runtime/issues/1571
	cancel()
	By("tearing down the test environment")
	gexec.KillAndWait(5 * time.Second)
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func SetIfNotSet(key, value string, t GinkgoTInterface) {
	if _, b := os.LookupEnv(key); !b {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Could not set environment variable %s", key)
		}
	}
}

func DefaultTestConfiguration(t GinkgoTInterface) {
	SetIfNotSet(environment.CatalogConnectorServiceAddressKey, "http://localhost:50085", t)
	SetIfNotSet(environment.VaultAddressKey, "http://127.0.0.1:8200/", t)
	SetIfNotSet(environment.EnableWebhooksKey, "false", t)
	SetIfNotSet(environment.MainPolicyManagerConnectorURLKey, "http://localhost:50090", t)
	SetIfNotSet(environment.MainPolicyManagerNameKey, "MOCK", t)
	SetIfNotSet(environment.LoggingVerbosityKey, "-1", t)
	SetIfNotSet(environment.PrettyLoggingKey, "true", t)
	SetIfNotSet(environment.LocalClusterName, "thegreendragon", t)
	SetIfNotSet(environment.LocalRegion, "theshire", t)
	SetIfNotSet(environment.LocalVaultAuthPath, "kind", t)
}
