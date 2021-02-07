// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	comv1alpha1 "github.com/IBM/dataset-lifecycle-framework/src/dataset-operator/pkg/apis/com/v1alpha1"
	apiv1alpha1 "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const timeout = time.Second * 30
const interval = time.Millisecond * 100

func allocateStorageAccounts() {
	dummySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dummy-secret",
			Namespace: utils.GetSystemNamespace(),
		},
		Data: map[string][]byte{"accessKeyID": []byte("value1"), "secretAccessKey": []byte("value2")},
		Type: "Opaque",
	}
	_ = k8sClient.Create(context.Background(), dummySecret)
	accountUS := &apiv1alpha1.M4DStorageAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "account1",
			Namespace: utils.GetSystemNamespace(),
		},
		Spec: apiv1alpha1.M4DStorageAccountSpec{
			Endpoint:  "http://endpoint1",
			SecretRef: "dummy-secret",
			Regions:   []string{"US"},
		},
	}
	accountGermany := &apiv1alpha1.M4DStorageAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "account2",
			Namespace: utils.GetSystemNamespace(),
		},
		Spec: apiv1alpha1.M4DStorageAccountSpec{
			Endpoint:  "http://endpoint2",
			SecretRef: "dummy-secret",
			Regions:   []string{"Germany"},
		},
	}

	_ = k8sClient.Create(context.Background(), accountUS)
	_ = k8sClient.Create(context.Background(), accountGermany)
}

func deleteStorageAccounts() {
	_ = k8sClient.Delete(context.Background(), &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "dummy-secret", Namespace: utils.GetSystemNamespace()}})
	_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DStorageAccount{ObjectMeta: metav1.ObjectMeta{Name: "account1", Namespace: utils.GetSystemNamespace()}})
	_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DStorageAccount{ObjectMeta: metav1.ObjectMeta{Name: "account2", Namespace: utils.GetSystemNamespace()}})
}

func createModules() {
	readModule := &apiv1alpha1.M4DModule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "read-path",
			Namespace: utils.GetSystemNamespace(),
		},
		Spec: apiv1alpha1.M4DModuleSpec{
			Flows: []apiv1alpha1.ModuleFlow{apiv1alpha1.Read},
			Capabilities: apiv1alpha1.Capability{
				CredentialsManagedBy: apiv1alpha1.SecretProvider,
				SupportedInterfaces: []apiv1alpha1.ModuleInOut{
					{
						Flow:   apiv1alpha1.Read,
						Source: &apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet},
					},
				},
				API: &apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.ArrowFlight, DataFormat: apiv1alpha1.Arrow},
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
	db2Module := &apiv1alpha1.M4DModule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "implicit-copy-db2-to-s3",
			Namespace: utils.GetSystemNamespace(),
		},
		Spec: apiv1alpha1.M4DModuleSpec{
			Flows: []apiv1alpha1.ModuleFlow{apiv1alpha1.Copy},
			Capabilities: apiv1alpha1.Capability{
				CredentialsManagedBy: apiv1alpha1.SecretProvider,
				SupportedInterfaces: []apiv1alpha1.ModuleInOut{
					{
						Source: &apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.JdbcDb2, DataFormat: apiv1alpha1.Table},
						Sink:   &apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet},
						Flow:   apiv1alpha1.Copy,
					},
				},
				Actions: []apiv1alpha1.SupportedAction{
					{
						ID:    "redact-ID",
						Level: pb.EnforcementAction_COLUMN,
					},
					{
						ID:    "encrypt-ID",
						Level: pb.EnforcementAction_COLUMN,
					},
				},
			},
			Chart: apiv1alpha1.ChartSpec{
				Name: "db2-chart",
			},
		},
	}
	s3Module := &apiv1alpha1.M4DModule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "implicit-copy-s3-to-s3",
			Namespace: utils.GetSystemNamespace(),
		},
		Spec: apiv1alpha1.M4DModuleSpec{
			Flows: []apiv1alpha1.ModuleFlow{apiv1alpha1.Copy},
			Capabilities: apiv1alpha1.Capability{
				CredentialsManagedBy: apiv1alpha1.Automatic,
				SupportedInterfaces: []apiv1alpha1.ModuleInOut{
					{
						Flow:   apiv1alpha1.Copy,
						Source: &apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet},
						Sink:   &apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet},
					},
				},
				Actions: []apiv1alpha1.SupportedAction{
					{
						ID:    "redact-ID",
						Level: pb.EnforcementAction_COLUMN,
					},
					{
						ID:    "encrypt-ID",
						Level: pb.EnforcementAction_COLUMN,
					},
				},
			},
			Chart: apiv1alpha1.ChartSpec{
				Name: "s3-s3",
			},
		},
	}
	_ = k8sClient.Create(context.Background(), readModule)
	_ = k8sClient.Create(context.Background(), db2Module)
	_ = k8sClient.Create(context.Background(), s3Module)
}

