// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package motion

import (
	"context"
	"io/ioutil"
	"time"

	kbatch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"

	motionv1 "github.com/mesh-for-data/mesh-for-data/manager/apis/motion/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var _ = Describe("BatchTransfer Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Millisecond * 100
	const batchtransferName = "batchtransfer-sample"
	const batchtransferNameSpace = "m4d-blueprints"

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
		f := &motionv1.BatchTransfer{}
		key := client.ObjectKey{
			Namespace: batchtransferNameSpace,
			Name:      batchtransferName,
		}
		if err := k8sClient.Get(context.Background(), key, f); err == nil {
			f.RemoveFinalizer()
			_ = k8sClient.Update(context.Background(), f)
			time.Sleep(interval)
			_ = k8sClient.Delete(context.Background(), f)
		}
	})

	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("BatchTransfer", func() {
		It("Should simulate a BatchTransfer successfully", func() {
			// Load batchtransfer from YAML
			var err error
			batchTransferYAML, err := ioutil.ReadFile("../../testdata/batchtransfer.yaml")
			Expect(err).ToNot(HaveOccurred())
			batchTransfer := &motionv1.BatchTransfer{}
			err = yaml.Unmarshal(batchTransferYAML, batchTransfer)
			Expect(err).ToNot(HaveOccurred())

			key := client.ObjectKeyFromObject(batchTransfer)

			// Create BatchTransfer
			Expect(k8sClient.Create(context.Background(), batchTransfer)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				f := &motionv1.BatchTransfer{ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name}}
				_ = k8sClient.Get(context.Background(), key, f)
				_ = k8sClient.Delete(context.Background(), f)
			}()

			job := &kbatch.Job{}
			By("Expecting Job to be created")
			Eventually(func() error {
				return k8sClient.Get(context.Background(), key, job)
			}, timeout, interval).Should(Succeed())

			By("Expecting Secret to be created")
			Eventually(func() error {
				secret := &v1.Secret{}
				return k8sClient.Get(context.Background(), key, secret)
			}, timeout, interval).Should(Succeed())

			// Depending on the external cluster the mover might run so fast that it's already running or finished
			// when the BatchTransfer status is updated for the first time. Thus the check has to be for
			// all the valid statuses
			By("Expecting BatchTransfer status to be set to starting")
			Eventually(func() motionv1.BatchStatus {
				Expect(k8sClient.Get(context.Background(), key, batchTransfer)).To(Succeed())
				return batchTransfer.Status.Status
			}, timeout, interval).Should(BeElementOf(motionv1.Starting, motionv1.Running, motionv1.Succeeded))

			if !noSimulatedProgress {
				// Simulate running phase of job
				job.Status.Active = 1
				Expect(k8sClient.Status().Update(context.Background(), job)).Should(Succeed())
			}

			// Depending on the external cluster the mover might run so fast that it's already running or finished
			// when the BatchTransfer status is updated for the first time. Thus the check has to be for
			// all the valid statuses
			By("Expecting BatchTransfer status to be set to running")
			Eventually(func() motionv1.BatchStatus {
				Expect(k8sClient.Get(context.Background(), key, batchTransfer)).To(Succeed())
				return batchTransfer.Status.Status
			}, timeout, interval).Should(BeElementOf(motionv1.Starting, motionv1.Running, motionv1.Succeeded))

			if !noSimulatedProgress {
				// Simulate succeeded phase of job
				job.Status.Active = 0
				job.Status.Succeeded = 1
				job.Status.Conditions = []kbatch.JobCondition{
					{
						Type:   kbatch.JobComplete,
						Status: v1.ConditionTrue,
					}}
				Expect(k8sClient.Status().Update(context.Background(), job)).Should(Succeed())
			}

			// Depending on the external cluster the mover might run so fast that it's already running or finished
			// when the BatchTransfer status is updated for the first time. Thus the check has to be for
			// all the valid statuses
			By("Expecting BatchTransfer status to be set to succeeded")
			Eventually(func() motionv1.BatchStatus {
				Expect(k8sClient.Get(context.Background(), key, batchTransfer)).To(Succeed())
				return batchTransfer.Status.Status
			}, timeout, interval).Should(BeElementOf(motionv1.Starting, motionv1.Running, motionv1.Succeeded))

			// Delete CRD and checking for finalizer
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &motionv1.BatchTransfer{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			finalizerPod := &v1.Pod{}
			By("Expect finalizer pod to be started")
			Eventually(func() error {

				finalizerKey := client.ObjectKey{
					Namespace: job.Namespace,
					Name:      job.Name + "-finalizer",
				}
				return k8sClient.Get(context.Background(), finalizerKey, finalizerPod)
			}, timeout, interval).Should(Succeed())

			if !noSimulatedProgress {
				// Simulate a succeeded finalizer
				finalizerPod.Status.Phase = v1.PodSucceeded
				Expect(k8sClient.Status().Update(context.Background(), finalizerPod)).Should(Succeed())
			}

			By("Expecting delete to finish")
			Eventually(func() error {
				f := &motionv1.BatchTransfer{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
