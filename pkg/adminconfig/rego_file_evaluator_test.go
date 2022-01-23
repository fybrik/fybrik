// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig_test

import (
	"context"
	"testing"

	"fybrik.io/fybrik/manager/apis/app/v1alpha1"
	adminconfig "fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/serde"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/open-policy-agent/opa/util"
)

func EvaluatorWithInvalidRules() *adminconfig.RegoPolicyEvaluator {
	module := `
		package adminconfig

		config[{"test": decision}] {
			input.workload.properties.rule == "test-deployment"
			policy := {"ID": "invlaid-status"}
			decision := {"policy": policy, "deploy": "anything"}
		}

		config[{"test": decision}] {
			input.workload.properties.rule == "test-required-policy-id"
			policy := {"name": "invalid-attribute"}
			decision := {"policy": policy}
		}
	`
	// Compile the module. The keys are used as identifiers in error messages.
	compiler, err := ast.CompileModules(map[string]string{
		"example.rego": module,
	})
	Expect(err).ToNot(HaveOccurred())

	rego := rego.New(
		rego.Query("data.adminconfig"),
		rego.Compiler(compiler),
	)
	query, err := rego.PrepareForEval(context.Background())
	Expect(err).ToNot(HaveOccurred())
	return &adminconfig.RegoPolicyEvaluator{Log: logging.LogInit("test", "ConfigPolicyEvaluator"), Query: query}
}

func BaseEvaluator() *adminconfig.RegoPolicyEvaluator {
	module := `
		package adminconfig

		# read scenario, same location
		config[{"test": decision}] {
			policy := {"policySetID": "1", "ID": "test-1"}
			input.request.usage.read == true
			input.request.usage.copy == false
			input.request.dataset.geography == input.workload.cluster.metadata.region
			decision := {"policy": policy, "deploy": "False"}
		}

		# read scenario, different locations
		config[{"test": decision}] {
			input.request.usage.read == true
			input.request.dataset.geography != input.workload.cluster.metadata.region
			clusters :=  { "name": [ "clusterB", "clusterD", "clusterC" ] }
			modules := {"scope": ["asset"]}
			policy := {"policySetID": "1", "ID": "test-2"}
			decision := {"policy": policy, "deploy": "True", "restrictions": {"clusters": clusters, "modules": modules}}
		}
		
		# copy scenario
		config[{"test": decision}] {
			input.request.usage.copy == true
			clusters :=  { "name": [ "clusterA", "clusterB", "clusterC" ] }
			modules := {"type": ["service","plugin","config"]}
			policy := {"policySetID": "1", "ID": "test-3"}
			decision := {"policy": policy, "deploy": "True", "restrictions": {"clusters": clusters, "modules": modules}}
		}

		# write scenario
		config[{"test": decision}] {
			input.request.usage.write == true
			policy := {"policySetID": "2", "ID": "test-4"}
			decision := {"policy": policy, "deploy": "False"}
		}

		# default scenario
		config[{"test": decision}] {
			policy := {"ID": "default", "policySetID": "1"}
			decision := {"policy": policy}
		}
	`
	// Compile the module. The keys are used as identifiers in error messages.
	compiler, err := ast.CompileModules(map[string]string{
		"example.rego": module,
	})
	Expect(err).ToNot(HaveOccurred())

	rego := rego.New(
		rego.Query("data.adminconfig"),
		rego.Compiler(compiler),
	)
	query, err := rego.PrepareForEval(context.Background())
	Expect(err).ToNot(HaveOccurred())
	return &adminconfig.RegoPolicyEvaluator{Log: logging.LogInit("test", "ConfigPolicyEvaluator"), Query: query}
}

