// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"testing"

	"github.com/onsi/gomega"
	"github.com/rs/zerolog"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/manager/controllers/mockup"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/datapath"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/infrastructure"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
)

var testLog = logging.LogInit("Solver", "Test")

func newEnvironment() *datapath.Environment {
	return &datapath.Environment{
		Clusters:        []multicluster.Cluster{},
		Modules:         map[string]*fapp.FybrikModule{},
		StorageAccounts: []*fapp.FybrikStorageAccount{},
		AttributeManager: &infrastructure.AttributeManager{
			Log:        testLog,
			Metrics:    infrastructure.MetricsDictionary{},
			Attributes: []taxonomy.InfrastructureElement{},
		},
	}
}

func addCluster(env *datapath.Environment, cluster multicluster.Cluster) {
	env.Clusters = append(env.Clusters, cluster)
}

func addModule(env *datapath.Environment, module *fapp.FybrikModule) {
	env.Modules[module.Name] = module
}

func addStorageAccount(env *datapath.Environment, account *fapp.FybrikStorageAccount) {
	env.StorageAccounts = append(env.StorageAccounts, account)
}

func addMetrics(env *datapath.Environment, m *taxonomy.InfrastructureMetrics) {
	env.AttributeManager.Metrics[m.Name] = *m
}

func addAttribute(env *datapath.Environment, attribute *taxonomy.InfrastructureElement) {
	env.AttributeManager.Attributes = append(env.AttributeManager.Attributes, *attribute)
}

// default: S3, csv
func createReadRequest() *datapath.DataInfo {
	return &datapath.DataInfo{
		DataDetails: &datacatalog.GetAssetResponse{Details: datacatalog.ResourceDetails{
			Connection: taxonomy.Connection{Name: mockup.S3},
			DataFormat: mockup.CSV,
		}},
		Actions:             []taxonomy.Action{},
		StorageRequirements: make(map[taxonomy.ProcessingLocation][]taxonomy.Action),
		Context: &fapp.DataContext{
			DataSetID: "id",
			Flow:      taxonomy.ReadFlow,
			Requirements: fapp.DataRequirements{
				Interface:  &taxonomy.Interface{Protocol: mockup.ArrowFlight},
				FlowParams: fapp.FlowRequirements{},
			},
		},
		Configuration: adminconfig.EvaluatorOutput{
			Valid: true,
			ConfigDecisions: adminconfig.DecisionPerCapabilityMap{
				"read":   adminconfig.Decision{Deploy: adminconfig.StatusTrue},
				"write":  adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"delete": adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"copy":   adminconfig.Decision{Deploy: adminconfig.StatusUnknown},
			},
			OptimizationStrategy: []adminconfig.AttributeOptimization{},
		},
	}
}

// copy flow s3,csv -> s3,csv
func createCopyRequest() *datapath.DataInfo {
	return &datapath.DataInfo{
		DataDetails: &datacatalog.GetAssetResponse{Details: datacatalog.ResourceDetails{
			Connection: taxonomy.Connection{Name: mockup.S3},
			DataFormat: mockup.CSV,
		}},
		Actions:             []taxonomy.Action{},
		StorageRequirements: make(map[taxonomy.ProcessingLocation][]taxonomy.Action),
		Context: &fapp.DataContext{
			DataSetID: "ingest",
			Flow:      taxonomy.CopyFlow,
			Requirements: fapp.DataRequirements{
				Interface:  &taxonomy.Interface{Protocol: mockup.S3, DataFormat: mockup.CSV},
				FlowParams: fapp.FlowRequirements{},
			},
		},
		Configuration: adminconfig.EvaluatorOutput{
			Valid: true,
			ConfigDecisions: adminconfig.DecisionPerCapabilityMap{
				"read":   adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"write":  adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"delete": adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"copy":   adminconfig.Decision{Deploy: adminconfig.StatusTrue},
			},
		},
	}
}

func createWriteNewAssetRequest() *datapath.DataInfo {
	return &datapath.DataInfo{
		Actions:             []taxonomy.Action{},
		StorageRequirements: make(map[taxonomy.ProcessingLocation][]taxonomy.Action),
		Context: &fapp.DataContext{
			DataSetID: "newAsset",
			Flow:      taxonomy.WriteFlow,
			Requirements: fapp.DataRequirements{
				Interface:  &taxonomy.Interface{Protocol: mockup.ArrowFlight},
				FlowParams: fapp.FlowRequirements{IsNewDataSet: true},
			},
		},
		Configuration: adminconfig.EvaluatorOutput{
			Valid: true,
			ConfigDecisions: adminconfig.DecisionPerCapabilityMap{
				"read":   adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"copy":   adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"delete": adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"write":  adminconfig.Decision{Deploy: adminconfig.StatusTrue},
			},
		},
	}
}

