// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package motion

import (
	"os"

	"emperror.dev/errors"
	motionv1 "fybrik.io/fybrik/manager/apis/motion/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// This function sets up all motion controllers including the webhooks given a controller manager.
// Webhooks can be activated/deactivated using the ENABLE_WEBHOOKS environment variable.
// This currently includes:
// - a manager for BatchTransfers
// - a manager for StreamTransfers
func SetupMotionControllers(mgr manager.Manager) error {
	if err := NewBatchTransferReconciler(mgr, "BatchTransferController").SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "unable to create BatchTransfer controller")
	}
	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err := (&motionv1.BatchTransfer{}).SetupWebhookWithManager(mgr); err != nil {
			return errors.Wrap(err, "unable to create BatchTransfer webhook")
		}
	}

	if err := NewStreamTransferReconciler(mgr, "StreamTransferController").SetupWithManager(mgr); err != nil {
		return errors.Wrap(err, "unable to create StreamTransfer controller")
	}

	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err := (&motionv1.StreamTransfer{}).SetupWebhookWithManager(mgr); err != nil {
			return errors.Wrap(err, "unable to create StreamTransfer webhook")
		}
	}

	return nil
}
