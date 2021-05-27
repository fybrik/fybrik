// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package motion

import (
	"context"
	"testing"

	"github.com/mesh-for-data/mesh-for-data/manager/controllers/utils"

	"github.com/onsi/gomega"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	motionv1 "github.com/mesh-for-data/mesh-for-data/manager/apis/motion/v1alpha1"

	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TestBatchTransferController runs BatchTransferReconciler.Reconcile() against a
// fake client that tracks a BatchTransfer object.
// This test does not require a Kubernetes environment to run.
// This mechanism of testing can be used to test corner cases of the reconcile function.
func TestBatchTransferController(t *testing.T) {
	t.Parallel()
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(zap.New(zap.UseDevMode(true)))
	g := gomega.NewGomegaWithT(t)

	var (
		name      = "sample-transfer"
		namespace = "m4d-system"
	)
	batchTransfer := &motionv1.BatchTransfer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: motionv1.BatchTransferSpec{
			Source: motionv1.DataStore{
				Database: &motionv1.Database{
					Db2URL:   "jdbc:db2://host:1234/DB",
					Table:    "MY.TABLE",
					User:     "user",
					Password: "password",
				},
			},
			Destination: motionv1.DataStore{
				S3: &motionv1.S3{
					Endpoint:   "my.endpoint",
					Region:     "eu-gb",
					Bucket:     "myBucket",
					AccessKey:  "ab",
					SecretKey:  "cd",
					ObjectKey:  "obj.parq",
					DataFormat: "parquet",
				},
			},
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{
		batchTransfer,
	}

	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	// Create a BatchTransferReconciler object with the scheme and fake client.
	r := &BatchTransferReconciler{
		Reconciler{
			Client: cl,
			Log:    ctrl.Log.WithName("test-controller"),
			Scheme: s,
		},
	}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// Check if Job has been created and has the correct size.
	job := &kbatch.Job{}
	err = cl.Get(context.TODO(), req.NamespacedName, job)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}

	// Check if the secret has been created
	secret := &corev1.Secret{}
	err = cl.Get(context.TODO(), req.NamespacedName, secret)
	if err != nil {
		t.Fatalf("get secret: (%v)", err)
	}
}
