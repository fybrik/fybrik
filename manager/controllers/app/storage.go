// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	statusErr "k8s.io/apimachinery/pkg/api/errors"

	"github.com/ibm/the-mesh-for-data/manager/controllers/utils"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dbuckets,verbs=get;list;watch;update;
// +kubebuilder:rbac:groups=app.m4d.ibm.com,resources=m4dbuckets/status,verbs=get;update;patch

// FindAvailableBucket finds an available storage asset
func (r *M4DApplicationReconciler) FindAvailableBucket(owner types.NamespacedName, id string, prefixBase string, canShare bool) *app.M4DBucket {
	ctx := context.Background()

	var buckets app.M4DBucketList
	r.Log.V(0).Info("Searching for an available bucket")
	if err := r.List(ctx, &buckets); err != nil {
		r.Log.V(0).Info(err.Error())
		return nil
	}
	for _, bucket := range buckets.Items {
		utils.PrintStructure(bucket, r.Log, "Bucket ")
		if IsBucketAvailable(&bucket, owner, id, canShare) {
			AddOwner(&bucket, owner)
			GetOrCreatePrefix(&bucket, id, prefixBase)
			if err := r.Client.Status().Update(ctx, &bucket); err != nil {
				// the object has been updated by someone else - continue to the next one
				if statusErr.IsConflict(err) {
					r.Log.V(0).Info("Conflict during an update of M4DBucket " + bucket.Name)
				}
				r.Log.V(0).Info("Could not update M4DBucket " + bucket.Name)
				continue
			}
			return &bucket
		}
	}
	return nil
}

// FreeStorageAssets removes the app identifier from the list of owners
func (r *M4DApplicationReconciler) FreeStorageAssets(owner types.NamespacedName) error {
	ctx := context.Background()
	var buckets app.M4DBucketList
	if err := r.List(ctx, &buckets); err != nil {
		return err
	}
	for _, bucket := range buckets.Items {
		if !HasOwner(&bucket, owner) {
			continue
		}
		RemoveOwner(&bucket, owner)
		if !HasOwners(&bucket) {
			bucket.Status.AssetPrefixPerDataset = map[string]string{}
		}
		if err := r.Client.Status().Update(ctx, &bucket); err != nil {
			r.Log.V(0).Info("Error during an update of M4DBucket: " + err.Error())
			// the object has been updated by someone else - continue to the next one
			if statusErr.IsConflict(err) {
				r.Log.V(0).Info("Conflict during an update of M4DBucket " + bucket.Name)
				return err
			}
		}
	}
	return nil
}

// Utility functions

// CreateOwnerID creates a string based on namespace and name values
func CreateOwnerID(owner types.NamespacedName) string {
	return owner.Namespace + "/" + owner.Name
}

// HasOwner checks whether the given owner owns the resource
func HasOwner(resource *app.M4DBucket, owner types.NamespacedName) bool {
	ownerID := CreateOwnerID(owner)
	for _, val := range resource.Status.Owners {
		if val == ownerID {
			return true
		}
	}
	return false
}

// AddOwner adds an owner to the resource
func AddOwner(resource *app.M4DBucket, owner types.NamespacedName) {
	if HasOwner(resource, owner) {
		return
	}
	ownerID := CreateOwnerID(owner)
	resource.Status.Owners = append(resource.Status.Owners, ownerID)
}

// RemoveOwner removes the owner from the resource
func RemoveOwner(resource *app.M4DBucket, owner types.NamespacedName) {
	ownerID := CreateOwnerID(owner)
	newOwners := make([]string, 0)
	for _, val := range resource.Status.Owners {
		if val != ownerID {
			newOwners = append(newOwners, val)
		}
	}
	resource.Status.Owners = newOwners
}

// HasOwners checks whether there are any owners for the given resource
func HasOwners(resource *app.M4DBucket) bool {
	return len(resource.Status.Owners) != 0
}

// GetPrefix returns a prefix earlier generated for the given data set if exists, and empty string otherwise
func GetPrefix(resource *app.M4DBucket, id string) string {
	if resource.Status.AssetPrefixPerDataset == nil {
		return ""
	}
	elem, ok := resource.Status.AssetPrefixPerDataset[id]
	if !ok {
		return ""
	}
	return elem
}

// AddPrefix adds a prefix generated for the given data set
func AddPrefix(resource *app.M4DBucket, id string, prefix string) {
	if resource.Status.AssetPrefixPerDataset == nil {
		resource.Status.AssetPrefixPerDataset = make(map[string]string)
	}
	resource.Status.AssetPrefixPerDataset[id] = prefix
}

// GetOrCreatePrefix creates a new prefix for the given data set if none exists, based on the given name
func GetOrCreatePrefix(resource *app.M4DBucket, id string, name string) string {
	prefix := GetPrefix(resource, id)
	if prefix != "" {
		return prefix
	}
	prefix = name + utils.Hash(name, 10)
	AddPrefix(resource, id, prefix)
	return prefix
}

// RemovePrefix removes a prefix generated for the given data set
func RemovePrefix(resource *app.M4DBucket, id string) {
	delete(resource.Status.AssetPrefixPerDataset, id)
}

// IsBucketAvailable checks whether the bucket is available for use
func IsBucketAvailable(resource *app.M4DBucket, owner types.NamespacedName, id string, canShare bool) bool {
	// is owned by any resource?
	if !HasOwners(resource) {
		return true
	}
	// is owned by this resource only?
	if HasOwner(resource, owner) && len(resource.Status.Owners) == 1 {
		return true
	}
	// sharing available for the given data set
	if canShare && GetPrefix(resource, id) != "" {
		return true
	}
	return false
}
