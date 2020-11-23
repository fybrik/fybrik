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
	pb "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// GetApplicationSignature returns a test name + namespace
func GetApplicationSignature() types.NamespacedName {
	return types.NamespacedName{Name: "dma-for-test", Namespace: "default"}
}

// GetStorageSignature returns a signature of M4DBucket
func GetStorageSignature() types.NamespacedName {
	return types.NamespacedName{Name: "available-bucket", Namespace: "default"}
}

// InitM4DApplication creates an empty resource with n data sets
func InitM4DApplication(n int) *apiv1alpha1.M4DApplication {
	appSignature := GetApplicationSignature()
	return &apiv1alpha1.M4DApplication{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appSignature.Name,
			Namespace: appSignature.Namespace,
		},
		Spec: apiv1alpha1.M4DApplicationSpec{AppInfo: apiv1alpha1.ApplicationDetails{ProcessingGeography: "US"}, Data: make([]apiv1alpha1.DataContext, n)},
	}
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
			Chart: "s3-flight",
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
			Chart: "xxx",
		},
	}
}

// CreateDb2ToS3CopyModule creates a copy module db2->s3
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
						Flow:   apiv1alpha1.Copy,
						Source: &apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.JdbcDb2, DataFormat: apiv1alpha1.Table},
						Sink:   &apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet},
					},
				},
				Actions: make([]pb.EnforcementAction, 2),
			},
			Chart: "yyy",
		},
	}
	db2Module.Spec.Capabilities.Actions[0] = pb.EnforcementAction{Name: "redact", Id: "redact-ID", Level: pb.EnforcementAction_COLUMN}
	db2Module.Spec.Capabilities.Actions[1] = pb.EnforcementAction{Name: "encrypt", Id: "encrypt-ID", Level: pb.EnforcementAction_COLUMN}
	return db2Module
}

