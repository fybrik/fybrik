// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	oras "oras.land/oras-go/pkg/registry"
)

var (
	debugOption = os.Getenv("HELM_DEBUG") == "true"
)

const chartsMountPath = "/opt/fybrik/charts/"

func getConfig(kubeNamespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	if kubeNamespace == "" {
		kubeNamespace = "default"
	}

	config := &genericclioptions.ConfigFlags{
		Namespace: &kubeNamespace,
	}
	err := actionConfig.Init(config, kubeNamespace, os.Getenv("HELM_DRIVER"), debugf)
	if err != nil {
		return nil, err
	}

	return actionConfig, err
}

func debugf(format string, v ...interface{}) {
	if debugOption {
		format = fmt.Sprintf("[debug] %s\n", format)
		_ = log.Output(2, fmt.Sprintf(format, v...))
	}
}

// Interface of a helm chart
type Interface interface {
	Uninstall(kubeNamespace string, releaseName string) (*release.UninstallReleaseResponse, error)
	Install(chart *chart.Chart, kubeNamespace string, releaseName string, vals map[string]interface{}) (*release.Release, error)
	Upgrade(chart *chart.Chart, kubeNamespace string, releaseName string, vals map[string]interface{}) (*release.Release, error)
	Status(kubeNamespace string, releaseName string) (*release.Release, error)
	RegistryLogin(hostname string, username string, password string, insecure bool) error
	RegistryLogout(hostname string) error
	Pull(ref string, destination string) error
	Load(ref string, chartPath string) (*chart.Chart, error)
	Package(chartPath string, destinationPath string, version string) error
	GetResources(kubeNamespace string, releaseName string) ([]*unstructured.Unstructured, error)
}

// Fake implementation
type Fake struct {
	release   *release.Release
	resources []*unstructured.Unstructured
}

// Uninstall helm release
func (r *Fake) Uninstall(kubeNamespace, releaseName string) (*release.UninstallReleaseResponse, error) {
	res := &release.UninstallReleaseResponse{}
	r.release = nil
	return res, nil
}

// Install helm release
func (r *Fake) Install(chrt *chart.Chart, kubeNamespace, releaseName string, vals map[string]interface{}) (*release.Release, error) {
	r.release = &release.Release{
		Name: releaseName,
		Info: &release.Info{Status: release.StatusDeployed},
	}
	return r.release, nil
}

// Upgrade helm release
func (r *Fake) Upgrade(chrt *chart.Chart, kubeNamespace, releaseName string, vals map[string]interface{}) (*release.Release, error) {
	r.release = &release.Release{
		Name: releaseName,
		Info: &release.Info{Status: release.StatusDeployed},
	}
	return r.release, nil
}

// Status of helm release
func (r *Fake) Status(kubeNamespace, releaseName string) (*release.Release, error) {
	return r.release, nil
}

// RegistryLogin to docker registry v2
func (r *Fake) RegistryLogin(hostname, username, password string, insecure bool) error {
	return nil
}

// RegistryLogout to docker registry v2
func (r *Fake) RegistryLogout(hostname string) error {
	return nil
}

// ChartPull helm chart from repo
func (r *Fake) Pull(ref, destination string) error {
	return nil
}

// Load helm chart
func (r *Fake) Load(ref, chartPath string) (*chart.Chart, error) {
	return nil, nil
}

// GetResources returns allocated resources for the specified release (their current state)
func (r *Fake) GetResources(kubeNamespace, releaseName string) ([]*unstructured.Unstructured, error) {
	return r.resources, nil
}

// Package helm chart from repo
func (r *Fake) Package(chartPath, destinationPath, version string) error {
	return nil
}

func NewEmptyFake() *Fake {
	return &Fake{
		release:   &release.Release{Info: &release.Info{}},
		resources: make([]*unstructured.Unstructured, 0),
	}
}

func NewFake(rls *release.Release, resources []*unstructured.Unstructured) *Fake {
	return &Fake{
		release:   rls,
		resources: resources,
	}
}

// Impl implementation
type Impl struct {
}

// Uninstall helm release
func (r *Impl) Uninstall(kubeNamespace, releaseName string) (*release.UninstallReleaseResponse, error) {
	cfg, err := getConfig(kubeNamespace)
	if err != nil {
		return nil, err
	}
	uninstall := action.NewUninstall(cfg)
	return uninstall.Run(releaseName)
}

// Load helm chart
func (r *Impl) Load(ref, chartPath string) (*chart.Chart, error) {
	// check for chart mounted in container
	chrt, err := loader.Load(chartsMountPath + ref)
	if err == nil {
		return chrt, nil
	}
	// Construct the packed chart path
	chartRef, err := parseReference(ref)
	if err != nil {
		return nil, err
	}
	_, chartName := filepath.Split(chartRef.Repository)
	packedChartPath := fmt.Sprintf("%s/%s-%s.tgz", chartPath, chartName, chartRef.Reference)

	return loader.Load(packedChartPath)
}

