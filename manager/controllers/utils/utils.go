// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"io"
	"net/url"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/retry"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/utils"
)

// IsDenied returns true if the data access is denied
func IsDenied(actionName taxonomy.ActionName) bool {
	return actionName == "Deny" // TODO FIX THIS
}

// Generating a release name based on the blueprint module and application name/uuid
func GetReleaseName(applicationName, uuid, instanceName string) string {
	fullName := applicationName + uuid + "-" + instanceName
	return utils.HelmConformName(fullName)
}

// Create a name for a step in a blueprint.
// Since this is part of the name of a release, this should be done in a central location to make testing easier
func CreateStepName(moduleName, assetID string, moduleScope fapp.CapabilityScope) string {
	if moduleScope == fapp.Asset {
		return moduleName + "-" + utils.Hash(assetID, utils.StepNameHashLength)
	}
	return moduleName
}

// UpdateStatus updates the resource status
func UpdateStatus(ctx context.Context, cl client.Client, obj client.Object, previousStatus interface{}) error {
	err := cl.Status().Update(ctx, obj)
	if err == nil {
		return nil
	}
	if !errors.IsConflict(err) {
		return err
	}
	values, err := utils.StructToMap(obj)
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

// ParseRawURL parses a string to return URl even if the schema is not set
// from the url.Parse comments "Trying to parse a hostname and path
// without a scheme is invalid but may not necessarily return an error, due to parsing ambiguities."
func ParseRawURL(rawURL string) (*url.URL, error) {
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL // we want to prevent parsing "host:port" to schema, which is done by url.Parse
	}
	u, err := url.Parse(rawURL)
	return u, err
}

func ExecPod(restClient restclient.Interface, config *restclient.Config, podName string, namespace string,
	command string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	req := restClient.Post().Resource("pods").Name(podName).
		Namespace(namespace).SubResource("exec")
	option := &corev1.PodExecOptions{
		Command: cmd,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	if stdin == nil {
		option.Stdin = false
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})

	return err
}
