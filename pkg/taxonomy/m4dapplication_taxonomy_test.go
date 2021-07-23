// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy


import (
	"fmt"
	"io/ioutil"
	"testing"

	tax "github.com/mesh-for-data/mesh-for-data/config/taxonomy"
	"sigs.k8s.io/yaml"
)

var (
	AppTaxValsName = "../../charts/m4d/files/taxonomy/application.values.schema.json"
)

func TestAppTaxonomy(t *testing.T) {
	applicationYaml, err := ioutil.ReadFile("../../samples/kubeflow/m4dapplication.yaml")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	appInfoGood, err := yaml.YAMLToJSON(applicationYaml)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	interfaceGood := appInfoGood

	appInfoBadYaml, err := ioutil.ReadFile("../../manager/testdata/unittests/m4dapplication-appInfoErrors.yaml")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	appInfoBad, err := yaml.YAMLToJSON(appInfoBadYaml)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	interfaceBadYaml, err := ioutil.ReadFile("../../manager/testdata/unittests/m4dapplication-interfaceErrors.yaml")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	interfaceBad, err := yaml.YAMLToJSON(interfaceBadYaml)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	tax.ValidateTaxonomy(t, AppTaxValsName, string(appInfoGood), "appInfoGood", true)
	tax.ValidateTaxonomy(t, AppTaxValsName, string(interfaceGood), "interfaceGood", true)
	tax.ValidateTaxonomy(t, AppTaxValsName, string(appInfoBad), "appInfoBad", false)
	tax.ValidateTaxonomy(t, AppTaxValsName, string(interfaceBad), "interfaceBad", false)
}