// Install helm release from packaged chart
func (r *Impl) Install(chrt *chart.Chart, kubeNamespace, releaseName string, vals map[string]interface{}) (*release.Release, error) {
	cfg, err := getConfig(kubeNamespace)
	if err != nil {
		return nil, err
	}
	install := action.NewInstall(cfg)
	install.ReleaseName = releaseName
	install.Namespace = kubeNamespace
	return install.Run(chrt, vals)
}

// Upgrade helm release
func (r *Impl) Upgrade(chrt *chart.Chart, kubeNamespace, releaseName string, vals map[string]interface{}) (*release.Release, error) {
	cfg, err := getConfig(kubeNamespace)
	if err != nil {
		return nil, err
	}
	upgrade := action.NewUpgrade(cfg)
	upgrade.Namespace = kubeNamespace
	return upgrade.Run(releaseName, chrt, vals)
}

// Status of helm release
func (r *Impl) Status(kubeNamespace, releaseName string) (*release.Release, error) {
	cfg, err := getConfig(kubeNamespace)
	if err != nil {
		return nil, err
	}
	status := action.NewStatus(cfg)
	return status.Run(releaseName)
}

// RegistryLogin to docker registry v2
func (r *Impl) RegistryLogin(hostname, username, password string, insecure bool) error {
	var settings = cli.New()
	client, err := registry.NewClient(registry.ClientOptDebug(settings.Debug),
		registry.ClientOptWriter(os.Stdout),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	)
	if err != nil {
		return err
	}
	return client.Login(hostname, registry.LoginOptBasicAuth(username, password),
		registry.LoginOptInsecure(insecure))
}

// RegistryLogout to docker registry v2
func (r *Impl) RegistryLogout(hostname string) error {
	var settings = cli.New()
	client, err := registry.NewClient(registry.ClientOptDebug(settings.Debug),
		registry.ClientOptWriter(os.Stdout),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	)
	if err != nil {
		return err
	}
	return client.Logout(hostname)
}

// Package helm chart from repo
func (r *Impl) Package(chartPath, destinationPath, version string) error {
	client := action.NewPackage()
	if version != "" {
		client.Version = version
	}
	client.Destination = destinationPath

	_, err := client.Run(chartPath, nil)
	return err
}

// Pull helm chart from repo
func (r *Impl) Pull(ref, destination string) error {
	// if chart mounted in container, no need to pull
	if _, err := os.Stat(chartsMountPath + ref); err == nil {
		return nil
	}

	chartRef, err := parseReference(ref)
	if err != nil {
		return err
	}
	cfg, err := getConfig("")
	if err != nil {
		return err
	}

	var settings = cli.New()
	registryClient, err := registry.NewClient(registry.ClientOptDebug(settings.Debug),
		registry.ClientOptWriter(os.Stdout),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	)
	if err != nil {
		return err
	}
	cfg.RegistryClient = registryClient
	client := action.NewPullWithOpts(action.WithConfig(cfg))
	client.Version = chartRef.Reference
	client.Settings = settings
	client.DestDir = destination
	_, err = client.Run("oci://" + chartRef.Registry + "/" + chartRef.Repository)
	return err
}

// GetResources returns allocated resources for the specified release (their current state)
func (r *Impl) GetResources(kubeNamespace, releaseName string) ([]*unstructured.Unstructured, error) {
	resources := make([]*unstructured.Unstructured, 0)
	var rel *release.Release
	var config *action.Configuration
	var err error
	var resourceList kube.ResourceList
	config, err = getConfig(kubeNamespace)
	if err != nil || config == nil {
		return resources, err
	}
	status := action.NewStatus(config)
	rel, err = status.Run(releaseName)
	if err != nil {
		return resources, err
	}
	resourceList, err = config.KubeClient.Build(bytes.NewBufferString(rel.Manifest), false)
	if err != nil {
		return resources, err
	}

	for _, res := range resourceList {
		if err := res.Get(); err != nil {
			return resources, err
		}
		obj := res.Object
		if unstr, ok := obj.(*unstructured.Unstructured); ok {
			resources = append(resources, unstr)
		} else {
			return resources, errors.New("invalid runtime object")
		}
	}
	return resources, nil
}

// parseReference will parse and validate the reference, and clean tags when
// applicable tags are only cleaned when plus (+) signs are present, and are
// converted to underscores (_) before pushing
// See https://github.com/helm/helm/issues/10166
// From https://github.com/helm/helm/blob/49819b4ef782e80b0c7f78c30bd76b51ebb56dc8/pkg/registry/util.go#L112
func parseReference(raw string) (oras.Reference, error) {
	// The sole possible reference modification is replacing plus (+) signs
	// present in tags with underscores (_). To do this properly, we first
	// need to identify a tag, and then pass it on to the reference parser
	// NOTE: Passing immediately to the reference parser will fail since (+)
	// signs are an invalid tag character, and simply replacing all plus (+)
	// occurrences could invalidate other portions of the URI
	parts := strings.Split(raw, ":")
	if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], "/") {
		tag := parts[len(parts)-1]

		if tag != "" {
			// Replace any plus (+) signs with known underscore (_) conversion
			newTag := strings.ReplaceAll(tag, "+", "_")
			raw = strings.ReplaceAll(raw, tag, newTag)
		}
	}

	return oras.ParseReference(raw)
}