var _ = Describe("M4DApplication Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Millisecond * 100

	Context("M4DApplication", func() {
		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
			// delete application
			appSignature := GetApplicationSignature()
			resource := &apiv1alpha1.M4DApplication{ObjectMeta: metav1.ObjectMeta{Name: appSignature.Name, Namespace: appSignature.Namespace}}
			_ = k8sClient.Delete(context.Background(), resource)

			Eventually(func() error {
				f := &apiv1alpha1.M4DApplication{}
				return k8sClient.Get(context.Background(), appSignature, f)
			}, timeout, interval).ShouldNot(Succeed())

			// delete storage
			storageSignature := GetStorageSignature()
			_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DBucket{ObjectMeta: metav1.ObjectMeta{Name: storageSignature.Name, Namespace: storageSignature.Namespace}})
			// delete modules
			_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DModule{ObjectMeta: metav1.ObjectMeta{Name: "implicit-copy-kafka-to-s3-stream", Namespace: "default"}})
			_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DModule{ObjectMeta: metav1.ObjectMeta{Name: "read-path", Namespace: "default"}})
			_ = k8sClient.Delete(context.Background(), &apiv1alpha1.M4DModule{ObjectMeta: metav1.ObjectMeta{Name: "implicit-copy-db2-to-s3", Namespace: "default"}})
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

			appSignature := GetApplicationSignature()
			resource := InitM4DApplication(1)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID: "{\"asset_id\": \"123\", \"catalog_id\": \"db2\"}",
				IFdetails: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet},
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

			appSignature := GetApplicationSignature()
			resource := InitM4DApplication(1)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID: "{\"asset_id\": \"deny-dataset\", \"catalog_id\": \"s3\"}",
				IFdetails: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet},
			}

			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())

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
			appSignature := GetApplicationSignature()
			resource := InitM4DApplication(1)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID: "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"db2\"}",
				IFdetails: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.S3, DataFormat: apiv1alpha1.Parquet},
			}
			Expect(k8sClient.Create(context.Background(), CreateReadPathModule())).Should(Succeed())

			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())

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
		// Db2 dataset
		// Read module with source=s3,parquet exists
		// Copy to s3 is required
		// Enforcement action for copy operation: Deny
		// Result: an error

		It("Test deny-on-copy", func() {
			module := CreateReadPathModule()
			Expect(k8sClient.Create(context.Background(), module)).Should(Succeed())

			appSignature := GetApplicationSignature()
			resource := InitM4DApplication(1)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID: "{\"asset_id\": \"deny-on-copy\", \"catalog_id\": \"db2\"}",
				IFdetails: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.ArrowFlight, DataFormat: apiv1alpha1.Arrow},
			}

			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())

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

			appSignature := GetApplicationSignature()
			resource := InitM4DApplication(2)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID: "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"db2\"}",
				IFdetails: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.ArrowFlight, DataFormat: apiv1alpha1.Arrow},
			}
			resource.Spec.Data[1] = apiv1alpha1.DataContext{
				DataSetID: "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"s3\"}",
				IFdetails: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.ArrowFlight, DataFormat: apiv1alpha1.Arrow},
			}

			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())

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
		// Applied copy module kafka->s3 and db2->s3 supporting redact and encrypt actions
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

			// Load kafka-s3 copy module
			kafkaModule := CreateKafkaToS3CopyModule()
			readPathModule := CreateReadPathModule()

			Expect(k8sClient.Create(context.Background(), kafkaModule)).Should(Succeed())
			Expect(k8sClient.Create(context.Background(), readPathModule)).Should(Succeed())

			// Load db2-s3 copy module
			db2Module := CreateDb2ToS3CopyModule()
			Expect(k8sClient.Create(context.Background(), db2Module)).Should(Succeed())

			appSignature := GetApplicationSignature()
			resource := InitM4DApplication(2)
			resource.Spec.Data[0] = apiv1alpha1.DataContext{
				DataSetID: "{\"asset_id\": \"default-dataset\", \"catalog_id\": \"db2\"}",
				IFdetails: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.ArrowFlight, DataFormat: apiv1alpha1.Arrow},
			}
			resource.Spec.Data[1] = apiv1alpha1.DataContext{
				DataSetID: "{\"asset_id\": \"allow-dataset\", \"catalog_id\": \"s3\"}",
				IFdetails: apiv1alpha1.InterfaceDetails{Protocol: apiv1alpha1.ArrowFlight, DataFormat: apiv1alpha1.Arrow},
			}

			// Create M4DApplication
			Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())
			By("Expecting a namespace to be allocated")
			Eventually(func() string {
				f := &apiv1alpha1.M4DApplication{}
				_ = k8sClient.Get(context.Background(), appSignature, f)
				return f.Status.BlueprintNamespace
			}, timeout, interval).ShouldNot(BeEmpty())

			By("Expecting blueprint to be generated")
			blueprint := &apiv1alpha1.Blueprint{}
			Eventually(func() error {
				Expect(k8sClient.Get(context.Background(), appSignature, resource)).Should(Succeed())
				key := appSignature
				key.Namespace = resource.Status.BlueprintNamespace
				return k8sClient.Get(context.Background(), key, blueprint)
			}, timeout, interval).Should(Succeed())

			// Check the generated blueprint
			// There should be a single read module with two datasets
			numReads := 0
			for _, step := range blueprint.Spec.Flow.Steps {
				if step.Template == readPathModule.Name {
					numReads++
					Expect(len(step.Arguments.Read)).To(Equal(2))
				}
			}
			Expect(numReads).To(Equal(1))
		})
	})

	Context("M4DApplication e2e", func() {
		BeforeEach(func() {
			// Add any teardown steps that needs to be executed after each test
			// delete application
			appSignature := GetApplicationSignature()
			resource := &apiv1alpha1.M4DApplication{ObjectMeta: metav1.ObjectMeta{Name: "notebook", Namespace: "default"}}
			_ = k8sClient.Delete(context.Background(), resource)

			Eventually(func() error {
				f := &apiv1alpha1.M4DApplication{}
				return k8sClient.Get(context.Background(), appSignature, f)
			}, timeout, interval).ShouldNot(Succeed())

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

			// A blueprint namespace should be set
			By("Expecting application to have a namespace in the status")
			Eventually(func() string {
				Expect(k8sClient.Get(context.Background(), applicationKey, application)).To(Succeed())
				return application.Status.BlueprintNamespace
			}, timeout, interval).ShouldNot(BeEmpty())

			// A blueprint namespace should be created
			namespace := &v1.Namespace{}
			ns := application.Status.BlueprintNamespace
			By("Expect namespace to be created")
			Eventually(func() error {
				return k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "", Name: ns}, namespace)
			}, timeout, interval).Should(Succeed())

			// The blueprint has to be created in the blueprint namespace
			blueprint := &apiv1alpha1.Blueprint{}
			blueprintObjectKey := client.ObjectKey{Namespace: ns, Name: application.Name}
			By("Expect blueprint to be created")
			Eventually(func() error {
				return k8sClient.Get(context.Background(), blueprintObjectKey, blueprint)
			}, timeout, interval).Should(Succeed())

			By("Expect blueprint to be ready at some point")
			Eventually(func() bool {
				Expect(k8sClient.Get(context.Background(), blueprintObjectKey, blueprint)).To(Succeed())
				return blueprint.Status.Ready
			}, timeout*10, interval).Should(BeTrue())

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
