// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"testing"

	"github.com/onsi/gomega"
	"github.com/rs/zerolog"

	"fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/datapath"
	"fybrik.io/fybrik/pkg/infrastructure"
	"fybrik.io/fybrik/pkg/logging"
	infraattributes "fybrik.io/fybrik/pkg/model/attributes"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
)

var testLog = logging.LogInit("Solver", "Test")

func newEnvironment() *datapath.Environment {
	return &datapath.Environment{
		Clusters:        []multicluster.Cluster{},
		Modules:         map[string]*v1alpha1.FybrikModule{},
		StorageAccounts: []*v1alpha1.FybrikStorageAccount{},
		AttributeManager: &infrastructure.AttributeManager{
			Log:            testLog,
			Infrastructure: infraattributes.Infrastructure{},
		},
	}
}

func addCluster(env *datapath.Environment, cluster multicluster.Cluster) {
	env.Clusters = append(env.Clusters, cluster)
}

func addModule(env *datapath.Environment, module *v1alpha1.FybrikModule) {
	env.Modules[module.Name] = module
}

func addStorageAccount(env *datapath.Environment, account *v1alpha1.FybrikStorageAccount) {
	env.StorageAccounts = append(env.StorageAccounts, account)
}

func addAttribute(env *datapath.Environment, attribute *taxonomy.InfrastructureElement) {
	env.AttributeManager.Infrastructure.Items = append(env.AttributeManager.Infrastructure.Items, *attribute)
}

// default: S3, csv
func createReadRequest() *datapath.DataInfo {
	return &datapath.DataInfo{
		DataDetails: &datacatalog.GetAssetResponse{Details: datacatalog.ResourceDetails{
			Connection: taxonomy.Connection{Name: v1alpha1.S3},
			DataFormat: v1alpha1.CSV,
		}},
		Actions:             []taxonomy.Action{},
		StorageRequirements: make(map[taxonomy.ProcessingLocation][]taxonomy.Action),
		Context: &v1alpha1.DataContext{
			DataSetID: "id",
			Flow:      taxonomy.ReadFlow,
			Requirements: v1alpha1.DataRequirements{
				Interface:  &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight},
				FlowParams: v1alpha1.FlowRequirements{},
			},
		},
		Configuration: adminconfig.EvaluatorOutput{
			Valid: true,
			ConfigDecisions: adminconfig.DecisionPerCapabilityMap{
				"read":   adminconfig.Decision{Deploy: adminconfig.StatusTrue},
				"write":  adminconfig.Decision{Deploy: adminconfig.StatusFalse},
				"delete": adminconfig.Decision{Deploy: adminconfig.StatusFalse},
			},
			OptimizationStrategy: []adminconfig.AttributeOptimization{},
		},
	}
}

