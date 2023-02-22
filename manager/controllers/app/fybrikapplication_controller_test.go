// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/environment"
)

const timeout = time.Second * 1000
const interval = time.Millisecond * 100

var _ = Describe("FybrikApplication Controller", func() {
	Context("FybrikApplication", func() {

		controllerNamespace := environment.GetControllerNamespace()
		fmt.Printf("FybrikApplication: controller namespace: %s\n", controllerNamespace)

		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
			module := &fapp.FybrikModule{}
			Expect(readObjectFromFile("../../testdata/e2e/module-read.yaml", module)).ToNot(HaveOccurred())
			module.Namespace = controllerNamespace
			application := &fapp.FybrikApplication{}

			Expect(readObjectFromFile("../../testdata/e2e/fybrikapplication.yaml", application)).ToNot(HaveOccurred())
			_ = k8sClient.Delete(context.Background(), application)
			_ = k8sClient.Delete(context.Background(), module)
			Expect(readObjectFromFile("../../testdata/e2e/productionApp.yaml", application)).ToNot(HaveOccurred())
			_ = k8sClient.Delete(context.Background(), application)
		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
		})
		It("Test restricted access to secrets", func() {
			if os.Getenv("USE_EXISTING_CONTROLLER") != "true" {
				// test access restriction: only secrets from system namespace can be accessed
				// Create secrets in default and system namespaces
				// A secret from the default namespace should not be listed
				secret1 := &corev1.Secret{Type: corev1.SecretTypeOpaque, StringData: map[string]string{"password": "123"}}
				secret1.Name = "test-secret"
				secret1.Namespace = "default"
				Expect(k8sClient.Create(context.TODO(), secret1)).NotTo(HaveOccurred(), "a secret could not be created")
				secret2 := &corev1.Secret{Type: corev1.SecretTypeOpaque, StringData: map[string]string{"password": "123"}}
				secret2.Name = "test-secret"

				secret2.Namespace = environment.GetSystemNamespace()
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
				module := &fapp.FybrikModule{}
				Expect(readObjectFromFile("../../testdata/e2e/module-read.yaml", module)).ToNot(HaveOccurred())
				module.Namespace = "default"
				Expect(k8sClient.Create(context.Background(), module)).Should(Succeed())
				fetchedModule := &fapp.FybrikModule{}
				moduleKey := client.ObjectKeyFromObject(module)
				Expect(k8sClient.Get(context.Background(), moduleKey, fetchedModule)).To(HaveOccurred(), "Should deny access")
			}
		})
		// test end to end run
		// test how policy change affects the data plane construction
		// the new policy requires copy for prod application which will fail the data plane construction
		It("Test end-to-end for FybrikApplication", func() {
			connector := os.Getenv("USE_MOCKUP_CONNECTOR")
			fmt.Printf("Connector:  %s\n", connector)
			if len(connector) > 0 && connector != "true" {
				Skip("Skipping test when not running with mockup connector!")
			}
			if os.Getenv("USE_EXISTING_CONTROLLER") != "true" {
				Skip("Skipping test when running locally")
			}
			module := &fapp.FybrikModule{}
			Expect(readObjectFromFile("../../testdata/e2e/module-read.yaml", module)).ToNot(HaveOccurred())
			module.Namespace = controllerNamespace
			application := &fapp.FybrikApplication{}
			prodApplication := &fapp.FybrikApplication{}
			Expect(readObjectFromFile("../../testdata/e2e/productionApp.yaml", prodApplication)).ToNot(HaveOccurred())
			origApplication := prodApplication.DeepCopy()
			prodAppKey := client.ObjectKeyFromObject(prodApplication)
			Expect(readObjectFromFile("../../testdata/e2e/fybrikapplication.yaml", application)).ToNot(HaveOccurred())
			application.Labels = map[string]string{"label1": "foo", "label2": "bar"}
			applicationKey := client.ObjectKeyFromObject(application)
			// Create FybrikApplication and FybrikModule
			Expect(k8sClient.Create(context.Background(), module)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), application)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), prodApplication)).Should(Succeed())

			By("Expecting applications to be created")
			Eventually(func() bool {
				return (k8sClient.Get(context.Background(), applicationKey, application) == nil) &&
					(k8sClient.Get(context.Background(), prodAppKey, prodApplication) == nil)
			}, timeout, interval).Should(BeTrue())
			By("Expecting plotters to be constructed")
			Eventually(func() *fapp.ResourceReference {
				_ = k8sClient.Get(context.Background(), applicationKey, application)
				return application.Status.Generated
			}, timeout, interval).ShouldNot(BeNil())
			Eventually(func() *fapp.ResourceReference {
				_ = k8sClient.Get(context.Background(), prodAppKey, prodApplication)
				return prodApplication.Status.Generated
			}, timeout, interval).ShouldNot(BeNil())

			// The plotter has to be created
			plotter := &fapp.Plotter{}
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
			blueprint := &fapp.Blueprint{}
			Eventually(func() error {
				return k8sClient.Get(context.Background(), blueprintObjectKey, blueprint)
			}, timeout, interval).Should(Succeed(), "Blueprint has not been created")
			Expect(blueprint.Spec.ModulesNamespace).To(Equal(environment.GetDefaultModulesNamespace()))

			Expect(blueprint.Labels["label1"]).To(Equal("foo"))
			Expect(blueprint.Labels["label2"]).To(Equal("bar"))
			Expect(utils.GetApplicationNameFromLabels(blueprint.Labels)).To(Equal(applicationKey.Name))
			Expect(utils.GetApplicationNamespaceFromLabels(blueprint.Labels)).To(Equal(applicationKey.Namespace))
			Expect(blueprint.Spec.Application.WorkloadSelector.MatchLabels["app"]).To(Equal("notebook"))
			Expect(blueprint.Spec.Application.Context.Items["intent"].(string)).To(Equal("Fraud Detection"))
			By("Expecting FybrikApplication to eventually be ready")
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), applicationKey, application)).To(Succeed())
				return application.Status.Ready
			}, timeout, interval).Should(BeTrue(), "FybrikApplication is not ready after timeout!")

			By("Connector messages should be propagated to the status")
			Expect(len(application.Status.AssetStates)).To(Equal(3))
			Expect(application.Status.AssetStates["s3/redact-dataset"].Conditions[ReadyConditionIndex].Message).To(BeEmpty())
			Expect(application.Status.AssetStates["s3-incomplete/allow-dataset"].Conditions[ReadyConditionIndex].Message).NotTo(BeEmpty())
			Expect(application.Status.AssetStates["s3-external/new-dataset"].Conditions[ReadyConditionIndex].Message).NotTo(BeEmpty())
			By("Status should contain the details of the endpoint")
			fqdn := "test-app-e2e-default-read-module-test-e2e." + blueprint.Spec.ModulesNamespace
			connection := application.Status.AssetStates["s3/redact-dataset"].Endpoint
			Expect(connection).ToNot(BeNil())
			connectionMap := connection.AdditionalProperties.Items
			Expect(connectionMap).To(HaveKey("fybrik-arrow-flight"))
			config := connectionMap["fybrik-arrow-flight"].(map[string]interface{})
			Expect(config["hostname"]).To(Equal(fqdn))
			Expect(config["scheme"]).To(Equal("grpc"))

			By("Changing a policy for PROD application")
			_ = k8sClient.Delete(context.Background(), origApplication)
			// change policy
			cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:      "fybrik-adminconfig",
				Namespace: controllerNamespace,
			}}
			Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(cm), cm)).Should(Succeed())
			bytes, err := os.ReadFile("../../testdata/e2e/new_policy.rego")
			Expect(err).To(BeNil())
			cm.Data["new_policy.rego"] = string(bytes)
			Expect(k8sClient.Update(context.Background(), cm)).Should(Succeed())
			By("Waiting for a change to be propagated")
			time.Sleep(2 * time.Minute)
			*prodApplication = *origApplication
			By("Re-applying PROD application")
			Expect(k8sClient.Create(context.Background(), prodApplication)).Should(Succeed())
			Eventually(func() error {
				return k8sClient.Get(context.Background(), prodAppKey, prodApplication)
			}, timeout, interval).Should(Succeed())
			By("Expecting PROD application to fail")
			Eventually(func() string {
				_ = k8sClient.Get(context.Background(), prodAppKey, prodApplication)
				return getErrorMessages(prodApplication)
			}, timeout, interval).ShouldNot(BeEmpty())
			By("Restoring the policies")
			delete(cm.Data, "new_policy.rego")
			Expect(k8sClient.Update(context.Background(), cm)).Should(Succeed())
		})
	})
})
