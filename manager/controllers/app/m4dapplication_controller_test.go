// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"os"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const timeout = time.Second * 30
const interval = time.Millisecond * 100

var _ = Describe("M4DApplication Controller", func() {
	Context("M4DApplication", func() {
		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
		})
		It("Test restricted access to secrets", func() {
			if os.Getenv("USE_EXISTING_CONTROLLER") != "true" {
				// test access restriction: only secrets from m4d-blueprints can be accessed
				// Create secrets in default and m4d-blueprints namespaces
				// A secret from the default namespace should not be listed
				secret1 := &corev1.Secret{Type: corev1.SecretTypeOpaque, StringData: map[string]string{"password": "123"}}
				secret1.Name = "test-secret"
				secret1.Namespace = "default"
				Expect(k8sClient.Create(context.TODO(), secret1)).NotTo(HaveOccurred(), "a secret could not be created")
				secret2 := &corev1.Secret{Type: corev1.SecretTypeOpaque, StringData: map[string]string{"password": "123"}}
				secret2.Name = "test-secret"
				secret2.Namespace = "m4d-blueprints"
				Expect(k8sClient.Create(context.TODO(), secret2)).NotTo(HaveOccurred(), "a secret could not be created")
				secretList := &corev1.SecretList{}
				Expect(k8sClient.List(context.Background(), secretList)).NotTo(HaveOccurred())
				Expect(len(secretList.Items)).To(Equal(1), "Secrets from other namespaces should not be listed")
			}
		})
		It("Test restricted access to modules", func() {
			if os.Getenv("USE_EXISTING_CONTROLLER") != "true" {
				// test access restriction: only modules from m4d-system can be accessed
				// Create a module in default namespace
				// An attempt to fetch it will fail
				module := &app.M4DModule{}
				Expect(readObjectFromFile("../../testdata/e2e/module-read.yaml", module)).ToNot(HaveOccurred())
				module.Namespace = "default"
				Expect(k8sClient.Create(context.Background(), module)).Should(Succeed())
				fetchedModule := &app.M4DModule{}
				moduleKey := client.ObjectKeyFromObject(module)
				Expect(k8sClient.Get(context.Background(), moduleKey, fetchedModule)).To(HaveOccurred(), "Should deny access")
			}
		})
		It("Test end-to-end for M4DApplication", func() {
			module := &app.M4DModule{}
			Expect(readObjectFromFile("../../testdata/e2e/module-read.yaml", module)).ToNot(HaveOccurred())
			moduleKey := client.ObjectKeyFromObject(module)
			application := &app.M4DApplication{}
			Expect(readObjectFromFile("../../testdata/e2e/m4dapplication.yaml", application)).ToNot(HaveOccurred())
			applicationKey := client.ObjectKeyFromObject(application)

			// Create M4DApplication and M4DModule
			Expect(k8sClient.Create(context.Background(), module)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), application)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				application := &apiv1alpha1.M4DApplication{ObjectMeta: metav1.ObjectMeta{Namespace: applicationKey.Namespace, Name: applicationKey.Name}}
				_ = k8sClient.Get(context.Background(), applicationKey, application)
				_ = k8sClient.Delete(context.Background(), application)
				module := &apiv1alpha1.M4DApplication{ObjectMeta: metav1.ObjectMeta{Namespace: moduleKey.Namespace, Name: moduleKey.Name}}
				_ = k8sClient.Get(context.Background(), moduleKey, module)
				_ = k8sClient.Delete(context.Background(), module)
			}()

			By("Expecting application to be created")
			Eventually(func() error {
				return k8sClient.Get(context.Background(), applicationKey, application)
			}, timeout, interval).Should(Succeed())
			By("Expecting plotter to be constructed")
			Eventually(func() *apiv1alpha1.ResourceReference {
				_ = k8sClient.Get(context.Background(), applicationKey, application)
				return application.Status.Generated
			}, timeout, interval).ShouldNot(BeNil())

			// The plotter has to be created
			plotter := &apiv1alpha1.Plotter{}
			plotterObjectKey := client.ObjectKey{Namespace: application.Status.Generated.Namespace, Name: application.Status.Generated.Name}
			Eventually(func() error {
				return k8sClient.Get(context.Background(), plotterObjectKey, plotter)
			}, timeout, interval).Should(Succeed())

			By("Expect plotter to be ready at some point")
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), plotterObjectKey, plotter)).To(Succeed())
				return plotter.Status.ObservedState.Ready
			}, timeout*10, interval).Should(BeTrue(), "plotter is not ready")

			By("Expecting M4DApplication to eventually be ready")
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), applicationKey, application)).To(Succeed())
				return application.Status.Ready
			}, timeout, interval).Should(BeTrue(), "M4DApplication is not ready after timeout!")

			By("Status should contain the details of the endpoint")
			Expect(len(application.Status.ReadEndpointsMap)).To(Equal(1))
			fqdn := "test-app-e2e-default-read-module-test-e2e-e24d69b99a.m4d-blueprints.svc.cluster.local"
			Expect(application.Status.ReadEndpointsMap["s3/redact-dataset"]).To(Equal(apiv1alpha1.EndpointSpec{
				Hostname: fqdn,
				Port:     80,
				Scheme:   "grpc",
			}))
		})
	})
})