func createUpdateRequest() *datapath.DataInfo {
	return &datapath.DataInfo{
		DataDetails: &datacatalog.GetAssetResponse{Details: datacatalog.ResourceDetails{
			Connection: taxonomy.Connection{Name: mockup.S3},
			DataFormat: mockup.CSV,
		}},
		Actions:             []taxonomy.Action{},
		StorageRequirements: make(map[taxonomy.ProcessingLocation][]taxonomy.Action),
		Context: &fapp.DataContext{
			DataSetID: "write",
			Flow:      taxonomy.WriteFlow,
			Requirements: fapp.DataRequirements{
				Interface:  &taxonomy.Interface{Protocol: mockup.ArrowFlight},
				FlowParams: fapp.FlowRequirements{},
			},
		},
		Configuration: adminconfig.EvaluatorOutput{
			Valid: true,
			ConfigDecisions: adminconfig.DecisionPerCapabilityMap{
				"read":   adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"copy":   adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"delete": adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"write":  adminconfig.Decision{Deploy: adminconfig.StatusTrue},
			},
		},
	}
}

func createDeleteRequest() *datapath.DataInfo {
	return &datapath.DataInfo{
		DataDetails: &datacatalog.GetAssetResponse{Details: datacatalog.ResourceDetails{
			Connection: taxonomy.Connection{Name: mockup.S3},
			DataFormat: mockup.CSV,
		}},
		Actions:             []taxonomy.Action{},
		StorageRequirements: make(map[taxonomy.ProcessingLocation][]taxonomy.Action),
		Context: &fapp.DataContext{
			DataSetID: "delete",
			Flow:      taxonomy.DeleteFlow,
		},
		Configuration: adminconfig.EvaluatorOutput{
			Valid: true,
			ConfigDecisions: adminconfig.DecisionPerCapabilityMap{
				"read":   adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"copy":   adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"write":  adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"delete": adminconfig.Decision{Deploy: adminconfig.StatusTrue},
			},
		},
	}
}

// no clusters/modules - data path can't be constructed
func TestEmptyEnvironment(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	_, err := solveSingleDataset(env, createReadRequest(), &testLog)
	g.Expect(err).To(gomega.HaveOccurred())
}

// transformations are required but not supported by the read module
// copy will be selected as well as read
func TestReadWithTransforms(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	copyModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	account := &fapp.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: string(account.Spec.Region)}})
	asset := createReadRequest()
	asset.Actions = []taxonomy.Action{{Name: "RedactAction"}}
	_, err := solveSingleDataset(env, asset, &testLog)
	// only read is not enough
	g.Expect(err).To(gomega.HaveOccurred())
	addModule(env, copyModule)
	addStorageAccount(env, account)
	asset.StorageRequirements[account.Spec.Region] = []taxonomy.Action{}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(2))
}

// check that a module has the appropriate source interface
func TestReadModuleSource(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModuleS3 := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModuleS3)).NotTo(gomega.HaveOccurred())
	addModule(env, readModuleS3)
	readModuleDB2 := readModuleS3.DeepCopy()
	readModuleDB2.Name = "readDB2"
	readModuleDB2.Spec.Capabilities[0].SupportedInterfaces[0] = fapp.ModuleInOut{Source: &taxonomy.Interface{Protocol: mockup.JdbcDB2}}
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: "xyz"}})
	asset := createReadRequest()
	asset.DataDetails.Details.Connection.Name = mockup.JdbcDB2
	asset.DataDetails.Details.DataFormat = ""
	_, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).To(gomega.HaveOccurred())
	addModule(env, readModuleDB2)
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	logging.LogStructure("TestReadModuleSource", &solution, &testLog, zerolog.InfoLevel, false, false)
	g.Expect(solution.DataPath).To(gomega.HaveLen(1))
	g.Expect(solution.DataPath[0].Module.Name).To(gomega.Equal(readModuleDB2.Name))
}

