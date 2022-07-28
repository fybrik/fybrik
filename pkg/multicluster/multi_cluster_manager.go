// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package multicluster

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"fybrik.io/fybrik/manager/apis/app/v12"
)

type ClusterLister interface {
	GetClusters() ([]Cluster, error)
	IsMultiClusterSetup() bool
}

type ClusterManager interface {
	ClusterLister
	GetBlueprint(cluster string, namespace string, name string) (*v12.Blueprint, error)
	CreateBlueprint(cluster string, blueprint *v12.Blueprint) error
	UpdateBlueprint(cluster string, blueprint *v12.Blueprint) error
	DeleteBlueprint(cluster string, namespace string, name string) error
}

type ClusterMetadata struct {
	Region        string `json:"region"`
	Zone          string `json:"zone,omitempty"`
	VaultAuthPath string `json:"vaultAuthPath,omitempty"`
}

type Cluster struct {
	Name     string          `json:"name"`
	Metadata ClusterMetadata `json:"metadata"`
}

func CreateCluster(cm corev1.ConfigMap) Cluster {
	cluster := Cluster{
		Name: cm.Data["ClusterName"],
		Metadata: ClusterMetadata{
			Region:        cm.Data["Region"],
			Zone:          cm.Data["Zone"],
			VaultAuthPath: cm.Data["VaultAuthPath"],
		},
	}
	return cluster
}

// Decode json into runtime.Object, which is a pointer (such as &corev1.ConfigMapList)
func Decode(json string, scheme *runtime.Scheme, object runtime.Object) error {
	decoder := serializer.NewCodecFactory(scheme).UniversalDecoder()
	err := runtime.DecodeInto(decoder, []byte(json), object)
	if err != nil {
		return err
	}
	return nil
}
