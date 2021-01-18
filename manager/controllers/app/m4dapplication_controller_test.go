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

	apiv1alpha1 "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const timeout = time.Second * 30
const interval = time.Millisecond * 100

// GetStorageSignature returns a signature of M4DBucket
func GetStorageSignature() types.NamespacedName {
	return types.NamespacedName{Name: "available-bucket", Namespace: "default"}
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

// CreateReadPathModule creates a read-path module
func CreateReadPathModule() *apiv1alpha1.M4DModule {
	return &apiv1alpha1.M4DModule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "read-path",
			Namespace: "default",
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
			},
			Chart: apiv1alpha1.ChartSpec{
				Name: "s3-flight",
			},
		},
	}
}

// CreateKafkaToS3CopyModule creates a copy module kafka->s3
func CreateKafkaToS3CopyModule() *apiv1alpha1.M4DModule {
	return &apiv1alpha1.M4DModule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "implicit-copy-kafka-to-s3-stream",
			Namespace: "default",
		},
		Spec: apiv1alpha1.M4DModuleSpec{
			Flows: []apiv1alpha1.ModuleFlow{apiv1alpha1.Copy},
			Capabilities: apiv1alpha1.Capability{
				CredentialsManagedBy: apiv1alpha1.SecretProvider,
				SupportedInterfaces: []apiv1alpha1.ModuleInOut{
					{
						Flow:   apiv1alpha1.Copy,
						Source: &apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.Kafka, DataFormat: apiv1alpha1.JSON},
						Sink:   &apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet},
					},
				},
			},
			Chart: apiv1alpha1.ChartSpec{
				Name: "xxx",
			},
		},
	}
}

// CreateDb2ToS3CopyModule creates a copy module that copies db2 data to s3
func CreateDb2ToS3CopyModule() *apiv1alpha1.M4DModule {
	db2Module := &apiv1alpha1.M4DModule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "implicit-copy-db2-to-s3",
			Namespace: "default",
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
				Actions: make([]pb.EnforcementAction, 2),
			},
			Chart: apiv1alpha1.ChartSpec{
				Name: "db2-chart",
			},
		},
	}
	// define actions
	db2Module.Spec.Capabilities.Actions[0] = pb.EnforcementAction{Name: "redact", Id: "redact-ID", Level: pb.EnforcementAction_COLUMN}
	db2Module.Spec.Capabilities.Actions[1] = pb.EnforcementAction{Name: "encrypt", Id: "encrypt-ID", Level: pb.EnforcementAction_COLUMN}
	return db2Module
}

// CreateS3ToS3CopyModule creates a copy module s3->s3
func CreateS3ToS3CopyModule() *apiv1alpha1.M4DModule {
	s3Module := &apiv1alpha1.M4DModule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "implicit-copy-s3-to-s3",
			Namespace: "default",
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
				Actions: make([]pb.EnforcementAction, 2),
			},
			Chart: apiv1alpha1.ChartSpec{
				Name: "s3-s3",
			},
		},
	}
	s3Module.Spec.Capabilities.Actions[0] = pb.EnforcementAction{Name: "redact", Id: "redact-ID", Level: pb.EnforcementAction_COLUMN}
	s3Module.Spec.Capabilities.Actions[1] = pb.EnforcementAction{Name: "encrypt", Id: "encrypt-ID", Level: pb.EnforcementAction_COLUMN}
	return s3Module
}

