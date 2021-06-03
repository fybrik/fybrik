// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"io/ioutil"
	"strings"

	"emperror.dev/errors"
	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// K8sClient contains the contextual info about the REST client for k8s
type K8sClient struct {
	client    kclient.Client
	namespace string
}

// GetCurrentNamespace returns the namespace in which the REST api service is deployed - which should be a user namespace
func GetCurrentNamespace() string {
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
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

	client, err := kclient.New(kconfig.GetConfigOrDie(), kclient.Options{Scheme: newScheme})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}
	return &K8sClient{client: client, namespace: GetCurrentNamespace()}, nil
}
