// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package datauser

import (
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// CRDPlural is the plural name of the resource
	CRDPlural string = "m4dapplications"

	// CRDGroup is the group with which the resource is associated
	CRDGroup string = "app.m4d.ibm.com"

	// CRDVersion version of the resource's implementation
	CRDVersion string = "v1alpha1"

	// FullCRDName is the concatenation of the plural name + the group
	FullCRDName string = CRDPlural + "." + CRDGroup
)

// K8sInit assumes we are running within the kubernetes cluster
func K8sInit() (*DMAClient, error) {
	// For local testing set KUBECONFIG to $HOME/.kube/config
	// It is unset for deployment

	kubeconfigArg := ""
	if kubeconfigpath := os.Getenv("KUBECONFIG"); kubeconfigpath != "" {
		kubeconf := flag.String("kubeconf", kubeconfigpath, "Path to a kube config. Only required if out-of-cluster.")
		flag.Parse()
		kubeconfigArg = *kubeconf
	}

	config, err := GetClientConfig(kubeconfigArg)

	if err != nil {
		return nil, err
	}

	dmaClient, err := NewForConfig(config)
	return dmaClient, err
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

// GetClientConfig returns config for accessing kubernetes
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

// NewForConfig configures the client
func NewForConfig(cfg *rest.Config) (*DMAClient, error) {
	config := *cfg
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.UserAgent = rest.DefaultKubernetesUserAgent()
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &DMAClient{client: client, namespace: GetCurrentNamespace(), plural: CRDPlural, codec: runtime.NewParameterCodec(scheme.Scheme)}, nil
}
