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
	mod := appApi.FybrikModule{}
	moduleMap := map[string]*appApi.FybrikModule{"mod1": &mod, "mod2": &mod}
	cluster := multicluster.Cluster{}
	clusters := []multicluster.Cluster{cluster}
	storageAccounts := []appApi.FybrikStorageAccount{}
	env := app.Environment{Modules: moduleMap, Clusters: clusters, StorageAccounts: storageAccounts, AttributeManager: nil}

	actions := []taxonomy.Action{
		{Name: "Reduct", AdditionalProperties: serde.Properties{}},
		{Name: "Encrypt", AdditionalProperties: serde.Properties{}},
	}
	evalOutput := adminconfig.EvaluatorOutput{}
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
