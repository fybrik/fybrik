// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package motion

import (
	"os"

	motionv1 "github.com/ibm/the-mesh-for-data/manager/apis/motion/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// This function sets up all motion controllers including the webhooks given a controller manager.
// Webhooks can be activated/deactivated using the ENABLE_WEBHOOKS environment variable.
// This currently includes:
// - a manager for BatchTransfers
// - a manager for StreamTransfers
func SetupMotionControllers(mgr manager.Manager) {
	setupLog := ctrl.Log.WithName("setup")

	if err := NewBatchTransferReconciler(mgr, "BatchTransferController").SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BatchTransfer")
		os.Exit(1)
	}

	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err := (&motionv1.BatchTransfer{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Captain")
			os.Exit(1)
		}
	}

	if err := NewStreamTransferReconciler(mgr, "StreamTransferController").SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "StreamTransfer")
		os.Exit(1)
	}

	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err := (&motionv1.StreamTransfer{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "StreamTransfer")
			os.Exit(1)
		}
	}
}
