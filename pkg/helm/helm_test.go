// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"

	"fybrik.io/fybrik/pkg/environment"
)

func buildTestChart() *chart.Chart {
	testManifestWithHook := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
data:
  key: value`

	return &chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion: "v1",
			Name:       "test-chart",
			Type:       "application",
			Version:    "0.1.0",
		},
		Templates: []*chart.File{
			{Name: "templates/config.yaml", Data: []byte(testManifestWithHook)},
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
	chartName     = "mychart"
	tagName       = "0.7.0"
	hostname      = os.Getenv("DOCKER_HOSTNAME")
	namespace     = os.Getenv("DOCKER_NAMESPACE")
	username      = os.Getenv("DOCKER_USERNAME")
	password      = os.Getenv("DOCKER_PASSWORD")
	insecure, _   = strconv.ParseBool(os.Getenv("DOCKER_INSECURE"))
	chartRef      = ChartRef(hostname, namespace, chartName, tagName)
)

func ChartRef(hostname, namespace, name, tagname string) string {
	return fmt.Sprintf("%s/%s/%s:%s", hostname, namespace, name, tagname)
}

func TestHelmRegistry(t *testing.T) {
	tmpChart := os.Getenv("TMP_CHART")
	var err error

	// Test should only run as integration test if registry is available
	if _, isSet := os.LookupEnv("DOCKER_HOSTNAME"); !isSet {
		t.Skip("No integration environment found. Skipping test...")
	}

	tmpDir, err := ioutil.TempDir(environment.GetDataDir(), "test-helm-")
	if err != nil {
		t.Errorf("Unable to create temporary directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	pulledChartDestPath := path.Join(tmpDir, "pulledChartDir")
	packedChartDir := path.Join(tmpDir, "packedChartDir")
	err = os.Mkdir(pulledChartDestPath, 0700)
	if err != nil {
		t.Errorf("Unable to setup test temp charts directory: %s", err)
	}
	err = os.Mkdir(packedChartDir, 0700)
	if err != nil {
		t.Errorf("Unable to setup test temp charts directory: %s", err)
	}
	impl := NewHelmerImpl("")

	if username != "" && password != "" {
		err = impl.RegistryLogin(hostname, username, password, insecure)
		assert.Nil(t, err)
		Log(t, "registry login", err)
	}

	err = impl.Package(tmpChart, packedChartDir, tagName)
	assert.Nil(t, err)
	Log(t, "package chart", err)

	cfg, err := impl.GetConfig("", t.Logf)
	assert.Nil(t, err)
	err = impl.Pull(cfg, chartRef, pulledChartDestPath)
	assert.Nil(t, err)
	Log(t, "pull chart", err)

	pulledChart, err := impl.Load(chartRef, pulledChartDestPath)
	assert.Nil(t, err)
	Log(t, "load chart", err)

	packagePath := packedChartDir + "/mychart-0.7.0.tgz"
	packedChart, err := loader.Load(packagePath)
	assert.Nil(t, err)

	assert.Equal(t, packedChart.Metadata.Name, pulledChart.Metadata.Name, "expected loaded chart equals saved chart")

	if username != "" && password != "" {
		err = impl.RegistryLogout(hostname)
		assert.Nil(t, err)
		Log(t, "registry logout", err)
	}
}

// TestLocalChartsMount tests the case where the helm charts
// are located in a local directory and thus there is no need
// to retrive them from the registry.
func TestLocalChartsMount(t *testing.T) {
	tmpChart := os.Getenv("TMP_CHART")
	var err error
	if tmpChart == "" {
		t.Skip("No chart path was defined as environment. Skipping test...")
	}

	rootPath := filepath.Dir(tmpChart)
	// remove the "charts" suffix from the file path
	impl := NewHelmerImpl(filepath.Dir(rootPath))

	tmpDir, err := ioutil.TempDir(environment.GetDataDir(), "test-helm-")
	if err != nil {
		t.Errorf("Unable to create temporary directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	packedChartDir := path.Join(tmpDir, "packedChartDir")
	err = os.Mkdir(packedChartDir, 0700)
	if err != nil {
		t.Errorf("Unable to setup test temp charts directory: %s", err)
	}

	err = impl.Package(tmpChart, packedChartDir, tagName)
	assert.Nil(t, err)
	Log(t, "package chart", err)

	cfg, err := impl.GetConfig("", t.Logf)
	assert.Nil(t, err)
	err = impl.Pull(cfg, chartName, "")
	assert.Nil(t, err)
	Log(t, "pull chart", err)

	pulledChart, err := impl.Load(chartName, "")
	assert.Nil(t, err)
	Log(t, "load chart", err)

	packagePath := packedChartDir + "/mychart-0.7.0.tgz"
	packedChart, err := loader.Load(packagePath)
	assert.Nil(t, err)

	assert.Equal(t, packedChart.Metadata.Name, pulledChart.Metadata.Name, "expected loaded chart equals saved chart")
}

func TestHelmRelease(t *testing.T) {
	// Test should only run as integration test if registry is available
	if _, isSet := os.LookupEnv("DOCKER_HOSTNAME"); !isSet {
		t.Skip("No integration environment found. Skipping test...")
	}
	var err error
	origChart := buildTestChart()
	impl := NewHelmerImpl("")
	cfg, err := impl.GetConfig(kubeNamespace, t.Logf)
	assert.Nil(t, err)

	_, _ = impl.Uninstall(cfg, releaseName)
	vals := map[string]interface{}{
		"data": map[string]interface{}{
			"key": "value1",
		},
	}
	cntx := context.Background()
	_, err = impl.Install(cntx, cfg, origChart, kubeNamespace, releaseName, vals)
	assert.Nil(t, err)
	Log(t, "install", err)

	_, err = impl.Upgrade(cntx, cfg, origChart, kubeNamespace, releaseName, vals)
	assert.Nil(t, err)
	Log(t, "upgrade", err)

	var rel *release.Release
	assert.Eventually(t, func() bool {
		rel, err = impl.Status(cfg, releaseName)
		assert.Nil(t, err)
		return rel.Info.Status == release.StatusDeployed
	}, time.Minute, time.Second)
	Log(t, "status", err)

	var resources []*unstructured.Unstructured
	for versionKind := range rel.Info.Resources {
		for _, obj := range rel.Info.Resources[versionKind] {
			if unstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj); err == nil {
				resources = append(resources, &unstructured.Unstructured{Object: unstr})
			} else {
				Log(t, "status", errors.New("status: could not obtain resources"))
			}
		}
	}
	assert.Len(t, resources, 1)
	computedResult, _ := kstatus.Compute(resources[0])
	assert.Equal(t, kstatus.CurrentStatus, computedResult.Status)
	Log(t, "getResources", err)

	_, err = impl.Uninstall(cfg, releaseName)
	assert.Nil(t, err)
	Log(t, "uninstall", err)
}
