// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"github.com/onsi/gomega"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	app "fybrik.io/fybrik/manager/apis/app/v1"
)

// Creates a scheme that can be used in unit tests
// The scheme will have the core and batch apis from K8s registered as well as
// the app api from Fybrik.
// This function can be tested with a gomega environment if passed or otherwise (if nil is passed) it will ignore tests.
func NewScheme(g *gomega.WithT) *runtime.Scheme {
	s := runtime.NewScheme()
	err := corev1.AddToScheme(s)
	if g != nil {
		g.Expect(err).NotTo(gomega.HaveOccurred())
	}
	err = kbatch.AddToScheme(s)
	if g != nil {
		g.Expect(err).NotTo(gomega.HaveOccurred())
	}
	err = app.AddToScheme(s)
	if g != nil {
		g.Expect(err).NotTo(gomega.HaveOccurred())
	}
	return s
}