var _ = Describe("M4DApplication Controller", func() {
	Context("M4DApplication", func() {
		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
			// delete storage
			storageSignature := GetStorageSignature()
			_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DBucket{ObjectMeta: metav1.ObjectMeta{Name: storageSignature.Name, Namespace: storageSignature.Namespace}})
			// delete modules
			_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DModule{ObjectMeta: metav1.ObjectMeta{Name: "implicit-copy-kafka-to-s3-stream", Namespace: "default"}})
			_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DModule{ObjectMeta: metav1.ObjectMeta{Name: "read-path", Namespace: "default"}})
			_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DModule{ObjectMeta: metav1.ObjectMeta{Name: "implicit-copy-db2-to-s3", Namespace: "default"}})
			_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DModule{ObjectMeta: metav1.ObjectMeta{Name: "implicit-copy-s3-to-s3", Namespace: "default"}})
		})

		// This test checks that the finalizers are properly reconciled upon creation and deletion of the resource
		// M4DBucket is used to demonstrate freeing owned objects

		// Assumptions on response from connectors:
		// Db2 dataset, will be received in parquet format - thus, a copy is required.
		// Transformations are required
		// Copy is required, thus a storage for destination is allocated
		// This M4DApplication will become an owner of a storage-defining resource

		It("Test reconcileFinalizers", func() {
			// Create M4DBucket
			storageSignature := GetStorageSignature()
			storage := &apiv1alpha1.M4DBucket{
				ObjectMeta: metav1.ObjectMeta{
					Name:      storageSignature.Name,
					Namespace: storageSignature.Namespace,
				},
				Spec: apiv1alpha1.M4DBucketSpec{
					Name:      "test-bucket",
					Endpoint:  "xxx",
					VaultPath: "yyy",
				},
			}
			Expect(k8sClient.Create(context.Background(), storage)).Should(Succeed())

			// Load db2-s3 copy module
			copyModule := CreateDb2ToS3CopyModule()
			readPathModule := CreateReadPathModule()
			Expect(k8sClient.Create(context.Background(), copyModule)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), readPathModule)).Should(Succeed())

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

			By("Expecting ownership on a bucket to be removed")
			_ = k8sClient.Get(context.Background(), storageSignature, storage)
			Expect(storage.Status.Owners).To(BeEmpty())

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
			Expect(k8sClient.Create(context.Background(), CreateReadPathModule())).Should(Succeed())

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
			module := CreateReadPathModule()
			Expect(k8sClient.Create(context.Background(), module)).Should(Succeed())

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
		// Db2 dataset, a copy is required.
		// S3 dataset, no copy is needed
		// Enforcement action for both operations and datasets: Allow
		// Applied one copy module (kafka->s3), not the one we need for the first data set
		// Result: an error
		It("Test wrong-copy-module", func() {

			// Load kafka-s3 copy module
			module := CreateKafkaToS3CopyModule()
			readPathModule := CreateReadPathModule()
			Expect(k8sClient.Create(context.Background(), module)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), readPathModule)).Should(Succeed())

			appSignature := types.NamespacedName{Name: "wrong-copy", Namespace: "default"}
			resource := InitM4DApplication(appSignature.Name, 2)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID:    "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"db2\"}",
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
			_ = k8sClient.Delete(context.Background(), module)

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
			// allocate storage
			storageSignature := GetStorageSignature()
			storage := &apiv1alpha1.M4DBucket{
				ObjectMeta: metav1.ObjectMeta{
					Name:      storageSignature.Name,
					Namespace: storageSignature.Namespace,
				},
				Spec: apiv1alpha1.M4DBucketSpec{
					Name:      "test-bucket",
					Endpoint:  "xxx",
					VaultPath: "yyy",
				},
			}
			Expect(k8sClient.Create(context.Background(), storage)).Should(Succeed())

			// Load s3-s3 copy module
			s3Module := CreateS3ToS3CopyModule()
			readPathModule := CreateReadPathModule()

			Expect(k8sClient.Create(context.Background(), s3Module)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), readPathModule)).Should(Succeed())

			// Load db2-s3 copy module
			db2Module := CreateDb2ToS3CopyModule()
			Expect(k8sClient.Create(context.Background(), db2Module)).Should(Succeed())

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

			By("Expecting a namespace to be allocated")
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
			// There should be a single read module with two datasets
			Expect(len(plotter.Spec.Blueprints)).To(Equal(1))
			blueprint := plotter.Spec.Blueprints["US-cluster"]
			Expect(blueprint).NotTo(BeNil())
			numReads := 0
			for _, step := range blueprint.Flow.Steps {
				if step.Template == readPathModule.Name {
					numReads++
					Expect(len(step.Arguments.Read)).To(Equal(2))
				}
			}
			Expect(numReads).To(Equal(1))
		})
		It("Test multiple-blueprints", func() {
			// allocate storage
			storageSignature := GetStorageSignature()
			storage := &apiv1alpha1.M4DBucket{
				ObjectMeta: metav1.ObjectMeta{
					Name:      storageSignature.Name,
					Namespace: storageSignature.Namespace,
				},
				Spec: apiv1alpha1.M4DBucketSpec{
					Name:      "test-bucket",
					Endpoint:  "xxx",
					VaultPath: "yyy",
				},
			}
			Expect(k8sClient.Create(context.Background(), storage)).Should(Succeed())

			// Load s3-s3 copy module
			s3Module := CreateS3ToS3CopyModule()
			readPathModule := CreateReadPathModule()

			Expect(k8sClient.Create(context.Background(), s3Module)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), readPathModule)).Should(Succeed())

			// Load db2-s3 copy module
			db2Module := CreateDb2ToS3CopyModule()
			Expect(k8sClient.Create(context.Background(), db2Module)).Should(Succeed())

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
			// allocate storage
			storageSignature := GetStorageSignature()
			storage := &apiv1alpha1.M4DBucket{
				ObjectMeta: metav1.ObjectMeta{
					Name:      storageSignature.Name,
					Namespace: storageSignature.Namespace,
				},
				Spec: apiv1alpha1.M4DBucketSpec{
					Name:      "newdata-bucket",
					Endpoint:  "xxx",
					VaultPath: "yyy",
				},
			}
			Expect(k8sClient.Create(context.Background(), storage)).Should(Succeed())

			// Load s3-s3 copy module
			s3Module := CreateS3ToS3CopyModule()
			Expect(k8sClient.Create(context.Background(), s3Module)).Should(Succeed())

			// Load db2-s3 copy module
			db2Module := CreateDb2ToS3CopyModule()
			Expect(k8sClient.Create(context.Background(), db2Module)).Should(Succeed())

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

			// Ensure getting cleaned up after tests finish
			defer func() {
				DeleteM4DApplication(appSignature.Name)
			}()

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
				if step.Template == s3Module.Name {
					moduleMatch = true
				}
			}
			Expect(numSteps).To(Equal(1))
			Expect(moduleMatch).To(Equal(true))
			DeleteM4DApplication(appSignature.Name)
		})
	})

	Context("M4DApplication e2e", func() {
		BeforeEach(func() {
			// Add any teardown steps that needs to be executed after each test
			// delete modules
			_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DModule{ObjectMeta: metav1.ObjectMeta{Name: "arrow-flight-module", Namespace: "default"}})
		})

		It("Test end-to-end for M4DApplication with arrow-flight module", func() {
			var err error
			readModuleYAML, err := ioutil.ReadFile("../../testdata/e2e/module-read.yaml")
			Expect(err).ToNot(HaveOccurred())
			module := &apiv1alpha1.M4DModule{}
			err = yaml.Unmarshal(readModuleYAML, module)
			Expect(err).ToNot(HaveOccurred())

			moduleKey, err := client.ObjectKeyFromObject(module)
			Expect(err).ToNot(HaveOccurred())

			// Create M4DModule
			Expect(k8sClient.Create(context.Background(), module)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				f := &apiv1alpha1.M4DModule{ObjectMeta: metav1.ObjectMeta{Namespace: moduleKey.Namespace, Name: moduleKey.Name}}
				_ = k8sClient.Get(context.Background(), moduleKey, f)
				_ = k8sClient.Delete(context.Background(), f)
			}()

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
			By("Expect plotter to be created")
			Eventually(func() error {
				return k8sClient.Get(context.Background(), plotterObjectKey, plotter)
			}, timeout, interval).Should(Succeed(), "Could not find Plotter "+application.Status.Generated.Namespace+"/"+application.Status.Generated.Name)
			By("Expect plotter to be ready at some point")
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), plotterObjectKey, plotter)).To(Succeed())
				return plotter.Status.ObservedState.Ready
			}, timeout*10, interval).Should(BeTrue(), "plotter is not ready")

			// Extra long timeout as deploying the arrow-flight module on a new cluster may take some time
			// depending on the download speed
			By("Expecting M4DApplication to eventually be ready")
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), applicationKey, application)).To(Succeed())
				return application.Status.Ready
			}, timeout*10, interval).Should(BeTrue(), "M4DApplication is not ready after timeout!")
		})
	})
})