// read + copy with transforms
func TestReadAndCopyWithTransforms(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	copyModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/copy-db2-parquet-no-transforms.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	account := &fapp.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: string(account.Spec.Region)}})
	asset := createReadRequest()
	asset.DataDetails.Details.Connection.Name = mockup.JdbcDB2
	asset.DataDetails.Details.DataFormat = ""
	addModule(env, copyModule)
	addStorageAccount(env, account)
	asset.StorageRequirements[account.Spec.Region] = []taxonomy.Action{}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(2))
	asset.Actions = []taxonomy.Action{{Name: "RedactAction"}}
	_, err = solveSingleDataset(env, asset, &testLog)
	g.Expect(err).To(gomega.HaveOccurred())
}

// read + transform
func TestReadAndTransformModules(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	transformModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-transform.yaml", transformModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addModule(env, transformModule)
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: "xyz"}})
	asset := createReadRequest()
	asset.DataDetails.Details.Connection.Name = mockup.S3
	asset.DataDetails.Details.DataFormat = mockup.Parquet
	asset.Actions = []taxonomy.Action{{Name: "RedactAction"}}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(2))
	g.Expect(solution.DataPath[0].Module.Name).To(gomega.Equal(readModule.Name))
	g.Expect(solution.DataPath[1].Module.Name).To(gomega.Equal(transformModule.Name))
	g.Expect(solution.DataPath[0].Actions).To(gomega.BeEmpty())
	g.Expect(solution.DataPath[1].Actions).To(gomega.HaveLen(1))
}

// Chaining two read modules when transformations are required
func TestReadAfterRead(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	transformModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-transform.yaml", transformModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	transformModule.Spec.Capabilities[0].Capability = "read"
	addModule(env, transformModule)
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: "xyz"}})
	asset := createReadRequest()
	asset.DataDetails.Details.Connection.Name = mockup.S3
	asset.DataDetails.Details.DataFormat = mockup.Parquet
	asset.Actions = []taxonomy.Action{{Name: "RedactAction"}}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(2))
	g.Expect(solution.DataPath[0].Module.Name).To(gomega.Equal(readModule.Name))
	g.Expect(solution.DataPath[1].Module.Name).To(gomega.Equal(transformModule.Name))
	g.Expect(solution.DataPath[0].Actions).To(gomega.BeEmpty())
	g.Expect(solution.DataPath[1].Actions).To(gomega.HaveLen(1))
}

// Transform close to the data
// The locations of the workload and the requested dataset are different
func TestTransformInDataLocation(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	copyModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/copy-csv-parquet.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addModule(env, copyModule)
	account := &fapp.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account)
	remoteGeo := "remote"
	cluster1 := multicluster.Cluster{Name: "c1", Metadata: multicluster.ClusterMetadata{Region: string(account.Spec.Region)}}
	cluster2 := multicluster.Cluster{Name: "c2", Metadata: multicluster.ClusterMetadata{Region: remoteGeo}}
	addCluster(env, cluster1)
	addCluster(env, cluster2)
	asset := createReadRequest()
	asset.DataDetails.ResourceMetadata.Geography = remoteGeo
	asset.WorkloadCluster = cluster1
	asset.Actions = []taxonomy.Action{{Name: "RedactAction"}}
	asset.Configuration.ConfigDecisions["copy"] = adminconfig.Decision{Deploy: adminconfig.StatusFalse}
	asset.Configuration.ConfigDecisions["read"] = adminconfig.Decision{
		Deploy: adminconfig.StatusTrue,
		DeploymentRestrictions: adminconfig.Restrictions{
			Clusters: []adminconfig.Restriction{{Property: "metadata.region", Values: adminconfig.StringList{string(account.Spec.Region)}}}},
	}
	asset.Configuration.ConfigDecisions[Transform] = adminconfig.Decision{
		Deploy: adminconfig.StatusUnknown,
		DeploymentRestrictions: adminconfig.Restrictions{
			Clusters: []adminconfig.Restriction{{Property: "metadata.region", Values: adminconfig.StringList{remoteGeo}}}},
	}
	asset.StorageRequirements[account.Spec.Region] = []taxonomy.Action{}
	asset.StorageRequirements[taxonomy.ProcessingLocation(remoteGeo)] = []taxonomy.Action{}
	_, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).To(gomega.HaveOccurred())
	// remove restriction on copy
	asset.Configuration.ConfigDecisions["copy"] = adminconfig.Decision{Deploy: adminconfig.StatusUnknown}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(2))
	// copy
	g.Expect(solution.DataPath[0].StorageAccount.Region).To(gomega.Equal(account.Spec.Region))
	g.Expect(solution.DataPath[0].Cluster).To(gomega.Equal(cluster2.Name))
	// read
	g.Expect(solution.DataPath[1].Cluster).To(gomega.Equal(cluster1.Name))
}

