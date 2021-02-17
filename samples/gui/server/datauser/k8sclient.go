// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"context"

	dm "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// DMAClient contains the contextual info about the REST client for the M4DApplication CRD
type DMAClient struct {
	client    *rest.RESTClient
	namespace string
	plural    string
	codec     runtime.ParameterCodec
}

// Create makes a new M4DApplication CRD
func (f *DMAClient) Create(obj *dm.M4DApplication) (*dm.M4DApplication, error) {
	var result dm.M4DApplication
	err := f.client.Post().
		Namespace(f.namespace).Resource(f.plural).
		Body(obj).Do(context.Background()).Into(&result)
	return &result, err
}

// Update updates an existing M4DApplication CRD
func (f *DMAClient) Update(name string, obj *dm.M4DApplication) (*dm.M4DApplication, error) {
	var result dm.M4DApplication
	err := f.client.Put().
		Namespace(f.namespace).Resource(f.plural).Name(name).
		Body(obj).Do(context.Background()).Into(&result)
	return &result, err
}

// Delete terminates the existing M4DApplication CRD and all its associated components in the m4d
func (f *DMAClient) Delete(name string, options *meta_v1.DeleteOptions) error {
	return f.client.Delete().
		Namespace(f.namespace).Resource(f.plural).
		Name(name).Body(options).Do(context.Background()).
		Error()
}

// Get returns an existing M4DApplication CRD, including its status information
func (f *DMAClient) Get(name string) (*dm.M4DApplication, error) {
	var result dm.M4DApplication
	err := f.client.Get().
		Namespace(f.namespace).Resource(f.plural).
		Name(name).Do(context.Background()).Into(&result)
	return &result, err
}

// List gets the list of existing M4DApplication CRDs
func (f *DMAClient) List(opts meta_v1.ListOptions) (*dm.M4DApplicationList, error) {
	var result dm.M4DApplicationList
	err := f.client.Get().
		Namespace(f.namespace).Resource(f.plural).
		//		VersionedParams(&opts, f.codec).
		Do(context.Background()).Into(&result)
	return &result, err
}

// NewListWatch creates a new List watch for our CRD
func (f *DMAClient) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(f.client, f.plural, f.namespace, fields.Everything())
}
