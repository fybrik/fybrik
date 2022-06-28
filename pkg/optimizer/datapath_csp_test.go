// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package optimizer

import (
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appApi "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/datapath"
	"fybrik.io/fybrik/pkg/infrastructure"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/serde"
)

func getTestEnv() *datapath.Environment {
	s3CSVInterface := taxonomy.Interface{Protocol: "s3", DataFormat: "csv"}
	s3ParquetInterface := taxonomy.Interface{Protocol: "s3", DataFormat: "parquet"}
	db2ParquetInterface := taxonomy.Interface{Protocol: "db2", DataFormat: "parquet"}
	arrowCSVInterface := taxonomy.Interface{Protocol: "fybrik-arrow-flight", DataFormat: "csv"}
	inOutS3Parquet := appApi.ModuleInOut{Source: &s3ParquetInterface, Sink: &s3ParquetInterface}
	ins3outDB2parquet := appApi.ModuleInOut{Source: &s3ParquetInterface, Sink: &db2ParquetInterface}
	inDB2outS3parquet := appApi.ModuleInOut{Source: &db2ParquetInterface, Sink: &s3ParquetInterface}
	inS3outArrowCSV := appApi.ModuleInOut{Source: &s3CSVInterface, Sink: &arrowCSVInterface}
	inParquetoutCsvS3 := appApi.ModuleInOut{Source: &s3ParquetInterface, Sink: &s3CSVInterface}

	encryptAction := appApi.ModuleSupportedAction{Name: "Encrypt"}
	reductAction := appApi.ModuleSupportedAction{Name: "Reduct"}
	copyAction := appApi.ModuleSupportedAction{Name: "Copy"}
	modCap1 := appApi.ModuleCapability{
		Capability: "read", Scope: "asset",
		Actions:             []appApi.ModuleSupportedAction{encryptAction, reductAction},
		SupportedInterfaces: []appApi.ModuleInOut{ins3outDB2parquet, inDB2outS3parquet, inOutS3Parquet},
	}
	modCap2 := appApi.ModuleCapability{
		Capability: "read", Scope: "asset",
		Actions:             []appApi.ModuleSupportedAction{encryptAction},
		SupportedInterfaces: []appApi.ModuleInOut{inParquetoutCsvS3, inS3outArrowCSV},
	}
	modCap3 := appApi.ModuleCapability{Capability: "copy", Scope: "asset", Actions: []appApi.ModuleSupportedAction{copyAction}}
	modSpec1 := appApi.FybrikModuleSpec{Capabilities: []appApi.ModuleCapability{modCap1, modCap3}}
	modSpec2 := appApi.FybrikModuleSpec{Capabilities: []appApi.ModuleCapability{modCap2}}
	mod1 := appApi.FybrikModule{Spec: modSpec1}
	mod2 := appApi.FybrikModule{Spec: modSpec2}
	mod1.Name = "ReaderCopier"
	mod2.Name = "Reader"
	moduleMap := map[string]*appApi.FybrikModule{mod1.Name: &mod1, mod2.Name: &mod2}
	cluster1 := multicluster.Cluster{Name: "cluster1"}
	cluster1Cost := taxonomy.InfrastructureElement{Name: "ClusterCost", Value: "56", MetricName: "cost",
		Object: taxonomy.Cluster, Instance: "cluster1"}
	cluster2 := multicluster.Cluster{Name: "cluster2"}
	cluster2Cost := taxonomy.InfrastructureElement{Name: "ClusterCost", Value: "1", MetricName: "cost",
		Object: taxonomy.Cluster, Instance: "cluster2"}
	cluster3 := multicluster.Cluster{Name: "cluster3"}
	cluster3Cost := taxonomy.InfrastructureElement{Name: "ClusterCost", Value: "58", MetricName: "cost",
		Object: taxonomy.Cluster, Instance: "cluster3"}
	clusters := []multicluster.Cluster{cluster1, cluster2, cluster3}
	sa1 := appApi.FybrikStorageAccount{ObjectMeta: metav1.ObjectMeta{Name: "sa1"}, Spec: appApi.FybrikStorageAccountSpec{Region: "us-south"}}
	storageAccounts := []*appApi.FybrikStorageAccount{&sa1}
	attrManager := infrastructure.AttributeManager{
		Attributes: []taxonomy.InfrastructureElement{cluster1Cost, cluster2Cost, cluster3Cost},
		Metrics:    infrastructure.MetricsDictionary{},
	}
	attrManager.Metrics["cost"] = taxonomy.InfrastructureMetrics{Name: "cost", Type: "numeric", Scale: &taxonomy.RangeType{Min: 0, Max: 200}}
	env := datapath.Environment{Modules: moduleMap, Clusters: clusters, StorageAccounts: storageAccounts, AttributeManager: &attrManager}
	return &env
}

