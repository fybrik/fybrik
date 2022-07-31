// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"os"
	"strings"

	"emperror.dev/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	app "fybrik.io/fybrik/manager/apis/app/v1"
)

const (
	currentNamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

// K8sClient contains the contextual info about the REST client for k8s
type K8sClient struct {
	client    kclient.Client
	namespace string
}

// GetCurrentNamespace returns the namespace in which the REST api service is deployed - which should be a user namespace
func GetCurrentNamespace() string {
	if data, err := os.ReadFile(currentNamespacePath); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return "default"
}

// K8sInit initializes a client to communicate with kubernetes
func K8sInit() (*K8sClient, error) {
	newScheme := runtime.NewScheme()
	_ = app.AddToScheme(newScheme)
	_ = v1.AddToScheme(newScheme)
	client, err := kclient.New(kconfig.GetConfigOrDie(), kclient.Options{Scheme: newScheme})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}
	return &K8sClient{client: client, namespace: GetCurrentNamespace()}, nil
}
