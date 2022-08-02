// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	app "fybrik.io/fybrik/manager/apis/app/v1beta1"
)

// CreateApplication makes a new FybrikApplication CRD
func (f *K8sClient) CreateApplication(obj *app.FybrikApplication) (*app.FybrikApplication, error) {
	err := f.client.Create(context.Background(), obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// UpdateApplication updates an existing FybrikApplication CRD
func (f *K8sClient) UpdateApplication(name string, obj *app.FybrikApplication) (*app.FybrikApplication, error) {
	var result app.FybrikApplication
	key := kclient.ObjectKeyFromObject(obj)
	err := f.client.Get(context.Background(), key, &result)
	if err != nil {
		return nil, err
	}
	result.Spec = obj.Spec
	err = f.client.Update(context.Background(), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteApplication terminates the existing FybrikApplication CRD and all its associated components in the fybrik
func (f *K8sClient) DeleteApplication(name string, options *meta_v1.DeleteOptions) error {
	var result app.FybrikApplication
	key := types.NamespacedName{Name: name, Namespace: k8sClient.namespace}
	err := f.client.Get(context.Background(), key, &result)
	if err != nil {
		return err
	}
	err = f.client.Delete(context.Background(), &result)
	return err
}

// GetApplication returns an existing FybrikApplication CRD, including its status information
func (f *K8sClient) GetApplication(name string) (*app.FybrikApplication, error) {
	var result app.FybrikApplication
	key := types.NamespacedName{Name: name, Namespace: k8sClient.namespace}
	err := f.client.Get(context.Background(), key, &result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// ListApplications gets the list of existing FybrikApplication CRDs
func (f *K8sClient) ListApplications() (*app.FybrikApplicationList, error) {
	var result app.FybrikApplicationList
	err := f.client.List(context.Background(), &result, kclient.InNamespace(k8sClient.namespace))
	return &result, err
}

// CreateOrUpdateSecret makes a new Secret or updates an existing one using received credentials
func (f *K8sClient) CreateOrUpdateSecret(obj *corev1.Secret) (*corev1.Secret, error) {
	var existing corev1.Secret
	key := kclient.ObjectKeyFromObject(obj)
	err := f.client.Get(context.Background(), key, &existing)
	if err == nil {
		// update (add new properties on top of the existing)
		if existing.Data == nil {
			existing.Data = map[string][]byte{}
		}
		if existing.StringData == nil {
			existing.StringData = map[string]string{}
		}
		for key, val := range obj.Data {
			existing.Data[key] = val
		}
		for key, val := range obj.StringData {
			existing.StringData[key] = val
		}
		err = f.client.Update(context.Background(), &existing)
		if err != nil {
			return nil, err
		}
		return &existing, nil
	}
	// create
	err = f.client.Create(context.Background(), obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// DeleteSecret deletes the existing secret
func (f *K8sClient) DeleteSecret(name string, options *meta_v1.DeleteOptions) error {
	var result corev1.Secret
	key := types.NamespacedName{Name: name, Namespace: k8sClient.namespace}
	err := f.client.Get(context.Background(), key, &result)
	if err != nil {
		return err
	}
	err = f.client.Delete(context.Background(), &result)
	return err
}

// GetSecret returns an existing secret
func (f *K8sClient) GetSecret(name string) (*corev1.Secret, error) {
	var result corev1.Secret
	key := types.NamespacedName{Name: name, Namespace: k8sClient.namespace}
	err := f.client.Get(context.Background(), key, &result)
	if err != nil {
		return nil, err
	}
	return &result, err
}
