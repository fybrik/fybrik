// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"os"
	"strconv"
	"testing"
)

func buildTestChart() *chart.Chart {
	testManifestWithHook := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  annotations:
    "helm.sh/hook": post-install,pre-delete,post-upgrade
data:
  key: value`

	return &chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion: "v1",
			Name:       "hello",
			Version:    "0.1.0",
		},
		Templates: []*chart.File{
			{Name: "templates/hooks", Data: []byte(testManifestWithHook)},
		},
	}
}

func Log(t *testing.T, label string, err error) {
	if err == nil {
		err = fmt.Errorf("succeeded")
	}
	t.Logf("%s: %s", label, err)
}

var (
	kubeNamespace = "default"
	releaseName   = "test-install-release"
	chartName     = "test-chart"
	tagName       = "0.1.0"
	hostname      = os.Getenv("DOCKER_HOSTNAME")
	namespace     = os.Getenv("DOCKER_NAMESPACE")
	username      = os.Getenv("DOCKER_USERNAME")
	password      = os.Getenv("DOCKER_PASSWORD")
	insecure, _   = strconv.ParseBool(os.Getenv("DOCKER_INSECURE"))
	chartRef      = ChartRef(hostname, namespace, chartName, tagName)
	impl          = new(Impl)
)

func TestHelmCache(t *testing.T) {
	var err error
	origChart := buildTestChart()

	err = impl.ChartSave(origChart, chartRef)
	assert.Nil(t, err)
	Log(t, "save chart", err)

	chart, err := impl.ChartLoad(chartRef)
	assert.Nil(t, err)
	Log(t, "load chart", err)

	err = impl.ChartRemove(chartRef)
	assert.Nil(t, err)
	Log(t, "remove chart", err)

	assert.Equal(t, origChart.Metadata.Name, chart.Metadata.Name, "expected loaded chart equals saved chart")
}

func TestHelmRegistry(t *testing.T) {
	var err error
	origChart := buildTestChart()

	err = impl.RegistryLogin(hostname, username, password, insecure)
	assert.Nil(t, err)
	Log(t, "registry login", err)

	err = impl.ChartSave(origChart, chartRef)
	assert.Nil(t, err)
	err = impl.ChartPush(origChart, chartRef)
	assert.Nil(t, err)
	Log(t, "push chart", err)

	err = impl.ChartPull(chartRef)
	assert.Nil(t, err)
	Log(t, "pull chart", err)
	chart, err := impl.ChartLoad(chartRef)
	assert.Nil(t, err)
	assert.Equal(t, origChart.Metadata.Name, chart.Metadata.Name, "expected pushed chart equals pulled chart")

	err = impl.RegistryLogout(hostname, username)
	assert.Nil(t, err)
	Log(t, "registry logout", err)
}

func TestHelmRelease(t *testing.T) {
	var err error
	origChart := buildTestChart()

	_, _ = impl.Uninstall(kubeNamespace, releaseName)
	vals := map[string]interface{}{
		"data": map[string]interface{}{
			"key": "value1",
		},
	}
	_, err = impl.Install(origChart, kubeNamespace, releaseName, vals)
	assert.Nil(t, err)
	Log(t, "install", err)

	_, err = impl.Upgrade(origChart, kubeNamespace, releaseName, vals)
	assert.Nil(t, err)
	Log(t, "upgrade", err)

	_, err = impl.Status(kubeNamespace, releaseName)
	assert.Nil(t, err)
	Log(t, "status", err)

	_, err = impl.Uninstall(kubeNamespace, releaseName)
	assert.Nil(t, err)
	Log(t, "uninstall", err)
}