// This test checks the copy scenario
// Two storage accounts are created. Data cannot be stored in one of them according to governance policies.
func TestCopyFlow(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	copyModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addModule(env, copyModule)
	account1 := &fapp.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-neverland.yaml", account1)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account1)
	account2 := &fapp.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account2)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account2)
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: string(account2.Spec.Region)}})
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: string(account1.Spec.Region)}})

	asset := createCopyRequest()
	_, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).To(gomega.HaveOccurred())
	asset.StorageRequirements[account2.Spec.Region] = []taxonomy.Action{}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(1))
	// copy
	g.Expect(solution.DataPath[0].StorageAccount.Region).To(gomega.Equal(account2.Spec.Region))
	g.Expect(solution.DataPath[0].Module.Name).To(gomega.Equal(copyModule.Name))
}

// restrictions on a storage account attribute
func TestStorageCostRestrictictions(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	copyModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	addModule(env, copyModule)
	account1 := &fapp.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-neverland.yaml", account1)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account1)
	account2 := &fapp.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account2)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account2)
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: string(account1.Spec.Region)}})
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: string(account2.Spec.Region)}})

	asset := createCopyRequest()
	asset.StorageRequirements[account1.Spec.Region] = []taxonomy.Action{}
	asset.StorageRequirements[account2.Spec.Region] = []taxonomy.Action{}
	asset.Configuration.ConfigDecisions["copy"] = adminconfig.Decision{
		Deploy: adminconfig.StatusTrue,
		DeploymentRestrictions: adminconfig.Restrictions{
			StorageAccounts: []adminconfig.Restriction{{Property: "storage-cost", Range: &taxonomy.RangeType{Max: 10}}}},
	}
	addMetrics(env, &taxonomy.InfrastructureMetrics{
		Name:  "cost",
		Type:  taxonomy.Numeric,
		Scale: &taxonomy.RangeType{Min: 0, Max: 200},
	})
	addAttribute(env, &taxonomy.InfrastructureElement{
		Name:       "storage-cost",
		MetricName: "cost",
		Value:      "20",
		Object:     taxonomy.StorageAccount,
		Instance:   account1.Name,
	})
	addAttribute(env, &taxonomy.InfrastructureElement{
		Name:       "storage-cost",
		MetricName: "cost",
		Value:      "12",
		Object:     taxonomy.StorageAccount,
		Instance:   account2.Name,
	})
	_, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).To(gomega.HaveOccurred())
	// change the restriction to fit one of the accounts
	asset.Configuration.ConfigDecisions["copy"] = adminconfig.Decision{
		Deploy: adminconfig.StatusTrue,
		DeploymentRestrictions: adminconfig.Restrictions{
			StorageAccounts: []adminconfig.Restriction{{Property: "storage-cost", Range: &taxonomy.RangeType{Max: 15}}}},
	}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(1))
	g.Expect(solution.DataPath[0].StorageAccount.Region).To(gomega.Equal(account2.Spec.Region))
}

// This test checks the write scenario
// Asset is not registered in the catalog
func TestWriteNewAsset(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	writeModule := &fapp.FybrikModule{}
	copyModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-write.yaml", writeModule)).NotTo(gomega.HaveOccurred())
	addModule(env, writeModule)
	addModule(env, copyModule)
	account1 := &fapp.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-neverland.yaml", account1)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account1)
	account2 := &fapp.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account2)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account2)
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: string(account2.Spec.Region)}})
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: string(account1.Spec.Region)}})
	asset := createWriteNewAssetRequest()
	asset.StorageRequirements[account1.Spec.Region] = []taxonomy.Action{}
	asset.StorageRequirements[account2.Spec.Region] = []taxonomy.Action{}
	asset.Configuration.ConfigDecisions["write"] = adminconfig.Decision{
		Deploy: adminconfig.StatusTrue,
		DeploymentRestrictions: adminconfig.Restrictions{
			StorageAccounts: []adminconfig.Restriction{{Property: "region", Values: adminconfig.StringList{string(account2.Spec.Region)}}}},
	}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(1))
	// write
	g.Expect(solution.DataPath[0].StorageAccount.Region).To(gomega.Equal(account2.Spec.Region))
	g.Expect(solution.DataPath[0].Module.Name).To(gomega.Equal(writeModule.Name))
}

