// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

const (
	stepNameHashLength       = 10
	hashPostfixLength        = 5
	k8sMaxConformNameLength  = 63
	helmMaxConformNameLength = 53
)

// IsDenied returns true if the data access is denied
func IsDenied(actionName taxonomy.ActionName) bool {
	return actionName == "Deny" // TODO FIX THIS
}

// StructToMap converts a struct to a map using JSON marshal
func StructToMap(data interface{}) (map[string]interface{}, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	mapData := make(map[string]interface{})
	err = json.Unmarshal(dataBytes, &mapData)
	if err != nil {
		return nil, err
	}
	return mapData, nil
}

func HasString(value string, values []string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

// Hash generates a name based on the unique identifier
func Hash(value string, hashLength int) string {
	data := sha512.Sum512([]byte(value))
	hashedStr := hex.EncodeToString(data[:])
	if hashLength >= len(hashedStr) {
		return hashedStr
	}
	return hashedStr[:hashLength]
}

// Generating release name based on blueprint module
func GetReleaseName(applicationName, namespace, instanceName string) string {
	return GetReleaseNameByStepName(applicationName, namespace, instanceName)
}

// Generate release name from blueprint module name
func GetReleaseNameByStepName(applicationName, namespace, moduleInstanceName string) string {
	fullName := applicationName + "-" + namespace + "-" + moduleInstanceName
	return HelmConformName(fullName)
}

// Some k8s objects only allow for a length of 63 characters.
// This method shortens the name keeping a prefix and using the last 5 characters of the
// new name for the hash of the postfix.
func K8sConformName(name string) string {
	return ShortenedName(name, k8sMaxConformNameLength, hashPostfixLength)
}

// Helm has stricter restrictions than K8s and restricts release names to 53 characters
func HelmConformName(name string) string {
	return ShortenedName(name, helmMaxConformNameLength, hashPostfixLength)
}

// Create a name for a step in a blueprint.
// Since this is part of the name of a release, this should be done in a central location to make testing easier
func CreateStepName(moduleName, assetID string) string {
	return moduleName + "-" + Hash(assetID, stepNameHashLength)
}

// This function shortens a name to the maximum length given and uses rest of the string that is too long
// as hash that gets added to the valid name.
func ShortenedName(name string, maxLength, hashLength int) string {
	if len(name) > maxLength {
		// The new name is in the form prefix-suffix
		// The prefix is the prefix of the original name (so it's human readable)
		// The suffix is a deterministic hash of the suffix of the original name
		// Overall, the new name is deterministic given the original name
		cutOffIndex := maxLength - hashLength - 1
		prefix := name[:cutOffIndex]
		suffix := Hash(name[cutOffIndex:], hashLength)
		return prefix + "-" + suffix
	}
	return name
}

func ListeningAddress(port int) string {
	address := fmt.Sprintf(":%d", port)
	if runtime.GOOS == "darwin" {
		address = "localhost" + address
	}
	return address
}

// Intersection finds a common subset of two given sets of strings
func Intersection(set1, set2 []string) []string {
	res := []string{}
	for _, elem1 := range set1 {
		for _, elem2 := range set2 {
			if elem1 == elem2 {
				res = append(res, elem1)
				break
			}
		}
	}
	return res
}

const FybrikAppUUID = "app.fybrik.io/app-uuid"
const FybrikModuleNamespace = "app.fybrik.io/modules-namespace"

// GetFybrikModuleNamespaceFromLabels returns the moduleNamespace passed to the resource in its labels
func GetFybrikModuleNamespaceFromLabels(labels map[string]string) string {
	moduleNs, foundmoduleNs := labels[FybrikModuleNamespace]
	if !foundmoduleNs {
		return ""
	}
	return moduleNs
}

// GetFybrikModuleNamespace returns the moduleNamespace
// The order of resolving the moduleNamespace is as follows:
// If `app.fybrik.io/modules-namespace` label is set, it must be given precedence over
// other value defined by an environment variable or a default value if the latter
// is not set.
func GetFybrikModuleNamespace(labels map[string]string) string {
	moduleNs := GetFybrikModuleNamespaceFromLabels(labels)
	if moduleNs == "" {
		moduleNs = GetDefaultModulesNamespace()
	}
	return moduleNs
}

// GetFybrikApplicationUUID returns a globally unique ID for the FybrikApplication instance.
// It must be unique over time and across clusters, even after the instance has been deleted,
// because this ID will be used for logging purposes.
func GetFybrikApplicationUUID(fapp *api.FybrikApplication) string {
	// Use the clusterwise unique kubernetes id.
	// No need to add cluster because FybrikApplication instances can only be created on the coordinator cluster.
	return string(fapp.GetObjectMeta().GetUID())
}

// GetFybrikApplicationUUIDfromAnnotations returns the UUID passed to the resource in its annotations
func GetFybrikApplicationUUIDfromAnnotations(annotations map[string]string) string {
	uuid, founduuid := annotations[FybrikAppUUID]
	if !founduuid {
		return "UUID missing"
	}
	return uuid
}

// UpdateFinalizers adds or removes finalizers for a resource
func UpdateFinalizers(ctx context.Context, cl client.Client, obj client.Object) error {
	err := cl.Update(ctx, obj)
	if !errors.IsConflict(err) {
		return err
	}
	finalizers := obj.GetFinalizers()
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of the object before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		if err := cl.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
			return err
		}
		obj.SetFinalizers(finalizers)
		return cl.Update(ctx, obj)
	})
}

// UpdateStatus updates the resource status
func UpdateStatus(ctx context.Context, cl client.Client, obj client.Object, previousStatus interface{}) error {
	err := cl.Status().Update(ctx, obj)
	if !errors.IsConflict(err) {
		return err
	}
	values, err := StructToMap(obj)
	if err != nil {
		return err
	}
	statusKey := "status"
	currentStatus := values[statusKey]
	if previousStatus != nil && equality.Semantic.DeepEqual(previousStatus, currentStatus) {
		return nil
	}

	res := &unstructured.Unstructured{}
	res.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
	res.SetName(obj.GetName())
	res.SetNamespace(obj.GetNamespace())

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of the object before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		if err := cl.Get(ctx, client.ObjectKeyFromObject(res), res); err != nil {
			return err
		}
		res.Object[statusKey] = currentStatus
		return cl.Status().Update(ctx, res)
	})
}
