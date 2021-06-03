// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package motion

import (
	"context"
	"io/ioutil"
	"time"

	corev1 "k8s.io/api/core/v1"

	motionv1 "github.com/mesh-for-data/mesh-for-data/manager/apis/motion/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var _ = Describe("StreamTransfer Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Millisecond * 100
	const streamtransferName = "streamtransfer-sample"
	const streamtransferNameSpace = "default"

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
		f := &motionv1.StreamTransfer{}
		key := client.ObjectKey{
			Namespace: streamtransferNameSpace,
			Name:      streamtransferName,
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
	Context("StreamTransfer", func() {
		It("Should simulate a StreamTranssfer successfully", func() {
			// Load stream transfer from YAML
			var err error
			streamTransferYAML, err := ioutil.ReadFile("../../testdata/streamtransfer.yaml")
			Expect(err).ToNot(HaveOccurred())
			streamTransfer := &motionv1.StreamTransfer{}
			err = yaml.Unmarshal(streamTransferYAML, streamTransfer)
			Expect(err).ToNot(HaveOccurred())

			key := client.ObjectKeyFromObject(streamTransfer)

			// Create StreamTransfer
			Expect(k8sClient.Create(context.Background(), streamTransfer)).Should(Succeed())

			// Ensure getting cleaned up after tests finish
			defer func() {
				f := &motionv1.StreamTransfer{ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name}}
				_ = k8sClient.Get(context.Background(), key, f)
				_ = k8sClient.Delete(context.Background(), f)
			}()

			By("Expecting Secret to be created")
			Eventually(func() error {
				secret := &corev1.Secret{}
				return k8sClient.Get(context.Background(), key, secret)
			}, timeout, interval).Should(Succeed())

			pvc := &corev1.PersistentVolumeClaim{}
			By("Expecting PVC to be created")
			Eventually(func() error {
				return k8sClient.Get(context.Background(), key, pvc)
			}, timeout, interval).Should(Succeed())

			deployment := &apps.Deployment{}
			By("Expecting Deployment to be created")
			Eventually(func() error {
				return k8sClient.Get(context.Background(), key, deployment)
			}, timeout, interval).Should(Succeed())

			// Depending on the external cluster the mover might run so fast that it's already running
			// when the StreamTransfer status is updated for the first time. Thus the check has to be for
			// all the valid statuses
			By("Expecting StreamTransfer status to be set to starting or starting/running")
			Eventually(func() motionv1.StreamStatus {
				Expect(k8sClient.Get(context.Background(), key, streamTransfer)).To(Succeed())
				return streamTransfer.Status.Status
			}, timeout, interval).Should(BeElementOf(motionv1.StreamStarting, motionv1.StreamStarting))

			if !noSimulatedProgress {
				// Simulate running phase of job
				deployment.Status.AvailableReplicas = 1
				deployment.Status.ReadyReplicas = 1
				deployment.Status.Replicas = 1
				Expect(k8sClient.Status().Update(context.Background(), deployment)).Should(Succeed())
			}

			// Depending on the external cluster the mover might run so fast that it's already running
			// when the StreamTransfer status is updated for the first time. Thus the check has to be for
			// all the valid statuses
			By("Expecting StreamTransfer status to be set to starting/running")
			Eventually(func() motionv1.StreamStatus {
				Expect(k8sClient.Get(context.Background(), key, streamTransfer)).To(Succeed())
				return streamTransfer.Status.Status
			}, timeout, interval).Should(BeElementOf(motionv1.StreamStarting, motionv1.StreamRunning))

			// Delete CRD and checking for finalizer
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &motionv1.StreamTransfer{}
				_ = k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			finalizerPod := &corev1.Pod{}
			By("Expect finalizer pod to be started")
			Eventually(func() error {

				finalizerKey := streamTransfer.FinalizerPodKey()
				return k8sClient.Get(context.Background(), finalizerKey, finalizerPod)
			}, timeout, interval).Should(Succeed())

			if !noSimulatedProgress {
				// Simulate a succeeded finalizer
				finalizerPod.Status.Phase = corev1.PodSucceeded
				Expect(k8sClient.Status().Update(context.Background(), finalizerPod)).Should(Succeed())
			}

			By("Expecting delete to finish")
			Eventually(func() error {
				f := &motionv1.StreamTransfer{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())
		})
	})
})
