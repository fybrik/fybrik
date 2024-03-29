// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"strings"

	api "fybrik.io/fybrik/manager/apis/app/v1beta1"
)

const (
	FybrikPrefix              = "app.fybrik.io"
	ApplicationClusterLabel   = FybrikPrefix + "/app-cluster"
	ApplicationNamespaceLabel = FybrikPrefix + "/app-namespace"
	ApplicationNameLabel      = FybrikPrefix + "/app-name"
	BlueprintNamespaceLabel   = FybrikPrefix + "/blueprint-namespace"
	BlueprintNameLabel        = FybrikPrefix + "/blueprint-name"
	FybrikAppUUID             = FybrikPrefix + "/app-uuid"
	KubernetesInstance        = "app.kubernetes.io/instance"
	KubernetesNamespaceName   = "kubernetes.io/metadata.name"
	KubernetesAppName         = "app.kubernetes.io/name"
	KubernetesAppNameOld      = "k8s-app"
	OpenShiftDNS              = "dns.operator.openshift.io/daemonset-dns"
)

func GetApplicationClusterFromLabels(labels map[string]string) string {
	return labels[ApplicationClusterLabel]
}

func GetApplicationNamespaceFromLabels(labels map[string]string) string {
	return labels[ApplicationNamespaceLabel]
}

func GetApplicationNameFromLabels(labels map[string]string) string {
	return labels[ApplicationNameLabel]
}

func GetBlueprintNamespaceFromLabels(labels map[string]string) string {
	return labels[BlueprintNamespaceLabel]
}

func GetBlueprintNameFromLabels(labels map[string]string) string {
	return labels[BlueprintNameLabel]
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

func CopyFybrikLabels(m map[string]string) map[string]string {
	labels := map[string]string{}
	for k, v := range m {
		if strings.HasPrefix(k, FybrikPrefix) {
			labels[k] = v
		}
	}
	return labels
}