// This test checks the write scenario
// Asset exists, no storage is required
func TestWriteExistingAsset(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	writeModule := &fapp.FybrikModule{}
	copyModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-write.yaml", writeModule)).NotTo(gomega.HaveOccurred())
	addModule(env, writeModule)
	addModule(env, copyModule)
	account1 := &fapp.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-neverland.yaml", account1)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account1)
	account2 := &fapp.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account2)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account2)
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: string(account2.Spec.Region)}})
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: string(account1.Spec.Region)}})
	asset := createUpdateRequest()
	asset.StorageRequirements[account1.Spec.Region] = []taxonomy.Action{}
	asset.StorageRequirements[account2.Spec.Region] = []taxonomy.Action{}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(1))
	// write
	g.Expect(solution.DataPath[0].StorageAccount.Region).To(gomega.BeEmpty())
	g.Expect(solution.DataPath[0].Module.Name).To(gomega.Equal(writeModule.Name))
}

// write + transform
func TestWriteAndTransformModules(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	writeModule := &fapp.FybrikModule{}
	transformModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-transform.yaml", transformModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-write.yaml", writeModule)).NotTo(gomega.HaveOccurred())
	addModule(env, writeModule)
	addModule(env, transformModule)
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: "xyz"}})
	asset := createUpdateRequest()
	asset.Actions = []taxonomy.Action{{Name: "RedactAction"}}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(2))
	g.Expect(solution.DataPath[0].Module.Name).To(gomega.Equal(writeModule.Name))
	g.Expect(solution.DataPath[1].Module.Name).To(gomega.Equal(transformModule.Name))
	g.Expect(solution.DataPath[0].Actions).To(gomega.BeEmpty())
	g.Expect(solution.DataPath[1].Actions).To(gomega.HaveLen(1))
}

// delete flow
func TestDeleteFlow(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	deleteModule := &fapp.FybrikModule{}
	writeModule := &fapp.FybrikModule{}
	transformModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-transform.yaml", transformModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-write.yaml", writeModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-delete.yaml", deleteModule)).NotTo(gomega.HaveOccurred())
	addModule(env, writeModule)
	addModule(env, deleteModule)
	addModule(env, transformModule)
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: "xyz"}})
	asset := createDeleteRequest()
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(1))
	g.Expect(solution.DataPath[0].Module.Name).To(gomega.Equal(deleteModule.Name))
}

// check restrictions on a module
func TestModuleSelection(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	workloadLevelModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", workloadLevelModule)).NotTo(gomega.HaveOccurred())
	assetLevelModule := workloadLevelModule.DeepCopy()
	assetLevelModule.Spec.Capabilities[0].Scope = fapp.Asset
	assetLevelModule.Name = "assetLevel"
	workloadLevelModule.Name = "workloadLevel"
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: "xyz"}})
	addModule(env, assetLevelModule)
	asset := createReadRequest()
	asset.Configuration.ConfigDecisions["read"] = adminconfig.Decision{
		Deploy: adminconfig.StatusTrue,
		DeploymentRestrictions: adminconfig.Restrictions{Modules: []adminconfig.Restriction{{
			Property: "capabilities.scope",
			Values:   adminconfig.StringList{"workload"}}}}}
	_, err := solveSingleDataset(env, asset, &testLog)
	// wrong scope
	g.Expect(err).To(gomega.HaveOccurred())
	addModule(env, workloadLevelModule)
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(1))
	g.Expect(solution.DataPath[0].Module.Name).To(gomega.Equal(workloadLevelModule.Name))
}

// a read scenario
// copy and read modules are deployed
// transformations are required but not supported by the read module
// 5 storage accounts exist: one is not allowed by governance, another needs a non-supported action
// optimization goal is to select the cheapest storage
func TestOptimalStorage(t *testing.T) {
	t.Parallel()
	if !environment.UseCSP() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	copyModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addModule(env, copyModule)
	clusterRegion := "theshire"
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: clusterRegion}})
	asset := createReadRequest()
	asset.Actions = []taxonomy.Action{{Name: "RedactAction"}}
	asset.Configuration.OptimizationStrategy = []adminconfig.AttributeOptimization{{
		Attribute: "storage-cost",
		Directive: adminconfig.Minimize,
		Weight:    "1.0",
	}}
	addMetrics(env, &taxonomy.InfrastructureMetrics{Name: "cost", Type: taxonomy.Numeric, Scale: &taxonomy.RangeType{Max: 200}})
	cost := 50
	for i := 0; i < 5; i++ {
		account := &fapp.FybrikStorageAccount{
			Spec: fapp.FybrikStorageAccountSpec{
				ID:        genName("account-", i),
				SecretRef: genName("credentials-", i),
				Region:    taxonomy.ProcessingLocation(genName("region", i)),
				Endpoint:  "dummy-endpoint",
			}}
		account.Name = account.Spec.ID
		addStorageAccount(env, account)
		if i == 1 {
			asset.StorageRequirements[account.Spec.Region] = []taxonomy.Action{{Name: "AgeFilterAction"}}
		} else if i >= 2 {
			asset.StorageRequirements[account.Spec.Region] = []taxonomy.Action{}
		}
		addAttribute(env, &taxonomy.InfrastructureElement{
			Name:       "storage-cost",
			MetricName: "cost",
			Value:      fmt.Sprintf("%d", cost),
			Object:     taxonomy.StorageAccount,
			Instance:   account.Name,
		})
		cost += 5
	}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(2))
	g.Expect(solution.DataPath[0].StorageAccount.Region).To(gomega.Equal(taxonomy.ProcessingLocation("region2")))
}

