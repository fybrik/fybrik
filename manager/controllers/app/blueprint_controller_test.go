// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"time"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var _ = Describe("Blueprint Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Millisecond * 100

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Blueprint", func() {
		It("Should create successfully", func() {
			// randomized := randomStringWithCharset(10, charset)

			// Load
			var err error
			blueprintYAML, err := ioutil.ReadFile("../../testdata/blueprint.yaml")
			Expect(err).ToNot(HaveOccurred())
			blueprint := &app.Blueprint{}
			err = yaml.Unmarshal(blueprintYAML, blueprint)
			Expect(err).ToNot(HaveOccurred())

			key, err := client.ObjectKeyFromObject(blueprint)
			Expect(err).ToNot(HaveOccurred())

			// Create Blueprint
			Expect(k8sClient.Create(context.Background(), blueprint)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				f := &app.Blueprint{ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name}}
				_ = k8sClient.Get(context.Background(), key, f)
				_ = k8sClient.Delete(context.Background(), f)
			}()

			/*
				By("Expecting BatchTransfer to be created")
				Eventually(func() error {
					name := utils.Hash(blueprint.Spec.Flow.Steps[0].Name, 20)
					expectedObject := utils.CreateUnstructured("motion.m4d.ibm.com", "v1alpha1", "BatchTransfer",
						name, blueprint.Namespace)
					expectedObjectKey, err := client.ObjectKeyFromObject(expectedObject)
					Expect(err).ToNot(HaveOccurred())
					return k8sClient.Get(context.Background(), expectedObjectKey, expectedObject)
				}, timeout, interval).Should(Succeed())
			*/

			// Update
			By("Expecting to update successfully")
			Eventually(func() error {
				f := &app.Blueprint{}
				if err := k8sClient.Get(context.Background(), key, f); err != nil {
					return err
				}
				f.Spec.Flow.Steps[0].Arguments.Copy.Destination.Connection.S3.Bucket = "placeholder"
				return k8sClient.Update(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &app.Blueprint{}
				if err := k8sClient.Get(context.Background(), key, f); err != nil {
					return err
				}
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting delete to finish")
			Eventually(func() error {
				f := &app.Blueprint{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
