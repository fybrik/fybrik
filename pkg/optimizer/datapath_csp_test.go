// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"testing"

	appApi "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/app"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/serde"
)

func TestBuildModel(t *testing.T) {
	encryptAction := appApi.ModuleSupportedAction{Name: "Encrypt"}
	reductAction := appApi.ModuleSupportedAction{Name: "Reduct"}
	copyAction := appApi.ModuleSupportedAction{Name: "Copy"}
	modCap1 := appApi.ModuleCapability{Capability: "read", Scope: "asset", Actions: []appApi.ModuleSupportedAction{encryptAction, reductAction}}
	modCap2 := appApi.ModuleCapability{Capability: "read", Scope: "asset", Actions: []appApi.ModuleSupportedAction{encryptAction}}
	modCap3 := appApi.ModuleCapability{Capability: "copy", Scope: "asset", Actions: []appApi.ModuleSupportedAction{copyAction}}
	modSpec1 := appApi.FybrikModuleSpec{Capabilities: []appApi.ModuleCapability{modCap1, modCap3}}
	modSpec2 := appApi.FybrikModuleSpec{Capabilities: []appApi.ModuleCapability{modCap2}}
	mod1 := appApi.FybrikModule{Spec: modSpec1}
	mod2 := appApi.FybrikModule{Spec: modSpec2}
	mod1.Name = "ReaderCopier"
	mod2.Name = "Reader"
	moduleMap := map[string]*appApi.FybrikModule{mod1.Name: &mod1, mod2.Name: &mod2}
	cluster := multicluster.Cluster{}
	clusters := []multicluster.Cluster{cluster}
	storageAccounts := []appApi.FybrikStorageAccount{}
	env := app.Environment{Modules: moduleMap, Clusters: clusters, StorageAccounts: storageAccounts, AttributeManager: nil}

	actions := []taxonomy.Action{
		{Name: "Reduct", AdditionalProperties: serde.Properties{}},
		{Name: "Encrypt", AdditionalProperties: serde.Properties{}},
	}

	decision := adminconfig.Decision{Deploy: adminconfig.StatusFalse}
	decisionMap := adminconfig.DecisionPerCapabilityMap{"copy": decision}
	evalOutput := adminconfig.EvaluatorOutput{ConfigDecisions: decisionMap}
	dataInfo := app.DataInfo{
		DataDetails:         nil,
		Context:             nil,
		Configuration:       evalOutput,
		WorkloadCluster:     cluster,
		Actions:             actions,
		StorageRequirements: map[taxonomy.ProcessingLocation][]taxonomy.Action{},
	}
	dpCSP := NewDataPathCSP(&dataInfo, &env)
	err := dpCSP.BuildFzModel(3)
	if err != nil {
		t.Errorf("Failed building a CSP model: %s", err)
	}
}
