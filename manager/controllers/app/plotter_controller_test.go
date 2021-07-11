// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package app

import (
	"context"
	"time"

	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Deploy the plotter and check that the result is what was expected.
func deployAndCheck(namespace string, shouldSucceed bool) {
	const timeout = time.Second * 30
	const interval = time.Millisecond * 100

	// Create the plotter yaml from the hard coded testdata file
	plotter := &app.Plotter{}
	Expect(readObjectFromFile("../../testdata/plotter.yaml", plotter)).ToNot(HaveOccurred())

	// Set the namespace we received
	plotter.SetNamespace(namespace)

	// Create the Plotter
	// Create the Plotter.  Need to wait in case it's still in the process of being deleted from a previous test.
	By("Expecting the plotter to be created.")
	Eventually(func() error {
		return k8sClient.Create(context.Background(), plotter)
	}, timeout, interval).Should(Succeed())

	// Don't forget to clean up after tests finish
	plotterKey := client.ObjectKeyFromObject(plotter)
	defer func() {
		plotter := &app.Plotter{ObjectMeta: metav1.ObjectMeta{Namespace: plotterKey.Namespace, Name: plotterKey.Name}}
		_ = k8sClient.Get(context.Background(), plotterKey, plotter)
		_ = k8sClient.Delete(context.Background(), plotter)
	}()

	By("Expecting plotter to be created")
	Eventually(func() error {
		return k8sClient.Get(context.Background(), plotterKey, plotter)
	}, timeout, interval).Should(Succeed())

	// Depending on the test being run, sometimes it should succeed and sometimes it shouldn't.
	if !shouldSucceed {
		By("Expecting Plotter to never be ready because reconcile should not be called")
		Eventually(func() bool {
			Expect(k8sClient.Get(context.Background(), plotterKey, plotter)).To(Succeed())
			return plotter.Status.ObservedState.Ready && plotter.Status.ObservedGeneration > 0
		}, timeout, interval).Should(BeFalse(), "Plotter should not be ready because reconciler should not have been invoked!")
	} else {
		By("Expecting Plotter to eventually be ready")
		Eventually(func() bool {
			Expect(k8sClient.Get(context.Background(), plotterKey, plotter)).To(Succeed())
			return plotter.Status.ObservedState.Ready
		}, timeout, interval).Should(BeTrue(), "Plotter is not ready after timeout!")
	}
}

var _ = Describe("Plotter Controller Illegal Create Event", func() {
	Context("Plotter", func() {
		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
		})

		// Plotters not deployed to m4d-system should not be successfully reconciled due to the filter preventing
		// reconcile from being called.
		It("Test Plotter Deploy to Bad Namespace", func() {
			deployAndCheck("default", false)
		})
	})
})

var _ = Describe("Plotter Controller Legal Create Event", func() {
	Context("Plotter", func() {
		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
		})

		// Plotter are successfully reconciled when deployed to m4d-system only
		It("Test Plotter Deploy to Correct Namespace", func() {
			deployAndCheck(utils.GetSystemNamespace(), true)
		})
	})
})

// Create and then Delete the plotter and check that the result is what was expected.
// Deploy the plotter and check that the result is what was expected.
func deployDeleteAndCheck(namespace string, shouldSucceed bool) {
	const timeout = time.Second * 30
	const interval = time.Millisecond * 100

	// Create the plotter yaml from the hard coded testdata file
	plotter := &app.Plotter{}
	Expect(readObjectFromFile("../../testdata/plotter.yaml", plotter)).ToNot(HaveOccurred())

	// Set the namespace to the system namespace so that the plotter will be successfully created
	plotter.SetNamespace(utils.GetSystemNamespace())

	// Create the Plotter.  Need to wait in case it's still in the process of being deleted from a previous test.
	By("Expecting the plotter to be created.")
	Eventually(func() error {
		return k8sClient.Create(context.Background(), plotter)
	}, timeout, interval).Should(Succeed())

	// Don't forget to clean up after tests finish
	plotterKey := client.ObjectKeyFromObject(plotter)
	defer func() {
		plotter := &app.Plotter{ObjectMeta: metav1.ObjectMeta{Namespace: plotterKey.Namespace, Name: plotterKey.Name}}
		_ = k8sClient.Get(context.Background(), plotterKey, plotter)
		_ = k8sClient.Delete(context.Background(), plotter)
	}()

	By("Expecting to find the plotter we created")
	Eventually(func() error {
		return k8sClient.Get(context.Background(), plotterKey, plotter)
	}, timeout, interval).Should(Succeed())

	By("Expecting Plotter to eventually be ready")
	Eventually(func() bool {
		Expect(k8sClient.Get(context.Background(), plotterKey, plotter)).To(Succeed())
		return plotter.Status.ObservedState.Ready
	}, timeout, interval).Should(BeTrue(), "Plotter is not ready after timeout!")

	// Now create a request to delete the same plotter
	// Set the namespace to the namespace being used for this test
	delPlotter := &app.Plotter{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: plotterKey.Name}}

	// Delete the Plotter
	if shouldSucceed {
		Expect(k8sClient.Delete(context.Background(), delPlotter)).Should(Succeed())
	} else {
		Expect(k8sClient.Delete(context.Background(), delPlotter)).ShouldNot(Succeed())

		// Cleanup - from correct namespace
		Expect(k8sClient.Delete(context.Background(), plotter)).Should(Succeed())
	}
}

var _ = Describe("Plotter Controller Legal Delete Event", func() {
	Context("Plotter", func() {
		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
		})

		// Plotter is successfully deleted when deletion is done from m4d-system only
		It("Test Plotter Deletion from Correct Namespace", func() {
			deployDeleteAndCheck(utils.GetSystemNamespace(), true)
		})
	})
})

var _ = Describe("Plotter Controller Illegal Delete Event", func() {
	Context("Plotter", func() {
		BeforeEach(func() {
			// Add any setup steps that needs to be executed before each test
		})

		AfterEach(func() {
			// Add any teardown steps that needs to be executed after each test
		})

		// Plotter is NOT deleted when deletion is done from a namespace other than m4d-system
		It("Test Plotter Deletion from Bad Namespace", func() {
			deployDeleteAndCheck("default", false)
		})
	})
})