func deleteModules() {
	_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DModule{ObjectMeta: metav1.ObjectMeta{Name: "implicit-copy-s3-to-s3", Namespace: utils.GetSystemNamespace()}})
	_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DStorageAccount{ObjectMeta: metav1.ObjectMeta{Name: "implicit-copy-db2-to-s3", Namespace: utils.GetSystemNamespace()}})
	_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DStorageAccount{ObjectMeta: metav1.ObjectMeta{Name: "read-path", Namespace: utils.GetSystemNamespace()}})
}

// InitM4DApplication creates an empty resource with n cataloged data sets
func InitM4DApplication(name string, n int) *apiv1alpha1.M4DApplication {
	appSignature := types.NamespacedName{Name: name, Namespace: "default"}
	labels := map[string]string{"key": "value"}
	return &apiv1alpha1.M4DApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appSignature.Name,
			Namespace: appSignature.Namespace,
		},
		Spec: apiv1alpha1.M4DApplicationSpec{
			Selector: apiv1alpha1.Selector{ClusterName: "US-cluster", WorkloadSelector: metav1.LabelSelector{MatchLabels: labels}},
			Data:     make([]apiv1alpha1.DataContext, n),
		},
	}
}

// InitM4DApplicationWithoutWorkload creates an empty resource with no workload reference
func InitM4DApplicationWithoutWorkload(name string, n int) *apiv1alpha1.M4DApplication {
	appSignature := types.NamespacedName{Name: name, Namespace: "default"}
	return &apiv1alpha1.M4DApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appSignature.Name,
			Namespace: appSignature.Namespace,
		},
		Spec: apiv1alpha1.M4DApplicationSpec{
			Selector: apiv1alpha1.Selector{ClusterName: "US-cluster"},
			Data:     make([]apiv1alpha1.DataContext, n),
		},
	}
}

func DeleteM4DApplication(name string) {
	// delete application
	appSignature := types.NamespacedName{Name: name, Namespace: "default"}
	resource := &apiv1alpha1.M4DApplication{ObjectMeta: metav1.ObjectMeta{Name: appSignature.Name, Namespace: appSignature.Namespace}}
	_ = k8sClient.Delete(context.Background(), resource)

	Eventually(func() error {
		f := &apiv1alpha1.M4DApplication{}
		return k8sClient.Get(context.Background(), appSignature, f)
	}, timeout, interval).ShouldNot(Succeed())
}