func EvaluatorWithInfrastructure() *adminconfig.RegoPolicyEvaluator {
	module := `
		package adminconfig

		# no copy for dev workloads
		config[{"copy": decision}] {
			input.request.usage.read == true
			input.workload.properties.stage == "DEV"
			policy := {"description": "do not copy in DEV workload", "ID": "copy-dev"}
			decision := {"policy": policy, "deploy": "False"}
		}

		# Production Workloads - read
		config[{"read": decision}] {
			input.request.usage.read == true
			input.workload.properties.stage == "PROD"
			workload_region := input.workload.cluster.metadata.region
			policy := {"description": "read in production workload", "ID": "read-prod"}
			clusters := { "metadata.region" : [ workload_region ] }
			decision := {"policy": policy, "deploy": "True", "restrictions": {"clusters": clusters}}
		}

		# Cost Efficient Production Workloads - copy
		config[{"copy": decision}] {
			input.request.usage.read == true
			input.workload.properties.stage == "PROD"
			input.workload.properties.priority != "high"
			dataset_region := input.request.dataset.geography
			workload_region := input.workload.cluster.metadata.region			
			data.infrastructure.bandwidth.values[dataset_region][workload_region] == "S"
			policy := {"description": "use cheaper storage", "ID": "copy-prod-med"}
			accounts := [ data.infrastructure.storageaccounts.values[i].id | data.infrastructure.storageaccounts.values[i].cost <= "80"; 
																	  		 data.infrastructure.storageaccounts.values[i].type == "object-storage";
																			 data.infrastructure.bandwidth.values[data.infrastructure.storageaccounts.values[i].region][workload_region] != "S" ]
			decision := {"policy": policy, "deploy": "True", "restrictions": {"storageaccounts": {"id": accounts}}}
		}

		# High Priority Production Workloads - copy
		config[{"copy": decision}] {
			input.request.usage.read == true
			input.workload.properties.stage == "PROD"
			input.workload.properties.priority == "high"
			dataset_region := input.request.dataset.geography
			workload_region := input.workload.cluster.metadata.region	
			dataset_region != workload_region		
			policy := {"description": "focus on high performance", "ID": "copy-prod-high"}
		    accounts := [data.infrastructure.storageaccounts.values[i].id | data.infrastructure.storageaccounts.values[i].region == workload_region; 
																	 		data.infrastructure.storageaccounts.values[i].type == "object-storage" ]
			decision := {"policy": policy, "deploy": "True", "restrictions": {"storageaccounts": {"id": accounts}}}
		}

		# Transform
		config[{"transform": decision}] {
			policy := {"ID": "transform-geo", "description":"Governance based transformations must take place in the geography where the data is stored"}
			clusters := { "metadata.region" : [ input.request.dataset.geography ] }
			decision := {"policy": policy, "restrictions": {"clusters": clusters}}
		}

	`
	// Compile the module. The keys are used as identifiers in error messages.
	compiler, err := ast.CompileModules(map[string]string{
		"example.rego": module,
	})
	Expect(err).ToNot(HaveOccurred())

	data := `{
		"infrastructure": {
			"bandwidth": {
				"units": "GB/sec",
				"scale": {},
				"values": {
					"region3": {"region3": "L", "region1": "S"},
					"region2": {"region2": "L", "region1": "M"}
				}
			},
			"storageaccounts": {
				"units": "dollar",
				"scale": {},
				"values": [
					{"id": "region1-DB-storage", "region": "region1", "type": "relational-database", "cost": "100"},
					{"id": "region1-object-store", "region": "region1", "type": "object-storage", "cost": "100"},
					{"id": "region2-object-store", "region": "region2", "type": "object-storage", "cost": "80"},
					{"id": "region2-DB-storage", "region": "region2", "type": "relational-database", "cost": "80"},
					{"id": "region3-DB-storage", "region": "region3", "type": "relational-database", "cost": "20"},
					{"id": "region3-object-store", "region": "region3", "type": "object-storage", "cost": "90"}
				]
			}
		}
    }`
	var json map[string]interface{}
	err = util.UnmarshalJSON([]byte(data), &json)
	Expect(err).ToNot(HaveOccurred())
	store := inmem.NewFromObject(json)

	rego := rego.New(
		rego.Query("data.adminconfig"),
		rego.Compiler(compiler),
		rego.Store(store),
	)
	query, err := rego.PrepareForEval(context.Background())
	Expect(err).ToNot(HaveOccurred())
	return &adminconfig.RegoPolicyEvaluator{Log: logging.LogInit("test", "ConfigPolicyEvaluator"), Query: query}
}

func TestRegoFileEvaluator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Policy Evaluator Suite")
}

var _ = Describe("Invalid structure", func() {
	evaluator := EvaluatorWithInvalidRules()
	//nolint:dupl
	It("Invalid deployment status", func() {
		in := adminconfig.EvaluatorInput{
			Workload: adminconfig.WorkloadInfo{
				Properties: taxonomy.AppInfo{
					Properties: serde.Properties{
						Items: map[string]interface{}{"rule": "test-deployment"},
					},
				},
			},
		}
		out, err := evaluator.Evaluate(&in)
		Expect(err).To(HaveOccurred())
		Expect(out.Valid).To(Equal(false))
	})

	//nolint:dupl
	It("Missing policy id", func() {
		in := adminconfig.EvaluatorInput{
			Workload: adminconfig.WorkloadInfo{
				Properties: taxonomy.AppInfo{
					Properties: serde.Properties{
						Items: map[string]interface{}{"rule": "test-required-policy-id"},
					},
				},
			},
		}
		out, err := evaluator.Evaluate(&in)
		Expect(err).To(HaveOccurred())
		Expect(out.Valid).To(Equal(false))
	})
})

