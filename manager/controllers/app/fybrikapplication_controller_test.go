// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
)

const timeout = time.Second * 1000
const interval = time.Millisecond * 100

var _ = Describe("FybrikApplication Controller", func() {
	Context("FybrikApplication", func() {

		controllerNamespace := utils.GetControllerNamespace()
		fmt.Printf("FybrikApplication: controller namespace: %s\n", controllerNamespace)

		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
			module := &apiv1alpha1.FybrikModule{}
			Expect(readObjectFromFile("../../testdata/e2e/module-read.yaml", module)).ToNot(HaveOccurred())
			module.Namespace = controllerNamespace
			application := &apiv1alpha1.FybrikApplication{}

			Expect(readObjectFromFile("../../testdata/e2e/fybrikapplication.yaml", application)).ToNot(HaveOccurred())
			_ = k8sClient.Delete(context.Background(), application)
			_ = k8sClient.Delete(context.Background(), module)
		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
		})
		It("Test restricted access to secrets", func() {
			if os.Getenv("USE_EXISTING_CONTROLLER") != "true" {
				// test access restriction: only secrets from blueprints namespace can be accessed
				// Create secrets in default and fybrik-blueprints namespaces
				// A secret from the default namespace should not be listed
				secret1 := &corev1.Secret{Type: corev1.SecretTypeOpaque, StringData: map[string]string{"password": "123"}}
				secret1.Name = "test-secret"
				secret1.Namespace = "default"
				Expect(k8sClient.Create(context.TODO(), secret1)).NotTo(HaveOccurred(), "a secret could not be created")
				secret2 := &corev1.Secret{Type: corev1.SecretTypeOpaque, StringData: map[string]string{"password": "123"}}
				secret2.Name = "test-secret"

				modulesNamespace := utils.GetDefaultModulesNamespace()
				fmt.Printf("Application test using data access module namespace: %s\n", modulesNamespace)
				secret2.Namespace = modulesNamespace
				Expect(k8sClient.Create(context.TODO(), secret2)).NotTo(HaveOccurred(), "a secret could not be created")
				secretList := &corev1.SecretList{}
				Expect(k8sClient.List(context.Background(), secretList)).NotTo(HaveOccurred())
				Expect(len(secretList.Items)).To(Equal(1), "Secrets from other namespaces should not be listed")
			}
		})
		It("Test restricted access to modules", func() {
			if os.Getenv("USE_EXISTING_CONTROLLER") != "true" {
				// test access restriction: only modules from the control plane can be accessed
				// Create a module in default namespace
				// An attempt to fetch it will fail
				module := &apiv1alpha1.FybrikModule{}
				Expect(readObjectFromFile("../../testdata/e2e/module-read.yaml", module)).ToNot(HaveOccurred())
				module.Namespace = "default"
				Expect(k8sClient.Create(context.Background(), module)).Should(Succeed())
				fetchedModule := &apiv1alpha1.FybrikModule{}
				moduleKey := client.ObjectKeyFromObject(module)
				Expect(k8sClient.Get(context.Background(), moduleKey, fetchedModule)).To(HaveOccurred(), "Should deny access")
			}
		})
		It("Test end-to-end for FybrikApplication", func() {
			connector := os.Getenv("USE_MOCKUP_CONNECTOR")
			fmt.Printf("Connector:  %s\n", connector)
			if len(connector) > 0 && connector != "true" {
				Skip("Skipping test when not running with mockup connector!")
			}
			module := &apiv1alpha1.FybrikModule{}
			Expect(readObjectFromFile("../../testdata/e2e/module-read.yaml", module)).ToNot(HaveOccurred())
			moduleKey := client.ObjectKeyFromObject(module)
			module.Namespace = controllerNamespace
			application := &apiv1alpha1.FybrikApplication{}

			Expect(readObjectFromFile("../../testdata/e2e/fybrikapplication.yaml", application)).ToNot(HaveOccurred())
			application.Labels = map[string]string{"label1": "foo", "label2": "bar"}
			applicationKey := client.ObjectKeyFromObject(application)
			fmt.Printf("Module:  %v\n", module.Namespace)
			fmt.Printf("Application:  %v\n", application.Namespace)
			// Create FybrikApplication and FybrikModule
			Expect(k8sClient.Create(context.Background(), module)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), application)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				application := &apiv1alpha1.FybrikApplication{ObjectMeta: metav1.ObjectMeta{Namespace: applicationKey.Namespace, Name: applicationKey.Name}}
				_ = k8sClient.Get(context.Background(), applicationKey, application)
				// _ = k8sClient.Delete(context.Background(), application)
				module := &apiv1alpha1.FybrikApplication{ObjectMeta: metav1.ObjectMeta{Namespace: moduleKey.Namespace, Name: moduleKey.Name}}
				_ = k8sClient.Get(context.Background(), moduleKey, module)
				// _ = k8sClient.Delete(context.Background(), module)
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
			By("Expecting plotter to be fetchable")
			Eventually(func() error {
				return k8sClient.Get(context.Background(), plotterObjectKey, plotter)
			}, timeout, interval).Should(Succeed())

			By("Expect plotter to be ready at some point")
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), plotterObjectKey, plotter)).To(Succeed())
				return plotter.Status.ObservedState.Ready
			}, timeout*10, interval).Should(BeTrue(), "plotter is not ready")

			blueprintObjectKey := client.ObjectKey{Namespace: plotter.Namespace, Name: plotter.Name}
			By("Expecting Blueprint to contain application labels")
			blueprint := &apiv1alpha1.Blueprint{}
			Eventually(func() error {
				return k8sClient.Get(context.Background(), blueprintObjectKey, blueprint)
			}, timeout, interval).Should(Succeed(), "Blueprint has not been created")
			Expect(blueprint.Spec.ModulesNamespace).To(Equal(utils.GetDefaultModulesNamespace()))

			for _, module := range blueprint.Spec.Modules {
				Expect(module.Arguments.Labels["label1"]).To(Equal("foo"))
				Expect(module.Arguments.Labels["label2"]).To(Equal("bar"))
				Expect(module.Arguments.Labels[apiv1alpha1.ApplicationNameLabel]).To(Equal(applicationKey.Name))
				Expect(module.Arguments.Labels[apiv1alpha1.ApplicationNamespaceLabel]).To(Equal(applicationKey.Namespace))
				Expect(module.Arguments.AppSelector.MatchLabels["app"]).To(Equal("notebook"))
			}
			By("Expecting FybrikApplication to eventually be ready")
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), applicationKey, application)).To(Succeed())
				return application.Status.Ready
			}, timeout, interval).Should(BeTrue(), "FybrikApplication is not ready after timeout!")

			By("Status should contain the details of the endpoint")
			Expect(len(application.Status.AssetStates)).To(Equal(1))
			fqdn := "test-app-e2e-default-read-module-test-e2e." + blueprint.Spec.ModulesNamespace
			connection := application.Status.AssetStates["s3/redact-dataset"].Endpoint
			Expect(connection).ToNot(BeNil())
			connectionMap := connection.AdditionalProperties.Items
			Expect(connectionMap).To(HaveKey("fybrik-arrow-flight"))
			config := connectionMap["fybrik-arrow-flight"].(map[string]interface{})
			Expect(config["hostname"]).To(Equal(fqdn))
			Expect(config["scheme"]).To(Equal("grpc"))
		})
	})
})
