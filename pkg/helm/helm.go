// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

var (
	settings = cli.New()
)

func getConfig(kubeNamespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)

	if kubeNamespace == "" {
		kubeNamespace = "default"
	}

	err := actionConfig.Init(settings.RESTClientGetter(), kubeNamespace, os.Getenv("HELM_DRIVER"), debug)
	if err != nil {
		return nil, err
	}

	return actionConfig, err
}

func debug(format string, v ...interface{}) {
	if settings.Debug {
		format = fmt.Sprintf("[debug] %s\n", format)
		_ = log.Output(2, fmt.Sprintf(format, v...))
	}
}

func ChartRef(hostname string, namespace string, name string, tagname string) string {
	return fmt.Sprintf("%s/%s/%s:%s", hostname, namespace, name, tagname)
}

// Interface of a helm chart
type Interface interface {
	Uninstall(kubeNamespace string, releaseName string) (*release.UninstallReleaseResponse, error)
	Install(chart *chart.Chart, kubeNamespace string, releaseName string, vals map[string]interface{}) (*release.Release, error)
	Upgrade(chart *chart.Chart, kubeNamespace string, releaseName string, vals map[string]interface{}) (*release.Release, error)
	Status(kubeNamespace string, releaseName string) (*release.Release, error)
	RegistryLogin(hostname string, username string, password string, insecure bool) error
	RegistryLogout(hostname string, username string) error
	ChartRemove(ref string) error
	ChartSave(chart *chart.Chart, ref string) error
	ChartLoad(ref string) (*chart.Chart, error)
	ChartPush(chart *chart.Chart, ref string) error
	ChartPull(ref string) error
}

// Fake implementation
type Fake struct {
}

// Uninstall helm release
func (r *Fake) Uninstall(kubeNamespace string, releaseName string) (*release.UninstallReleaseResponse, error) {
	res := &release.UninstallReleaseResponse{}
	return res, nil
}

// Install helm release
func (r *Fake) Install(chart *chart.Chart, kubeNamespace string, releaseName string, vals map[string]interface{}) (*release.Release, error) {
	rel := &release.Release{Info: &release.Info{}}
	return rel, nil
}

// Upgrade helm release
func (r *Fake) Upgrade(chart *chart.Chart, kubeNamespace string, releaseName string, vals map[string]interface{}) (*release.Release, error) {
	rel := &release.Release{Info: &release.Info{}}
	return rel, nil
}

// Status of helm release
func (r *Fake) Status(kubeNamespace string, releaseName string) (*release.Release, error) {
	rel := &release.Release{Info: &release.Info{}}
	return rel, nil
}

// RegistryLogin to docker registry v2
func (r *Fake) RegistryLogin(hostname string, username string, password string, insecure bool) error {
	return nil
}

// RegistryLogout to docker registry v2
func (r *Fake) RegistryLogout(hostname string, username string) error {
	return nil
}

// ChartRemove helm chart from cache
func (r *Fake) ChartRemove(ref string) error {
	return nil
}

// ChartSave helm chart from cache
func (r *Fake) ChartSave(chart *chart.Chart, ref string) error {
	return nil
}

// ChartLoad helm chart from cache
func (r *Fake) ChartLoad(ref string) (*chart.Chart, error) {
	return nil, nil
}

// ChartPush helm chart to repo
func (r *Fake) ChartPush(chart *chart.Chart, ref string) error {
	return nil
}

// ChartPull helm chart from repo
func (r *Fake) ChartPull(ref string) error {
	return nil
}

// Impl implementation
type Impl struct {
}

// Uninstall helm release
func (r *Impl) Uninstall(kubeNamespace string, releaseName string) (*release.UninstallReleaseResponse, error) {
	cfg, err := getConfig(kubeNamespace)
	if err != nil {
		return nil, err
	}
	uninstall := action.NewUninstall(cfg)
	return uninstall.Run(releaseName)
}

// Install helm release
func (r *Impl) Install(chart *chart.Chart, kubeNamespace string, releaseName string, vals map[string]interface{}) (*release.Release, error) {
	cfg, err := getConfig(kubeNamespace)
	if err != nil {
		return nil, err
	}
	install := action.NewInstall(cfg)
	install.ReleaseName = releaseName
	install.Namespace = kubeNamespace
	return install.Run(chart, vals)
}

// Upgrade helm release
func (r *Impl) Upgrade(chart *chart.Chart, kubeNamespace string, releaseName string, vals map[string]interface{}) (*release.Release, error) {
	cfg, err := getConfig(kubeNamespace)
	if err != nil {
		return nil, err
	}
	upgrade := action.NewUpgrade(cfg)
	upgrade.Namespace = kubeNamespace
	return upgrade.Run(releaseName, chart, vals)
}

// Status of helm release
func (r *Impl) Status(kubeNamespace string, releaseName string) (*release.Release, error) {
	cfg, err := getConfig(kubeNamespace)
	if err != nil {
		return nil, err
	}
	status := action.NewStatus(cfg)
	return status.Run(releaseName)
}

// RegistryLogin to docker registry v2
func (r *Impl) RegistryLogin(hostname string, username string, password string, insecure bool) error {
	if username != "" {
		cfg, err := getConfig("")
		if err != nil {
			return err
		}
		login := action.NewRegistryLogin(cfg)
		var buf bytes.Buffer
		return login.Run(&buf, hostname, username, password, insecure)
	}
	return nil
}

// RegistryLogout to docker registry v2
func (r *Impl) RegistryLogout(hostname string, username string) error {
	if username != "" {
		cfg, err := getConfig("")
		if err != nil {
			return err
		}
		logout := action.NewRegistryLogout(cfg)
		var buf bytes.Buffer
		return logout.Run(&buf, hostname)
	}
	return nil
}

// ChartRemove helm chart from cache
func (r *Impl) ChartRemove(ref string) error {
	cfg, err := getConfig("")
	if err != nil {
		return err
	}
	remove := action.NewChartRemove(cfg)
	var buf bytes.Buffer
	return remove.Run(&buf, ref)
}

// ChartSave helm chart from cache
func (r *Impl) ChartSave(chart *chart.Chart, ref string) error {
	cfg, err := getConfig("")
	if err != nil {
		return err
	}
	save := action.NewChartSave(cfg)
	var buf bytes.Buffer
	return save.Run(&buf, chart, ref)
}

// ChartLoad helm chart from cache
func (r *Impl) ChartLoad(ref string) (*chart.Chart, error) {
	cfg, err := getConfig("")
	if err != nil {
		return nil, err
	}
	load := action.NewChartLoad(cfg)
	return load.Run(ref)
}

// ChartPush helm chart to repo
func (r *Impl) ChartPush(chart *chart.Chart, ref string) error {
	cfg, err := getConfig("")
	if err != nil {
		return err
	}
	push := action.NewChartPush(cfg)
	var buf bytes.Buffer
	return push.Run(&buf, ref)
}

// ChartPull helm chart from repo
func (r *Impl) ChartPull(ref string) error {
	cfg, err := getConfig("")
	if err != nil {
		return err
	}
	push := action.NewChartPull(cfg)
	var buf bytes.Buffer
	return push.Run(&buf, ref)
}
