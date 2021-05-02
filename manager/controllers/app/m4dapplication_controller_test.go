// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"io/ioutil"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	apiv1alpha1 "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const timeout = time.Second * 30
const interval = time.Millisecond * 100
const ReadPathModuleName = "read-path"

func createModules() {
	readModule := &apiv1alpha1.M4DModule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ReadPathModuleName,
			Namespace: utils.GetSystemNamespace(),
		},
		Spec: apiv1alpha1.M4DModuleSpec{
			Flows: []apiv1alpha1.ModuleFlow{apiv1alpha1.Read},
			Capabilities: apiv1alpha1.Capability{
				SupportedInterfaces: []apiv1alpha1.ModuleInOut{
					{
						Flow:   apiv1alpha1.Read,
						Source: &apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet},
					},
				},
				API: &apiv1alpha1.ModuleAPI{
					InterfaceDetails: apiv1alpha1.InterfaceDetails{
						Protocol:   apiv1alpha1.ArrowFlight,
						DataFormat: apiv1alpha1.Arrow,
					},
					Endpoint: apiv1alpha1.EndpointSpec{
						Hostname: "arrow-flight",
						Port:     80,
						Scheme:   "grpc",
					},
				},
				Actions: []apiv1alpha1.SupportedAction{
					{
						ID:    "redact-ID",
						Level: pb.EnforcementAction_COLUMN,
					},
				},
			},
			Chart: apiv1alpha1.ChartSpec{
				Name: "localhost:5000/m4d-system/m4d-template:0.1.0",
			},
		},
	}
	_ = k8sClient.Create(context.Background(), readModule)
}

func deleteModules() {
	_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DStorageAccount{ObjectMeta: metav1.ObjectMeta{Name: "read-path", Namespace: utils.GetSystemNamespace()}})
}

var _ = Describe("M4DApplication Controller", func() {
	Context("M4DApplication", func() {
		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
			createModules()
		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
			// delete modules
			deleteModules()
		})
		It("Test end-to-end for M4DApplication", func() {
			var err error
			applicationYAML, err := ioutil.ReadFile("../../testdata/e2e/m4dapplication.yaml")
			Expect(err).ToNot(HaveOccurred())
			application := &apiv1alpha1.M4DApplication{}
			err = yaml.Unmarshal(applicationYAML, application)
			Expect(err).ToNot(HaveOccurred())

			applicationKey, err := client.ObjectKeyFromObject(application)
			Expect(err).ToNot(HaveOccurred())

			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), application)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				f := &apiv1alpha1.M4DApplication{ObjectMeta: metav1.ObjectMeta{Namespace: applicationKey.Namespace, Name: applicationKey.Name}}
				_ = k8sClient.Get(context.Background(), applicationKey, f)
				_ = k8sClient.Delete(context.Background(), f)
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
			fqdn := "notebook-default-read-path-ffea578653.m4d-blueprints.svc.cluster.local"
			Expect(application.Status.ReadEndpointsMap["asset_id-xxx-catalog_id-s3"]).To(Equal(apiv1alpha1.EndpointSpec{
				Hostname: fqdn,
				Port:     80,
				Scheme:   "grpc",
			}))
		})
	})
})