var _ = Describe("Evaluate a policy", func() {
	evaluator := BaseEvaluator()
	geo := "theshire"
	//nolint:dupl
	It("Conflict", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: true, v1alpha1.CopyFlow: true},
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{Cluster: multicluster.Cluster{Name: "neverland-cluster", Metadata: multicluster.ClusterMetadata{Region: "neverland"}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(false))
	})

	//nolint:dupl
	It("ValidSolution", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: false, v1alpha1.CopyFlow: false},
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{Cluster: multicluster.Cluster{Name: "thegreendragon", Metadata: multicluster.ClusterMetadata{Region: "theshire"}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions["test"].Deploy).To(Equal(taxonomy.StatusFalse))
	})

	//nolint:dupl
	It("Merge", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: false, v1alpha1.CopyFlow: true},
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{Cluster: multicluster.Cluster{Name: "neverland-cluster", Metadata: multicluster.ClusterMetadata{Region: "neverland"}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions["test"].DeploymentRestrictions["clusters"]["name"]).To(ContainElements("clusterB", "clusterC"))
		Expect(out.ConfigDecisions["test"].DeploymentRestrictions["modules"]["type"]).To(ContainElements("service", "config", "plugin"))
		Expect(out.ConfigDecisions["test"].DeploymentRestrictions["modules"]["scope"]).To(ContainElements("asset"))
	})

	//nolint:dupl
	It("No conflict for policy set 2", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: true, v1alpha1.CopyFlow: true},
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{
				PolicySetID: "2",
				Cluster:     multicluster.Cluster{Name: "neverland-cluster", Metadata: multicluster.ClusterMetadata{Region: "neverland"}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions["test"].Deploy).To(Equal(taxonomy.StatusFalse))
	})

	//nolint:dupl
	It("No decisions for policy set 99", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: true, v1alpha1.CopyFlow: true},
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{
				PolicySetID: "99",
				Cluster:     multicluster.Cluster{Name: "neverland-cluster", Metadata: multicluster.ClusterMetadata{Region: "neverland"}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions).To(BeEmpty())
	})
})

var _ = Describe("Hard policy enforcement", func() {
	evaluator := EvaluatorWithInfrastructure()
	geo := "region3"
	//nolint:dupl
	It("No Copy for DEV Workloads", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: false, v1alpha1.CopyFlow: false},
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{Cluster: multicluster.Cluster{Name: "region1-cluster", Metadata: multicluster.ClusterMetadata{Region: "region1"}},
				Properties: taxonomy.AppInfo{Properties: serde.Properties{Items: map[string]interface{}{"intent": "Fraud Detection", "stage": "DEV", "priority": "low"}}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions["copy"].Deploy).To(Equal(taxonomy.StatusFalse))
	})

	//nolint:dupl
	It("Production Workloads - read", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: false, v1alpha1.CopyFlow: false},
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{Cluster: multicluster.Cluster{Name: "region1-cluster", Metadata: multicluster.ClusterMetadata{Region: "region1"}},
				Properties: taxonomy.AppInfo{Properties: serde.Properties{Items: map[string]interface{}{"intent": "Fraud Detection", "stage": "PROD", "priority": "low"}}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions["read"].DeploymentRestrictions["clusters"]["metadata.region"]).To(ContainElements("region1"))
	})

	//nolint:dupl
	It("Cost Efficient Production Workloads - copy", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: false, v1alpha1.CopyFlow: false},
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{Cluster: multicluster.Cluster{Name: "region1-cluster", Metadata: multicluster.ClusterMetadata{Region: "region1"}},
				Properties: taxonomy.AppInfo{Properties: serde.Properties{Items: map[string]interface{}{"intent": "Fraud Detection", "stage": "PROD", "priority": "low"}}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions["copy"].DeploymentRestrictions["storageaccounts"]["id"]).To(ContainElements("region2-object-store"))
	})

	//nolint:dupl
	It("High Priority Production Workloads - copy", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: false, v1alpha1.CopyFlow: false},
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{Cluster: multicluster.Cluster{Name: "region1-cluster", Metadata: multicluster.ClusterMetadata{Region: "region1"}},
				Properties: taxonomy.AppInfo{Properties: serde.Properties{Items: map[string]interface{}{"intent": "Fraud Detection", "stage": "PROD", "priority": "high"}}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions["copy"].DeploymentRestrictions["storageaccounts"]["id"]).To(ContainElements("region1-object-store"))
	})
})