// a read scenario
// copy and read modules are deployed
// transformations are required but not supported by the read module
// 5 storage accounts exist: one is not allowed by governance, another needs a non-supported action
// 2 clusters exist: one is cheap and the other is expensive
// optimization goal is to minimize the cost of both storage accounts and clusters
func TestOptimalStorageAndClusterCost(t *testing.T) {
	t.Parallel()
	if !environment.UseCSP() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	copyModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addModule(env, copyModule)
	clusterRegion := "theshire"
	addCluster(env, multicluster.Cluster{Name: "Cheap", Metadata: multicluster.ClusterMetadata{Region: clusterRegion}})
	addCluster(env, multicluster.Cluster{Name: "Expensive", Metadata: multicluster.ClusterMetadata{Region: clusterRegion}})
	asset := createReadRequest()
	asset.Actions = []taxonomy.Action{{Name: "RedactAction"}}
	asset.Configuration.OptimizationStrategy = []adminconfig.AttributeOptimization{{
		Attribute: "cost",
		Directive: adminconfig.Minimize,
		Weight:    "1.0",
	}}
	addMetrics(env, &taxonomy.InfrastructureMetrics{Name: "cost", Type: taxonomy.Numeric, Scale: &taxonomy.RangeType{Max: 200}})
	cost := 50
	for i := 0; i < 5; i++ {
		account := &fapp.FybrikStorageAccount{
			Spec: fapp.FybrikStorageAccountSpec{
				ID:        genName("account-", i),
				SecretRef: genName("credentials-", i),
				Region:    taxonomy.ProcessingLocation(genName("region", i)),
				Endpoint:  "dummy-endpoint",
			}}
		account.Name = account.Spec.ID
		addStorageAccount(env, account)
		if i == 1 {
			asset.StorageRequirements[account.Spec.Region] = []taxonomy.Action{{Name: "AgeFilterAction"}}
		} else if i >= 2 {
			asset.StorageRequirements[account.Spec.Region] = []taxonomy.Action{}
		}
		addAttribute(env, &taxonomy.InfrastructureElement{
			Name:       "cost",
			MetricName: "cost",
			Value:      fmt.Sprintf("%d", cost),
			Object:     taxonomy.StorageAccount,
			Instance:   account.Name,
		})
		cost += 5
	}
	addAttribute(env, &taxonomy.InfrastructureElement{
		Name:       "cost",
		MetricName: "cost",
		Value:      fmt.Sprintf("%d", 50),
		Object:     taxonomy.Cluster,
		Instance:   "Cheap",
	})
	addAttribute(env, &taxonomy.InfrastructureElement{
		Name:       "cost",
		MetricName: "cost",
		Value:      fmt.Sprintf("%d", 200),
		Object:     taxonomy.Cluster,
		Instance:   "Expensive",
	})
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(2))
	g.Expect(solution.DataPath[0].StorageAccount.Region).To(gomega.Equal(taxonomy.ProcessingLocation("region2")))
	g.Expect(solution.DataPath[0].Cluster).To(gomega.Equal("Cheap"))
	g.Expect(solution.DataPath[1].Cluster).To(gomega.Equal("Cheap"))
}

func genName(prefix string, ind int) string {
	return fmt.Sprintf("%s%d", prefix, ind)
}

