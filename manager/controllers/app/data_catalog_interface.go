// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"errors"

	"encoding/json"

	app "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RegisterAsset registers a new asset in the specified catalog
// Input arguments:
// - catalogID: the destination catalog identifier
// - info: connection and credential details
// Returns:
// - an error if happened
// - the new asset identifier
func (r *FybrikApplicationReconciler) RegisterAsset(catalogID string, info *app.DatasetDetails, input *app.FybrikApplication) (string, error) {
	return "", errors.New("Unsupported feature")
}

var translationMap = map[string]string{
	"accessKeyID":        "access_key",
	"accessKey":          "access_key",
	"secretAccessKey":    "secret_key",
	"SecretKey":          "secret_key",
	"apiKey":             "api_key",
	"resourceInstanceId": "resource_instance_id",
}

// SecretToCredentialMap fetches a secret and converts into a map matching credentials proto
func SecretToCredentialMap(cl client.Client, secretRef types.NamespacedName) (map[string]interface{}, error) {
	// fetch a secret
	secret := &corev1.Secret{}
	if err := cl.Get(context.Background(), secretRef, secret); err != nil {
		return nil, err
	}
	credsMap := make(map[string]interface{})
	for key, val := range secret.Data {
		if translated, found := translationMap[key]; found {
			credsMap[translated] = string(val)
		} else {
			credsMap[key] = string(val)
		}
	}
	return credsMap, nil
}

// SecretToCredentials fetches a secret and constructs Credentials structure
func SecretToCredentials(cl client.Client, secretRef types.NamespacedName) (string, error) {
	credsMap, err := SecretToCredentialMap(cl, secretRef)
	if err != nil {
		return "", err
	}
	jsonStr, err := json.Marshal(credsMap)
	if err != nil {
		return "", err
	}
	return string(jsonStr), nil
}
