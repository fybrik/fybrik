// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	oras "oras.land/oras-go/pkg/registry"

	"fybrik.io/fybrik/pkg/environment"
)

// Relevant only when helm charts are placed in
// local directory.
const chartsDir = "/charts/"

// Interface of a helm chart
type Interface interface {
	GetConfig(kubeNamespace string, log action.DebugLog) (*action.Configuration, error)

	Uninstall(cfg *action.Configuration, releaseName string) (*release.UninstallReleaseResponse, error)
	Install(ctx context.Context, cfg *action.Configuration, chart *chart.Chart, kubeNamespace string,
		releaseName string, vals map[string]interface{}) (*release.Release, error)
	Upgrade(ctx context.Context, cfg *action.Configuration, chart *chart.Chart, kubeNamespace string,
		releaseName string, vals map[string]interface{}) (*release.Release, error)
	Status(cfg *action.Configuration, releaseName string) (*release.Release, error)
	Pull(cfg *action.Configuration, ref string, destination string) error
	IsInstalled(cfg *action.Configuration, releaseName string) (bool, error)

	RegistryLogin(hostname string, username string, password string, insecure bool) error
	RegistryLogout(hostname string) error
	Load(ref string, chartPath string) (*chart.Chart, error)
	Package(chartPath string, destinationPath string, version string) error
	GetResources(cfg *action.Configuration, manifest string) ([]*unstructured.Unstructured, error)
}

// Fake implementation
type Fake struct {
	release   *release.Release
	resources []*unstructured.Unstructured
}

func (r *Fake) GetConfig(kubeNamespace string, log action.DebugLog) (*action.Configuration, error) {
	return nil, nil
}

// Uninstall helm release
func (r *Fake) Uninstall(cfg *action.Configuration, releaseName string) (*release.UninstallReleaseResponse, error) {
	res := &release.UninstallReleaseResponse{}
	r.release = nil
	return res, nil
}

// Install helm release
func (r *Fake) Install(ctx context.Context, cfg *action.Configuration, chrt *chart.Chart, kubeNamespace,
	releaseName string, vals map[string]interface{}) (*release.Release, error) {
	r.release = &release.Release{
		Name: releaseName,
		Info: &release.Info{Status: release.StatusDeployed},
	}
	return r.release, nil
}

// Upgrade helm release
func (r *Fake) Upgrade(ctx context.Context, cfg *action.Configuration, chrt *chart.Chart, kubeNamespace,
	releaseName string, vals map[string]interface{}) (*release.Release, error) {
	r.release = &release.Release{
		Name: releaseName,
		Info: &release.Info{Status: release.StatusDeployed},
	}
	return r.release, nil
}

// Status of helm release
func (r *Fake) Status(cfg *action.Configuration, releaseName string) (*release.Release, error) {
	return r.release, nil
}

func (r *Fake) IsInstalled(cfg *action.Configuration, releaseName string) (bool, error) {
	return r.release.Info.Status == release.StatusDeployed, nil
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
func (r *Fake) Pull(cfg *action.Configuration, ref, destination string) error {
	return nil
}

// Load helm chart
func (r *Fake) Load(ref, chartPath string) (*chart.Chart, error) {
	return nil, nil
}

// GetResources returns allocated resources for the specified release (their current state)
func (r *Fake) GetResources(cfg *action.Configuration, manifest string) ([]*unstructured.Unstructured, error) {
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
	// if set, the "Load" and "pull" methods will try to check locally mounted charts
	localChartsMountPath string
	discoveryBurst       int
	discoveryQPS         float32
}

func NewHelmerImpl(chartsPath string) *Impl {
	impl := Impl{localChartsMountPath: chartsPath}
	// If an error exists it is logged in LogEnvVariables
	impl.discoveryBurst, _ = environment.GetDiscoveryBurst()
	// If an error exists it is logged in LogEnvVariables
	impl.discoveryQPS, _ = environment.GetDiscoveryQPS()
	return &impl
}

// Uninstall helm release
func (r *Impl) Uninstall(cfg *action.Configuration, releaseName string) (*release.UninstallReleaseResponse, error) {
	uninstall := action.NewUninstall(cfg)
	return uninstall.Run(releaseName)
}

// Load helm chart
func (r *Impl) Load(ref, chartPath string) (*chart.Chart, error) {
	if r.localChartsMountPath != "" {
		var err error
		// check for chart mounted in container
		chrt, err := loader.Load(r.localChartsMountPath + chartsDir + ref)
		if err == nil {
			return chrt, nil
		}
		return nil, err
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
func (r *Impl) Install(ctx context.Context, cfg *action.Configuration, chrt *chart.Chart, kubeNamespace,
	releaseName string, vals map[string]interface{}) (*release.Release, error) {
	install := action.NewInstall(cfg)
	install.ReleaseName = releaseName
	install.Namespace = kubeNamespace

	return install.RunWithContext(ctx, chrt, vals)
}

// Upgrade helm release
func (r *Impl) Upgrade(ctx context.Context, cfg *action.Configuration, chrt *chart.Chart, kubeNamespace,
	releaseName string, vals map[string]interface{}) (*release.Release, error) {
	upgrade := action.NewUpgrade(cfg)
	upgrade.Namespace = kubeNamespace

	return upgrade.RunWithContext(ctx, releaseName, chrt, vals)
}

// Status of helm release
func (r *Impl) Status(cfg *action.Configuration, releaseName string) (*release.Release, error) {
	status := action.NewStatus(cfg)
	return status.Run(releaseName)
}

func (r *Impl) IsInstalled(cfg *action.Configuration, releaseName string) (bool, error) {
	histClient := action.NewHistory(cfg)
	histClient.Max = 1

	_, err := histClient.Run(releaseName)
	if err == driver.ErrReleaseNotFound {
		return false, nil
	}
	return true, err
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
func (r *Impl) Pull(cfg *action.Configuration, ref, destination string) error {
	if r.localChartsMountPath != "" {
		// if chart mounted in container, no need to pull
		if _, err := os.Stat(r.localChartsMountPath + chartsDir + ref); err == nil {
			return nil
		}
	}

	chartRef, err := parseReference(ref)
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

func (r *Impl) GetConfig(kubeNamespace string, log action.DebugLog) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	if kubeNamespace == "" {
		kubeNamespace = "default"
	}

	config := &genericclioptions.ConfigFlags{
		Namespace: &kubeNamespace,
	}
	if r.discoveryBurst != -1 {
		config.WithDiscoveryBurst(r.discoveryBurst)
	}
	if r.discoveryQPS != -1 {
		config.WithDiscoveryQPS(r.discoveryQPS)
	}

	err := actionConfig.Init(config, kubeNamespace, os.Getenv("HELM_DRIVER"), log)
	if err != nil {
		return nil, err
	}

	return actionConfig, err
}

// GetResources returns allocated resources for the specified by its manifest release (their current state)
func (r *Impl) GetResources(cfg *action.Configuration, manifest string) ([]*unstructured.Unstructured, error) {
	resources := make([]*unstructured.Unstructured, 0)

	resourceList, err := cfg.KubeClient.Build(bytes.NewBufferString(manifest), false)
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