// Conflicting optimization goals
// Read scenario, different clusters with costs
// Conflicting goals: minimize and maximize cluster costs
// Result: the module is deployed somewhere
func TestGoalConflict(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addMetrics(env, &taxonomy.InfrastructureMetrics{Name: "cost", Type: taxonomy.Numeric, Scale: &taxonomy.RangeType{Max: 200}})
	cost := 10
	for i := 0; i < 5; i++ {
		name := genName("cluster", i)
		addCluster(env, multicluster.Cluster{Name: name, Metadata: multicluster.ClusterMetadata{Region: genName("region", i)}})
		addAttribute(env, &taxonomy.InfrastructureElement{
			Name:       "cluster-cost",
			MetricName: "cost",
			Value:      fmt.Sprintf("%d", cost),
			Object:     taxonomy.Cluster,
			Instance:   name,
		})
		cost -= 1
	}
	asset := createReadRequest()
	asset.Configuration.OptimizationStrategy = []adminconfig.AttributeOptimization{
		{
			Attribute: "cluster-cost",
			Directive: adminconfig.Minimize,
			Weight:    "0.2",
		},
		{
			Attribute: "cluster-cost",
			Directive: adminconfig.Maximize,
			Weight:    "0.8",
		},
	}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(1))
	g.Expect(solution.DataPath[0].Cluster).To(gomega.HavePrefix("cluster"))
}

// Read scenario, different clusters with costs
// Two minimize goals with different weights: 9:1
// Costs: (10,0), (9,0), (8,10), (7,20), (6,30)
// The second cluster should be selected
func TestMinMultipleGoals(t *testing.T) {
	t.Parallel()
	if !environment.UseCSP() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addMetrics(env, &taxonomy.InfrastructureMetrics{Name: "rate", Type: taxonomy.Numeric, Scale: &taxonomy.RangeType{Max: 100}})
	cpuCost := 10
	errRate := 0
	for i := 1; i <= 5; i++ {
		name := genName("cluster", i)
		addCluster(env, multicluster.Cluster{Name: name, Metadata: multicluster.ClusterMetadata{Region: genName("region", i)}})
		addAttribute(env, &taxonomy.InfrastructureElement{
			Name:       "cluster-cpu-cost",
			MetricName: "rate",
			Value:      fmt.Sprintf("%d", cpuCost),
			Object:     taxonomy.Cluster,
			Instance:   name,
		})
		addAttribute(env, &taxonomy.InfrastructureElement{
			Name:       "cluster-err-rate",
			MetricName: "rate",
			Value:      fmt.Sprintf("%d", errRate),
			Object:     taxonomy.Cluster,
			Instance:   name,
		})
		cpuCost -= 1
		if i >= 2 {
			errRate += 10
		}
	}
	asset := createReadRequest()
	asset.Configuration.OptimizationStrategy = []adminconfig.AttributeOptimization{
		{
			Attribute: "cluster-cpu-cost",
			Directive: adminconfig.Minimize,
			Weight:    "0.9",
		},
		{
			Attribute: "cluster-err-rate",
			Directive: adminconfig.Minimize,
			Weight:    "0.1",
		},
	}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(1))
	g.Expect(solution.DataPath[0].Cluster).To(gomega.Equal("cluster2"))
}

// Read scenario, different clusters with costs
// Min & max goals with different weights: 6:4
// Costs: (10,0), (4,0), (9,5), (3,5), (8,5)
// cluster4 should be selected
func TestMinMaxGoals(t *testing.T) {
	t.Parallel()
	if !environment.UseCSP() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addMetrics(env, &taxonomy.InfrastructureMetrics{Name: "rate", Type: taxonomy.Numeric, Scale: &taxonomy.RangeType{Max: 100}})
	cpuCost := 10
	stableRate := 0
	for i := 1; i <= 5; i++ {
		name := genName("cluster", i)
		addCluster(env, multicluster.Cluster{Name: name, Metadata: multicluster.ClusterMetadata{Region: genName("region", i)}})
		addAttribute(env, &taxonomy.InfrastructureElement{
			Name:       "cluster-cpu-cost",
			MetricName: "rate",
			Value:      fmt.Sprintf("%d", cpuCost),
			Object:     taxonomy.Cluster,
			Instance:   name,
		})
		addAttribute(env, &taxonomy.InfrastructureElement{
			Name:       "cluster-stability-rate",
			MetricName: "rate",
			Value:      fmt.Sprintf("%d", stableRate),
			Object:     taxonomy.Cluster,
			Instance:   name,
		})
		cpuCost = 15 - cpuCost - i
		if i == 2 {
			stableRate += 5
		}
	}
	asset := createReadRequest()
	asset.Configuration.OptimizationStrategy = []adminconfig.AttributeOptimization{
		{
			Attribute: "cluster-cpu-cost",
			Directive: adminconfig.Minimize,
			Weight:    "0.6",
		},
		{
			Attribute: "cluster-stability-rate",
			Directive: adminconfig.Maximize,
			Weight:    "0.4",
		},
	}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(1))
	g.Expect(solution.DataPath[0].Cluster).To(gomega.Equal("cluster4"))
}