var _ = Describe("M4DApplication Controller", func() {
	Context("M4DApplication", func() {
		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
			allocateStorageAccounts()
			createModules()
		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
			// delete storage
			deleteStorageAccounts()
			// delete modules
			deleteModules()
			time.Sleep(interval)
		})

		// This test checks that the finalizers are properly reconciled upon creation and deletion of the resource
		// M4DBucket is used to demonstrate freeing owned objects

		// Assumptions on response from connectors:
		// Db2 dataset, will be received in parquet format - thus, a copy is required.
		// Transformations are required
		// Copy is required, thus a storage for destination is allocated
		// This M4DApplication will become an owner of a storage-defining resource

		It("Test reconcileFinalizers", func() {
			appSignature := types.NamespacedName{Name: "with-finalizers", Namespace: "default"}
			resource := InitM4DApplication(appSignature.Name, 1)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID:    "{\"asset_id\": \"123\", \"catalog_id\": \"db2\"}",
				Requirements: apiv1alpha1.DataRequirements{Interface: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet}},
			}
			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())

			By("Expecting reconcilers to be added")
			Eventually(func() []string {
				f := &apiv1alpha1.M4DApplication{}
				_ = k8sClient.Get(context.Background(), appSignature, f)
				return f.Finalizers
			}, timeout, interval).ShouldNot(BeEmpty())

			_ = k8sClient.Get(context.Background(), appSignature, resource)
			_ = k8sClient.Delete(context.Background(), resource)
			By("Expecting reconcilers to be removed")
			Eventually(func() []string {
				f := &apiv1alpha1.M4DApplication{}
				_ = k8sClient.Get(context.Background(), appSignature, f)
				return f.Finalizers
			}, timeout, interval).Should(BeEmpty())

			By("Expecting delete to finish")
			Eventually(func() error {
				f := &apiv1alpha1.M4DApplication{}
				return k8sClient.Get(context.Background(), appSignature, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
		// Tests denial of the access to data

		// Assumptions on response from connectors:
		// S3 dataset, needs to be consumed in the same way
		// Enforcement action for read operation: Deny
		// Result: an error

		It("Test deny-on-read", func() {

			appSignature := types.NamespacedName{Name: "deny-on-read", Namespace: "default"}
			resource := InitM4DApplication(appSignature.Name, 1)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID:    "{\"asset_id\": \"deny-dataset\", \"catalog_id\": \"s3\"}",
				Requirements: apiv1alpha1.DataRequirements{Interface: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet}},
			}

			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				DeleteM4DApplication(appSignature.Name)
			}()

			By("Expecting access denied on read")
			Eventually(func() string {
				f := &apiv1alpha1.M4DApplication{}
				_ = k8sClient.Get(context.Background(), appSignature, f)
				return getErrorMessages(f)
			}, timeout, interval).ShouldNot(BeEmpty())

			_ = k8sClient.Get(context.Background(), appSignature, resource)
			Expect(getErrorMessages(resource)).To(ContainSubstring(apiv1alpha1.ReadAccessDenied))
		})
		// Tests selection of read-path module

		// Assumptions on response from connectors:
		// db2 dataset, will be received in s3/parquet
		// Read module does not have api for s3/parquet
		// Result: an error

		It("Test no-read-path", func() {
			appSignature := types.NamespacedName{Name: "read-expected", Namespace: "default"}
			resource := InitM4DApplication(appSignature.Name, 1)

			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID:    "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"db2\"}",
				Requirements: apiv1alpha1.DataRequirements{Interface: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet}},
			}
			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				DeleteM4DApplication(appSignature.Name)
			}()

			By("Expecting an error")
			Eventually(func() string {
				f := &apiv1alpha1.M4DApplication{}
				_ = k8sClient.Get(context.Background(), appSignature, f)
				return getErrorMessages(f)
			}, timeout, interval).ShouldNot(BeEmpty())

			_ = k8sClient.Get(context.Background(), appSignature, resource)
			Expect(getErrorMessages(resource)).To(ContainSubstring(apiv1alpha1.ModuleNotFound))
			Expect(getErrorMessages(resource)).To(ContainSubstring("read"))
		})

		// Tests denial of the necessary copy operation

		// Assumptions on response from connectors:
		// s3 dataset
		// Read module with source=s3,parquet exists
		// Copy to s3 is required by the user
		// Enforcement action for write operation: Deny
		// Result: an error

		It("Test deny-on-copy", func() {
			appSignature := types.NamespacedName{Name: "with-copy", Namespace: "default"}
			resource := InitM4DApplication(appSignature.Name, 1)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID: "{\"asset_id\": \"deny-on-copy\", \"catalog_id\": \"s3\"}",
				Requirements: apiv1alpha1.DataRequirements{
					Interface: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.ArrowFlight, DataFormat: apiv1alpha1.Arrow},
					Copy:      apiv1alpha1.CopyRequirements{Required: true},
				},
			}

			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				DeleteM4DApplication(appSignature.Name)
			}()

			By("Expecting an error")
			Eventually(func() string {
				f := &apiv1alpha1.M4DApplication{}
				_ = k8sClient.Get(context.Background(), appSignature, f)
				return getErrorMessages(f)
			}, timeout, interval).ShouldNot(BeEmpty())

			_ = k8sClient.Get(context.Background(), appSignature, resource)
			Expect(getErrorMessages(resource)).To(ContainSubstring(apiv1alpha1.CopyNotAllowed))
		})

		// Tests finding a module for copy

		// Assumptions on response from connectors:
		// Two datasets:
		// Kafka dataset, a copy is required.
		// S3 dataset, no copy is needed
		// Enforcement action for both operations and datasets: Allow
		// No copy module (kafka->s3)
		// Result: an error
		It("Test wrong-copy-module", func() {
			appSignature := types.NamespacedName{Name: "wrong-copy", Namespace: "default"}
			resource := InitM4DApplication(appSignature.Name, 2)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID:    "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"kafka\"}",
				Requirements: apiv1alpha1.DataRequirements{Interface: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.ArrowFlight, DataFormat: apiv1alpha1.Arrow}},
			}
			resource.Spec.Data[1] = apiv1alpha1.DataContext{
				DataSetID:    "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"s3\"}",
				Requirements: apiv1alpha1.DataRequirements{Interface: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.ArrowFlight, DataFormat: apiv1alpha1.Arrow}},
			}

			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())
			// Ensure getting cleaned up after tests finish
			defer func() {
				DeleteM4DApplication(appSignature.Name)
			}()

			By("Expecting an error")
			Eventually(func() string {
				f := &apiv1alpha1.M4DApplication{}
				_ = k8sClient.Get(context.Background(), appSignature, f)
				return getErrorMessages(f)
			}, timeout, interval).ShouldNot(BeEmpty())

			_ = k8sClient.Get(context.Background(), appSignature, resource)
			Expect(getErrorMessages(resource)).To(ContainSubstring(apiv1alpha1.ModuleNotFound))
			Expect(getErrorMessages(resource)).To(ContainSubstring("copy"))
		})
		// Assumptions on response from connectors:
		// Two datasets:
		// Db2 dataset, a copy is required.
		// S3 dataset, no copy is needed
		// Enforcement actions for the first dataset: redact on read, encrypt on copy
		// Enforcement action for the second dataset: Allow
		// Applied copy module s3->s3 and db2->s3 supporting redact and encrypt actions
		// Result: blueprint is created successfully, a read module is applied once for both datasets

		It("Test blueprint-created", func() {
			appSignature := types.NamespacedName{Name: "m4d-test", Namespace: "default"}
			resource := InitM4DApplication(appSignature.Name, 2)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID:    "{\"asset_id\": \"default-dataset\", \"catalog_id\": \"db2\"}",
				Requirements: apiv1alpha1.DataRequirements{Interface: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.ArrowFlight, DataFormat: apiv1alpha1.Arrow}},
			}
			resource.Spec.Data[1] = apiv1alpha1.DataContext{
				DataSetID:    "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"s3\"}",
				Requirements: apiv1alpha1.DataRequirements{Interface: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.ArrowFlight, DataFormat: apiv1alpha1.Arrow}},
			}

			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				DeleteM4DApplication(appSignature.Name)
			}()
			if os.Getenv("USE_EXISTING_CONTROLLER") == "true" {
				By("Expecting a dataset to be allocated")
				Eventually(func() int {
					_ = k8sClient.Get(context.Background(), appSignature, resource)
					return len(resource.Status.ProvisionedStorage)
				}, timeout, interval).Should(Equal(1))
			} else {
				By("Expecting plotter to be generated")
				Eventually(func() *apiv1alpha1.ResourceReference {
					_ = k8sClient.Get(context.Background(), appSignature, resource)
					return resource.Status.Generated
				}, timeout, interval).ShouldNot(BeNil())
				plotter := &apiv1alpha1.Plotter{}
				Eventually(func() error {
					key := types.NamespacedName{Name: resource.Status.Generated.Name, Namespace: resource.Status.Generated.Namespace}
					return k8sClient.Get(context.Background(), key, plotter)
				}, timeout, interval).Should(Succeed())

				// Check the generated blueprint
				// There should be a single read module with two datasets
				Expect(len(plotter.Spec.Blueprints)).To(Equal(1))
				blueprint := plotter.Spec.Blueprints["US-cluster"]
				Expect(blueprint).NotTo(BeNil())
				numReads := 0
				for _, step := range blueprint.Flow.Steps {
					if step.Template == "read-path" {
						numReads++
						Expect(len(step.Arguments.Read)).To(Equal(2))
					}
				}
				Expect(numReads).To(Equal(1))
			}
		})
		It("Test multiple-blueprints", func() {
			appSignature := types.NamespacedName{Name: "multiple-regions", Namespace: "default"}
			resource := InitM4DApplication(appSignature.Name, 1)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID:    "{\"asset_id\": \"default-dataset\", \"catalog_id\": \"s3-external\"}",
				Requirements: apiv1alpha1.DataRequirements{Interface: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.ArrowFlight, DataFormat: apiv1alpha1.Arrow}},
			}
			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())

			// work-around: we don't have currently a setup for multicluster environment in tests
			if os.Getenv("USE_EXISTING_CONTROLLER") == "true" {
				Eventually(func() string {
					f := &apiv1alpha1.M4DApplication{}
					_ = k8sClient.Get(context.Background(), appSignature, f)
					resource = f
					return getErrorMessages(f)
				}, timeout, interval).ShouldNot(BeEmpty())
				Expect(getErrorMessages(resource)).To(ContainSubstring(apiv1alpha1.InvalidClusterConfiguration))
			} else {
				Eventually(func() *apiv1alpha1.ResourceReference {
					f := &apiv1alpha1.M4DApplication{}
					_ = k8sClient.Get(context.Background(), appSignature, f)
					resource = f
					return f.Status.Generated
				}, timeout, interval).ShouldNot(BeNil())
				By("Expecting plotter to be generated")
				plotter := &apiv1alpha1.Plotter{}
				Eventually(func() error {
					key := types.NamespacedName{Name: resource.Status.Generated.Name, Namespace: resource.Status.Generated.Namespace}
					return k8sClient.Get(context.Background(), key, plotter)
				}, timeout, interval).Should(Succeed())
			}
			DeleteM4DApplication(appSignature.Name)
		})

		It("Test new data blueprint created", func() {
			appSignature := types.NamespacedName{Name: "m4d-newdata-test", Namespace: "default"}
			resource := InitM4DApplicationWithoutWorkload(appSignature.Name, 1)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID: "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"s3\"}",
				Requirements: apiv1alpha1.DataRequirements{
					Interface: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet},
					Copy: apiv1alpha1.CopyRequirements{Required: true,
						Catalog: apiv1alpha1.CatalogRequirements{CatalogID: "ingest_test", CatalogService: "Egeria"}}},
			}

			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())
			By("Expecting a dataset to be allocated")
			Eventually(func() int {
				_ = k8sClient.Get(context.Background(), appSignature, resource)
				return len(resource.Status.ProvisionedStorage)
			}, timeout, interval).Should(Equal(1))
			if os.Getenv("USE_EXISTING_CONTROLLER") == "true" {
				for _, info := range resource.Status.ProvisionedStorage {
					dataset := &comv1alpha1.Dataset{}
					Eventually(func() error {
						key := types.NamespacedName{Name: info.DatasetRef, Namespace: utils.GetSystemNamespace()}
						return k8sClient.Get(context.Background(), key, dataset)
					}, timeout, interval).Should(Succeed())
					Expect(dataset.Spec.Local["secret-name"]).To(Equal("dummy-secret"))
					Expect(dataset.Spec.Local["endpoint"]).To(Equal("http://endpoint1"))
					Expect(dataset.Spec.Local["provision"]).To(Equal("true"))
				}
			} else {
				Eventually(func() *apiv1alpha1.ResourceReference {
					_ = k8sClient.Get(context.Background(), appSignature, resource)
					return resource.Status.Generated
				}, timeout, interval).ShouldNot(BeNil())
				By("Expecting plotter to be generated")
				plotter := &apiv1alpha1.Plotter{}
				Eventually(func() error {
					key := types.NamespacedName{Name: resource.Status.Generated.Name, Namespace: resource.Status.Generated.Namespace}
					return k8sClient.Get(context.Background(), key, plotter)
				}, timeout, interval).Should(Succeed())

				// Check the generated blueprint
				// There should be a single copy module
				Expect(len(plotter.Spec.Blueprints)).To(Equal(1))
				blueprint := plotter.Spec.Blueprints["US-cluster"]
				Expect(blueprint).NotTo(BeNil())
				numSteps := 0
				moduleMatch := false
				for _, step := range blueprint.Flow.Steps {
					numSteps++
					if step.Template == "implicit-copy-s3-to-s3" {
						moduleMatch = true
					}
				}
				Expect(numSteps).To(Equal(1))
				Expect(moduleMatch).To(Equal(true))
			}
			DeleteM4DApplication(appSignature.Name)
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
		})
	})
})
