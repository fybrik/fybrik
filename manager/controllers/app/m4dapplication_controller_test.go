// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const timeout = time.Second * 30
const interval = time.Millisecond * 100

var _ = Describe("FybrikApplication Controller", func() {
	Context("FybrikApplication", func() {
		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
		})
		It("Test end-to-end for FybrikApplication", func() {
			module := &app.FybrikModule{}
			Expect(readObjectFromFile("../../testdata/e2e/module-read.yaml", module)).ToNot(HaveOccurred())
			moduleKey := client.ObjectKeyFromObject(module)
			application := &app.FybrikApplication{}
			Expect(readObjectFromFile("../../testdata/e2e/m4dapplication.yaml", application)).ToNot(HaveOccurred())
			applicationKey := client.ObjectKeyFromObject(application)

			// Create FybrikApplication and FybrikModule
			Expect(k8sClient.Create(context.Background(), module)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), application)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				application := &apiv1alpha1.FybrikApplication{ObjectMeta: metav1.ObjectMeta{Namespace: applicationKey.Namespace, Name: applicationKey.Name}}
				_ = k8sClient.Get(context.Background(), applicationKey, application)
				_ = k8sClient.Delete(context.Background(), application)
				module := &apiv1alpha1.FybrikApplication{ObjectMeta: metav1.ObjectMeta{Namespace: moduleKey.Namespace, Name: moduleKey.Name}}
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

			By("Expecting FybrikApplication to eventually be ready")
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), applicationKey, application)).To(Succeed())
				return application.Status.Ready
			}, timeout, interval).Should(BeTrue(), "FybrikApplication is not ready after timeout!")

			By("Status should contain the details of the endpoint")
			Expect(len(application.Status.ReadEndpointsMap)).To(Equal(1))
			fqdn := "test-app-e2e-default-read-module-test-e2e-e24d69b99a.fybrik-blueprints.svc.cluster.local"
			Expect(application.Status.ReadEndpointsMap["s3/redact-dataset"]).To(Equal(apiv1alpha1.EndpointSpec{
				Hostname: fqdn,
				Port:     80,
				Scheme:   "grpc",
			}))
		})
	})
})
