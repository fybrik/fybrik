// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"emperror.dev/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// K8sInit initializes a client to communicate with kubernetes
func K8sInit() (kclient.Client, error) {
	newScheme := runtime.NewScheme()
	_ = v1.AddToScheme(newScheme)
	client, err := kclient.New(kconfig.GetConfigOrDie(), kclient.Options{Scheme: newScheme})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}
	return client, nil
}