func getDataInfo(env *datapath.Environment) *datapath.DataInfo {
	actions := []taxonomy.Action{
		{Name: "Reduct", AdditionalProperties: serde.Properties{}},
		{Name: "Encrypt", AdditionalProperties: serde.Properties{}},
	}

	decision := adminconfig.Decision{Deploy: adminconfig.StatusFalse}
	decisionMap := adminconfig.DecisionPerCapabilityMap{"copy": decision}
	attrOptimization := adminconfig.AttributeOptimization{Attribute: "ClusterCost", Weight: "1.0", Directive: adminconfig.Minimize}
	optimizationStrategy := []adminconfig.AttributeOptimization{attrOptimization}
	evalOutput := adminconfig.EvaluatorOutput{ConfigDecisions: decisionMap, OptimizationStrategy: optimizationStrategy}
	appRequirements := appApi.DataRequirements{Interface: env.Modules["Reader"].Spec.Capabilities[0].SupportedInterfaces[1].Sink}
	appContext := appApi.DataContext{Requirements: appRequirements}
	dataSetConnection := taxonomy.Connection{Name: "s3"}
	resourceDetails := datacatalog.ResourceDetails{Connection: dataSetConnection, DataFormat: "parquet"}
	dataDetails := datacatalog.GetAssetResponse{Details: resourceDetails}
	dataInfo := datapath.DataInfo{
		DataDetails:         &dataDetails,
		Context:             &appContext,
		Configuration:       evalOutput,
		WorkloadCluster:     env.Clusters[0],
		Actions:             actions,
		StorageRequirements: map[taxonomy.ProcessingLocation][]taxonomy.Action{"us-south": {}},
	}
	return &dataInfo
}

func TestBuildModel(t *testing.T) {
	env := getTestEnv()
	dataInfo := getDataInfo(env)
	dpCSP := NewDataPathCSP(dataInfo, env)
	modelFile, err := dpCSP.BuildFzModel(3)
	if modelFile != "" {
		os.Remove(modelFile)
	}
	if err != nil {
		t.Errorf("Failed building a CSP model: %s", err)
	}
}

func TestRequiredCapability(t *testing.T) {
	env := getTestEnv()
	dataInfo := getDataInfo(env)
	dataInfo.Configuration.ConfigDecisions["read"] = adminconfig.Decision{Deploy: adminconfig.StatusTrue}
	dpCSP := NewDataPathCSP(dataInfo, env)
	modelFile, err := dpCSP.BuildFzModel(3)
	if modelFile != "" {
		os.Remove(modelFile)
	}
	if err != nil {
		t.Errorf("Failed building a CSP model: %s", err)
	}
}

func TestRequiredCapabilityMissing(t *testing.T) {
	env := getTestEnv()
	dataInfo := getDataInfo(env)
	dataInfo.Configuration.ConfigDecisions["transform"] = adminconfig.Decision{Deploy: adminconfig.StatusTrue}
	dpCSP := NewDataPathCSP(dataInfo, env)
	modelFile, err := dpCSP.BuildFzModel(3)
	if modelFile != "" {
		os.Remove(modelFile)
	}
	if err == nil {
		t.Error("This test should result in an error - no module has the required capability")
	}
}