// copy flow s3,csv -> s3,csv
func createCopyRequest() *datapath.DataInfo {
	return &datapath.DataInfo{
		DataDetails: &datacatalog.GetAssetResponse{Details: datacatalog.ResourceDetails{
			Connection: taxonomy.Connection{Name: v1alpha1.S3},
			DataFormat: v1alpha1.CSV,
		}},
		Actions:             []taxonomy.Action{},
		StorageRequirements: make(map[taxonomy.ProcessingLocation][]taxonomy.Action),
		Context: &v1alpha1.DataContext{
			DataSetID: "ingest",
			Flow:      taxonomy.CopyFlow,
			Requirements: v1alpha1.DataRequirements{
				Interface:  &taxonomy.Interface{Protocol: v1alpha1.S3, DataFormat: v1alpha1.CSV},
				FlowParams: v1alpha1.FlowRequirements{},
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
		Context: &v1alpha1.DataContext{
			DataSetID: "newAsset",
			Flow:      taxonomy.WriteFlow,
			Requirements: v1alpha1.DataRequirements{
				Interface:  &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight},
				FlowParams: v1alpha1.FlowRequirements{IsNewDataSet: true},
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
			Connection: taxonomy.Connection{Name: v1alpha1.S3},
			DataFormat: v1alpha1.CSV,
		}},
		Actions:             []taxonomy.Action{},
		StorageRequirements: make(map[taxonomy.ProcessingLocation][]taxonomy.Action),
		Context: &v1alpha1.DataContext{
			DataSetID: "write",
			Flow:      taxonomy.WriteFlow,
			Requirements: v1alpha1.DataRequirements{
				Interface:  &taxonomy.Interface{Protocol: v1alpha1.ArrowFlight},
				FlowParams: v1alpha1.FlowRequirements{},
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
			Connection: taxonomy.Connection{Name: v1alpha1.S3},
			DataFormat: v1alpha1.CSV,
		}},
		Actions:             []taxonomy.Action{},
		StorageRequirements: make(map[taxonomy.ProcessingLocation][]taxonomy.Action),
		Context: &v1alpha1.DataContext{
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
	readModule := &v1alpha1.FybrikModule{}
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	account := &v1alpha1.FybrikStorageAccount{}
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
	readModuleS3 := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModuleS3)).NotTo(gomega.HaveOccurred())
	addModule(env, readModuleS3)
	readModuleDB2 := readModuleS3.DeepCopy()
	readModuleDB2.Name = "readDB2"
	readModuleDB2.Spec.Capabilities[0].SupportedInterfaces[0] = v1alpha1.ModuleInOut{Source: &taxonomy.Interface{Protocol: v1alpha1.JdbcDB2}}
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: "xyz"}})
	asset := createReadRequest()
	asset.DataDetails.Details.Connection.Name = v1alpha1.JdbcDB2
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
	readModule := &v1alpha1.FybrikModule{}
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/copy-db2-parquet-no-transforms.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	account := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: string(account.Spec.Region)}})
	asset := createReadRequest()
	asset.DataDetails.Details.Connection.Name = v1alpha1.JdbcDB2
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
	readModule := &v1alpha1.FybrikModule{}
	transformModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-transform.yaml", transformModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addModule(env, transformModule)
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: "xyz"}})
	asset := createReadRequest()
	asset.DataDetails.Details.Connection.Name = v1alpha1.S3
	asset.DataDetails.Details.DataFormat = v1alpha1.Parquet
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
	readModule := &v1alpha1.FybrikModule{}
	transformModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-transform.yaml", transformModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	transformModule.Spec.Capabilities[0].Capability = "read"
	addModule(env, transformModule)
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: "xyz"}})
	asset := createReadRequest()
	asset.DataDetails.Details.Connection.Name = v1alpha1.S3
	asset.DataDetails.Details.DataFormat = v1alpha1.Parquet
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
	readModule := &v1alpha1.FybrikModule{}
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/copy-csv-parquet.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-parquet.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addModule(env, copyModule)
	account := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-theshire.yaml", account)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account)
	remoteGeo := "remote"
	cluster1 := multicluster.Cluster{Name: "c1", Metadata: multicluster.ClusterMetadata{Region: string(account.Spec.Region)}}
	cluster2 := multicluster.Cluster{Name: "c2", Metadata: multicluster.ClusterMetadata{Region: remoteGeo}}
	addCluster(env, cluster1)
	addCluster(env, cluster2)
	asset := createReadRequest()
	asset.DataDetails.ResourceMetadata.Geography = remoteGeo
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
	readModule := &v1alpha1.FybrikModule{}
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addModule(env, copyModule)
	account1 := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-neverland.yaml", account1)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account1)
	account2 := &v1alpha1.FybrikStorageAccount{}
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
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	addModule(env, copyModule)
	account1 := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-neverland.yaml", account1)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account1)
	account2 := &v1alpha1.FybrikStorageAccount{}
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
			StorageAccounts: []adminconfig.Restriction{{Property: "cost", Range: &taxonomy.RangeType{Max: 10}}}},
	}
	addAttribute(env, &taxonomy.InfrastructureElement{
		Attribute: taxonomy.Attribute("cost"),
		Type:      taxonomy.Numeric,
		Value:     "20",
		Object:    taxonomy.StorageAccount,
		Instance:  account1.Name,
	})
	addAttribute(env, &taxonomy.InfrastructureElement{
		Attribute: taxonomy.Attribute("cost"),
		Type:      taxonomy.Numeric,
		Value:     "12",
		Object:    taxonomy.StorageAccount,
		Instance:  account2.Name,
	})
	_, err := solveSingleDataset(env, asset, &testLog)
	g.Expect(err).To(gomega.HaveOccurred())
	// change the restriction to fit one of the accounts
	asset.Configuration.ConfigDecisions["copy"] = adminconfig.Decision{
		Deploy: adminconfig.StatusTrue,
		DeploymentRestrictions: adminconfig.Restrictions{
			StorageAccounts: []adminconfig.Restriction{{Property: "cost", Range: &taxonomy.RangeType{Max: 15}}}},
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
	writeModule := &v1alpha1.FybrikModule{}
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-write.yaml", writeModule)).NotTo(gomega.HaveOccurred())
	addModule(env, writeModule)
	addModule(env, copyModule)
	account1 := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-neverland.yaml", account1)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account1)
	account2 := &v1alpha1.FybrikStorageAccount{}
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
	writeModule := &v1alpha1.FybrikModule{}
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-write.yaml", writeModule)).NotTo(gomega.HaveOccurred())
	addModule(env, writeModule)
	addModule(env, copyModule)
	account1 := &v1alpha1.FybrikStorageAccount{}
	g.Expect(readObjectFromFile("../../testdata/unittests/account-neverland.yaml", account1)).NotTo(gomega.HaveOccurred())
	addStorageAccount(env, account1)
	account2 := &v1alpha1.FybrikStorageAccount{}
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
	writeModule := &v1alpha1.FybrikModule{}
	transformModule := &v1alpha1.FybrikModule{}
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
	deleteModule := &v1alpha1.FybrikModule{}
	writeModule := &v1alpha1.FybrikModule{}
	transformModule := &v1alpha1.FybrikModule{}
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
	workloadLevelModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", workloadLevelModule)).NotTo(gomega.HaveOccurred())
	assetLevelModule := workloadLevelModule.DeepCopy()
	assetLevelModule.Spec.Capabilities[0].Scope = v1alpha1.Asset
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
	if !utils.UseCSP() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &v1alpha1.FybrikModule{}
	copyModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/implicit-copy-batch-module-csv.yaml", copyModule)).NotTo(gomega.HaveOccurred())
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	addModule(env, copyModule)
	clusterRegion := "theshire"
	addCluster(env, multicluster.Cluster{Metadata: multicluster.ClusterMetadata{Region: clusterRegion}})
	asset := createReadRequest()
	asset.Actions = []taxonomy.Action{{Name: "RedactAction"}}
	asset.Configuration.OptimizationStrategy = []adminconfig.AttributeOptimization{{
		Attribute: taxonomy.Attribute("storage-cost"),
		Directive: adminconfig.Minimize,
		Weight:    "1.0",
	}}
	cost := 50
	for i := 0; i < 5; i++ {
		account := &v1alpha1.FybrikStorageAccount{
			Spec: v1alpha1.FybrikStorageAccountSpec{
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
			Attribute: taxonomy.Attribute("storage-cost"),
			Type:      taxonomy.Numeric,
			Value:     fmt.Sprintf("%d", cost),
			Object:    taxonomy.StorageAccount,
			Instance:  account.Name,
		})
		cost += 5
	}
	solution, err := solve(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(2))
	g.Expect(solution.DataPath[0].StorageAccount.Region).To(gomega.Equal(taxonomy.ProcessingLocation("region2")))
	// change the optimization directive to MAX
	asset.Configuration.OptimizationStrategy[0].Directive = adminconfig.Maximize
	solution, err = solve(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(2))
	g.Expect(solution.DataPath[0].StorageAccount.Region).To(gomega.Equal(taxonomy.ProcessingLocation("region4")))
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
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	cost := 10
	for i := 0; i < 5; i++ {
		name := genName("cluster", i)
		addCluster(env, multicluster.Cluster{Name: name, Metadata: multicluster.ClusterMetadata{Region: genName("region", i)}})
		addAttribute(env, &taxonomy.InfrastructureElement{
			Attribute: taxonomy.Attribute("cluster-cost"),
			Type:      taxonomy.Numeric,
			Value:     fmt.Sprintf("%d", cost),
			Object:    taxonomy.Cluster,
			Instance:  name,
		})
		cost -= 1
	}
	asset := createReadRequest()
	asset.Configuration.OptimizationStrategy = []adminconfig.AttributeOptimization{
		{
			Attribute: taxonomy.Attribute("cluster-cost"),
			Directive: adminconfig.Minimize,
			Weight:    "0.2",
		},
		{
			Attribute: taxonomy.Attribute("cluster-cost"),
			Directive: adminconfig.Maximize,
			Weight:    "0.8",
		},
	}
	solution, err := solve(env, asset, &testLog)
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
	if !utils.UseCSP() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	cpuCost := 10
	errRate := 0
	for i := 1; i <= 5; i++ {
		name := genName("cluster", i)
		addCluster(env, multicluster.Cluster{Name: name, Metadata: multicluster.ClusterMetadata{Region: genName("region", i)}})
		addAttribute(env, &taxonomy.InfrastructureElement{
			Attribute: taxonomy.Attribute("cluster-cpu-cost"),
			Type:      taxonomy.Numeric,
			Value:     fmt.Sprintf("%d", cpuCost),
			Object:    taxonomy.Cluster,
			Instance:  name,
		})
		addAttribute(env, &taxonomy.InfrastructureElement{
			Attribute: taxonomy.Attribute("cluster-err-rate"),
			Type:      taxonomy.Numeric,
			Value:     fmt.Sprintf("%d", errRate),
			Object:    taxonomy.Cluster,
			Instance:  name,
		})
		cpuCost -= 1
		if i >= 2 {
			errRate += 10
		}
	}
	asset := createReadRequest()
	asset.Configuration.OptimizationStrategy = []adminconfig.AttributeOptimization{
		{
			Attribute: taxonomy.Attribute("cluster-cpu-cost"),
			Directive: adminconfig.Minimize,
			Weight:    "0.9",
		},
		{
			Attribute: taxonomy.Attribute("cluster-err-rate"),
			Directive: adminconfig.Minimize,
			Weight:    "0.1",
		},
	}
	solution, err := solve(env, asset, &testLog)
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
	if !utils.UseCSP() {
		t.Skip()
	}
	g := gomega.NewGomegaWithT(t)
	env := newEnvironment()
	readModule := &v1alpha1.FybrikModule{}
	g.Expect(readObjectFromFile("../../testdata/unittests/module-read-csv.yaml", readModule)).NotTo(gomega.HaveOccurred())
	addModule(env, readModule)
	cpuCost := 10
	stableRate := 0
	for i := 1; i <= 5; i++ {
		name := genName("cluster", i)
		addCluster(env, multicluster.Cluster{Name: name, Metadata: multicluster.ClusterMetadata{Region: genName("region", i)}})
		addAttribute(env, &taxonomy.InfrastructureElement{
			Attribute: taxonomy.Attribute("cluster-cpu-cost"),
			Type:      taxonomy.Numeric,
			Value:     fmt.Sprintf("%d", cpuCost),
			Object:    taxonomy.Cluster,
			Instance:  name,
		})
		addAttribute(env, &taxonomy.InfrastructureElement{
			Attribute: taxonomy.Attribute("cluster-stability-rate"),
			Type:      taxonomy.Numeric,
			Value:     fmt.Sprintf("%d", stableRate),
			Object:    taxonomy.Cluster,
			Instance:  name,
		})
		cpuCost = 15 - cpuCost - i
		if i == 2 {
			stableRate += 5
		}
	}
	asset := createReadRequest()
	asset.Configuration.OptimizationStrategy = []adminconfig.AttributeOptimization{
		{
			Attribute: taxonomy.Attribute("cluster-cpu-cost"),
			Directive: adminconfig.Minimize,
			Weight:    "0.6",
		},
		{
			Attribute: taxonomy.Attribute("cluster-stability-rate"),
			Directive: adminconfig.Maximize,
			Weight:    "0.4",
		},
	}
	solution, err := solve(env, asset, &testLog)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(solution.DataPath).To(gomega.HaveLen(1))
	g.Expect(solution.DataPath[0].Cluster).To(gomega.Equal("cluster4"))
}
