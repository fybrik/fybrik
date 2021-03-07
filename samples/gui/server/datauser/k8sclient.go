// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"context"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateApplication makes a new M4DApplication CRD
func (f *K8sClient) CreateApplication(obj *app.M4DApplication) (*app.M4DApplication, error) {
	err := f.client.Create(context.Background(), obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// UpdateApplication updates an existing M4DApplication CRD
func (f *K8sClient) UpdateApplication(name string, obj *app.M4DApplication) (*app.M4DApplication, error) {
	var result app.M4DApplication
	key, _ := kclient.ObjectKeyFromObject(obj)
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

// DeleteApplication terminates the existing M4DApplication CRD and all its associated components in the m4d
func (f *K8sClient) DeleteApplication(name string, options *meta_v1.DeleteOptions) error {
	var result app.M4DApplication
	key := types.NamespacedName{Name: name, Namespace: k8sClient.namespace}
	err := f.client.Get(context.Background(), key, &result)
	if err != nil {
		return err
	}
	err = f.client.Delete(context.Background(), &result)
	return err
}

// GetApplication returns an existing M4DApplication CRD, including its status information
func (f *K8sClient) GetApplication(name string) (*app.M4DApplication, error) {
	var result app.M4DApplication
	key := types.NamespacedName{Name: name, Namespace: k8sClient.namespace}
	err := f.client.Get(context.Background(), key, &result)
	if err != nil {
		return nil, err
	}
	return &result, err
}

// ListApplications gets the list of existing M4DApplication CRDs
func (f *K8sClient) ListApplications(opts meta_v1.ListOptions) (*app.M4DApplicationList, error) {
	var result app.M4DApplicationList
	err := f.client.List(context.Background(), &result, kclient.InNamespace(k8sClient.namespace))
	return &result, err
}

// CreateSecret makes a new Secret
func (f *K8sClient) CreateSecret(obj *corev1.Secret) (*corev1.Secret, error) {
	var result corev1.Secret
	key, _ := kclient.ObjectKeyFromObject(obj)
	err := f.client.Get(context.Background(), key, &result)
	if err == nil {
		return f.UpdateSecret(obj.Name, obj)
	}
	err = f.client.Create(context.Background(), obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// UpdateSecret updates an existing secret
func (f *K8sClient) UpdateSecret(name string, obj *corev1.Secret) (*corev1.Secret, error) {
	var result corev1.Secret
	key, _ := kclient.ObjectKeyFromObject(obj)
	err := f.client.Get(context.Background(), key, &result)
	if err != nil {
		return nil, err
	}
	result.Data = obj.Data
	result.StringData = obj.StringData
	err = f.client.Update(context.Background(), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
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