// a read scenario, data is in a remote location
// copy and read modules are deployed
// optimization goal is to minimize the distance
// copy is expected to be deployed
func TestMinDistance(t *testing.T) {
	t.Parallel()
	if !environment.UseCSP() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	copyModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addModule(env, copyModule)
	workloadCluster := "theshire"
	remoteCluster := "neverland"
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: workloadCluster}})
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: remoteCluster}})
	asset := createReadRequest()
	asset.Configuration.OptimizationStrategy = []adminconfig.AttributeOptimization{{
		Attribute: "distance",
		Directive: adminconfig.Minimize,
		Weight:    "1.0",
	}}
	addMetrics(env, &taxonomy.InfrastructureMetrics{Name: "distance", Type: taxonomy.Numeric, Scale: &taxonomy.RangeType{Max: 20000}})
	account := &fapp.FybrikStorageAccount{
		Spec: fapp.FybrikStorageAccountSpec{
			ID:        "account-theshire",
			SecretRef: "credentials-theshire",
			Region:    taxonomy.ProcessingLocation(workloadCluster),
			Endpoint:  "dummy-endpoint",
		}}
	account.Name = account.Spec.ID
	addStorageAccount(env, account)
	asset.StorageRequirements[account.Spec.Region] = []taxonomy.Action{}
	asset.WorkloadCluster = multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: workloadCluster}}
	asset.DataDetails.ResourceMetadata.Geography = remoteCluster
	addAttribute(env, &taxonomy.InfrastructureElement{
		Name:       "distance",
		MetricName: "distance",
		Value:      "2000",
		Object:     taxonomy.InterRegion,
		Arguments:  []string{"theshire", "neverland"},
	})
	addAttribute(env, &taxonomy.InfrastructureElement{
		Name:       "distance",
		MetricName: "distance",
		Value:      "0",
		Object:     taxonomy.InterRegion,
		Arguments:  []string{"neverland", "neverland"},
	})
	addAttribute(env, &taxonomy.InfrastructureElement{
		Name:       "distance",
		MetricName: "distance",
		Value:      "0",
		Object:     taxonomy.InterRegion,
		Arguments:  []string{"theshire", "theshire"},
	})
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(2))
	g.Expect(solution.DataPath[0].StorageAccount.Region).To(gomega.Equal(taxonomy.ProcessingLocation(workloadCluster)))
}

// a read scenario
// a read module is deployed
// two clusters, only one is allowed because of a compliance attribute
func TestClusterAttribute(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &fapp.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	allowedCluster := multicluster.Cluster{Name: "clusterA", Metadata: multicluster.ClusterMetadata{Region: "regionA"}}
	forbiddenCluster := multicluster.Cluster{Name: "clusterB", Metadata: multicluster.ClusterMetadata{Region: "regionB"}}
	addCluster(env, allowedCluster)
	addCluster(env, forbiddenCluster)
	asset := createReadRequest()
	asset.WorkloadCluster = forbiddenCluster
	asset.DataDetails.ResourceMetadata.Geography = "regionC"
	addMetrics(env, &taxonomy.InfrastructureMetrics{Name: "compliance", Type: taxonomy.Bool})
	addAttribute(env, &taxonomy.InfrastructureElement{
		Name:       "compliance",
		MetricName: "compliance",
		Value:      "true",
		Object:     taxonomy.Cluster,
		Instance:   allowedCluster.Name,
	})
	addAttribute(env, &taxonomy.InfrastructureElement{
		Name:       "compliance",
		MetricName: "compliance",
		Value:      "false",
		Object:     taxonomy.Cluster,
		Instance:   forbiddenCluster.Name,
	})
	asset.Configuration.ConfigDecisions["read"] = adminconfig.Decision{
		Deploy: adminconfig.StatusTrue,
		DeploymentRestrictions: adminconfig.Restrictions{
			Clusters: []adminconfig.Restriction{{Property: "compliance", Values: adminconfig.StringList{{"true"}}}},
	}
	solution, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath[0].Cluster).To(gomega.Equal(allowedCluster.Name))
}
