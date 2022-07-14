// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"

	"fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/serde"
)

func EvaluatorWithOptimizations() *adminconfig.RegoPolicyEvaluator {
	module := `
		package adminconfig
		config[{"capability": "read", "decision": decision}] {
			input.request.usage == "read"
			policy := {"ID": "read-request", "version": "0.2"}
			decision := {"policy": policy, "deploy": "True"}
		}
		optimize[decision] {
			input.request.usage == "copy"
			policy := {"ID": "save-cost", "description":"Save storage costs", "version": "0.1"}
			decision := {"policy": policy, "strategy": [{"attribute": "storage-cost", "directive": "min"}]}
		}
		
		optimize[decision] {
			input.request.usage == "read"
			policy := {"ID": "general-strategy", "description":"focus on higher performance while saving storage costs", "version": "0.1"}
			optimize_bandwidth := {"attribute": "bandwidth", "directive": "max", "weight": "0.8"}
			optimize_storage := {"attribute": "storage-cost", "directive": "min", "weight": "0.2"}
			decision := {"policy": policy, "strategy": [optimize_bandwidth,optimize_storage]}
		}	

		optimize[decision] {
			input.request.usage == "read"
			policy := {"ID": "save-cost", "description":"Save storage costs", "version": "0.1"}
			decision := {"policy": policy, "strategy": [{"attribute": "storage-cost", "directive": "min"}]}
		}
	`
	// Compile the module. The keys are used as identifiers in error messages.
	compiler, err := ast.CompileModules(map[string]string{
		"example.rego": module,
	})
	Expect(err).ToNot(HaveOccurred())

	rg := rego.New(
		rego.Query("data.adminconfig"),
		rego.Compiler(compiler),
	)
	query, err := rg.PrepareForEval(context.Background())
	Expect(err).ToNot(HaveOccurred())
	return adminconfig.NewRegoPolicyEvaluatorWithQuery(query)
}

func BaseEvaluator() *adminconfig.RegoPolicyEvaluator {
	module := `
		package test
		# read scenario, same location
		config[{"capability": "copy", "decision": decision}] {
			policy := {"policySetID": "1", "ID": "test-1"}
			input.workload.properties.stage == "PROD"
			input.workload.properties.severity != "critical"
			input.request.dataset.geography == input.workload.cluster.metadata.region
			decision := {"policy": policy, "deploy": "False"}
		}
		# read scenario, different locations
		config[{"capability": "copy", "decision": decision}] {
			input.workload.properties.stage == "PROD"
			input.request.dataset.geography != input.workload.cluster.metadata.region
			clusters :=  { "property": "name", "values": [ "clusterB", "clusterD", "clusterC" ] }
			modules := {"property": "scope", "values": ["asset"]}
			policy := {"policySetID": "1", "ID": "test-2"}
			decision := {"policy": policy, "deploy": "True", "restrictions": {"clusters": [clusters], "modules": [modules]}}
		}
		
		# copy scenario
		config[{"capability": "copy", "decision": decision}] {
			input.workload.properties.severity == "critical"
			clusters :=  { "property": "name", "values": [ "clusterB", "clusterA", "clusterC" ] }
			modules := {"property": "type", "values": ["service","plugin","config"]}
			policy := {"policySetID": "1", "ID": "test-3"}
			decision := {"policy": policy, "deploy": "True", "restrictions": {"clusters": [clusters], "modules": [modules]}}
		}
		# write scenario
		config[{"capability": "copy", "decision": decision}] {
			input.workload.properties.priority == "high"
			policy := {"policySetID": "2", "ID": "test-4"}
			decision := {"policy": policy, "deploy": "False"}
		}
		# default scenario
		config[{"capability": "copy", "decision": decision}] {
			policy := {"ID": "default", "policySetID": "1"}
			decision := {"policy": policy}
		}
	`
	// Compile the module. The keys are used as identifiers in error messages.
	compiler, err := ast.CompileModules(map[string]string{
		"example.rego": module,
	})
	Expect(err).ToNot(HaveOccurred())

	rg := rego.New(
		rego.Query("data.test"),
		rego.Compiler(compiler),
	)
	query, err := rg.PrepareForEval(context.Background())
	Expect(err).ToNot(HaveOccurred())
	return adminconfig.NewRegoPolicyEvaluatorWithQuery(query)
}

func EvaluatorforExpiringPolicies() *adminconfig.RegoPolicyEvaluator {
	module := `
		package test
		# copy scenario, same location, vaild from 2022.1.1, expire on 2022.6.1
		config[{"capability": "copy", "decision": decision}] {
			policy := {"policySetID": "1", "ID": "test-1"}
			input.workload.properties.stage == "PROD"
			input.workload.properties.severity != "critical"
			input.request.dataset.geography == input.workload.cluster.metadata.region
			nowDate := time.now_ns()
			startDate := time.parse_rfc3339_ns("2022-01-01T00:00:00Z")
			nowDate >= startDate
			decision := {"policy": policy, "deploy": "False"}
		}
		# copy scenario, different locations, vaild from 2022.1.1, expire on 2022.2.1
		config[{"capability": "copy", "decision": decision}] {
			policy := {"policySetID": "1", "ID": "test-1"}
			input.workload.properties.stage == "PROD"
			input.workload.properties.severity != "critical"
			input.request.dataset.geography != input.workload.cluster.metadata.region
			nowDate := time.now_ns()
			startDate := time.parse_rfc3339_ns("2022-01-01T00:00:00Z")
			expiration := time.parse_rfc3339_ns("2022-02-01T00:00:00Z")
			nowDate >= startDate
			nowDate < expiration
			decision := {"policy": policy, "deploy": "False"}
		}
	`
	// Compile the module. The keys are used as identifiers in error messages.
	compiler, err := ast.CompileModules(map[string]string{
		"example.rego": module,
	})
	Expect(err).ToNot(HaveOccurred())

	rg := rego.New(
		rego.Query("data.test"),
		rego.Compiler(compiler),
	)
	query, err := rg.PrepareForEval(context.Background())
	Expect(err).ToNot(HaveOccurred())
	return adminconfig.NewRegoPolicyEvaluatorWithQuery(query)
}

func TestRegoFileEvaluator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Policy Evaluator Suite")
}

var _ = Describe("Evaluate a policy", func() {
	evaluator := BaseEvaluator()
	geo := "theshire"

	It("Conflict", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{
				Cluster: multicluster.Cluster{
					Name:     "neverland-cluster",
					Metadata: multicluster.ClusterMetadata{Region: "neverland"},
				},
				Properties: taxonomy.AppInfo{
					Properties: serde.Properties{
						Items: map[string]interface{}{
							"stage":    "PROD",
							"priority": "high",
							"severity": "critical",
						},
					},
				},
			},
		}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(false))
	})

	It("ValidSolution", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{
				Cluster: multicluster.Cluster{
					Name:     "thegreendragon",
					Metadata: multicluster.ClusterMetadata{Region: "theshire"},
				},
				Properties: taxonomy.AppInfo{
					Properties: serde.Properties{
						Items: map[string]interface{}{
							"stage":    "PROD",
							"priority": "medium",
							"severity": "low",
						},
					},
				},
			},
		}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions["copy"].Deploy).To(Equal(adminconfig.StatusFalse))
	})

	It("Merge", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{
				Cluster: multicluster.Cluster{
					Name:     "neverland-cluster",
					Metadata: multicluster.ClusterMetadata{Region: "neverland"},
				},
				Properties: taxonomy.AppInfo{
					Properties: serde.Properties{
						Items: map[string]interface{}{
							"stage":    "PROD",
							"priority": "normal",
							"severity": "critical",
						},
					},
				},
			},
		}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		for _, restrict := range out.ConfigDecisions["copy"].DeploymentRestrictions.Clusters {
			Expect(restrict.Property).To(Equal("name"))
			Expect(restrict.Values).To(ContainElements("clusterB", "clusterC"))
		}
		for _, restrict := range out.ConfigDecisions["copy"].DeploymentRestrictions.Modules {
			if restrict.Property == "type" {
				Expect(restrict.Values).To(ContainElements("service", "config", "plugin"))
			} else {
				Expect(restrict.Values).To(ContainElement("asset"))
			}
		}
	})

	It("No conflict for policy set 2", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{
				PolicySetID: "2",
				Properties: taxonomy.AppInfo{
					Properties: serde.Properties{
						Items: map[string]interface{}{
							"stage":    "PROD",
							"priority": "high",
							"severity": "critical",
						},
					},
				},
				Cluster: multicluster.Cluster{
					Name:     "neverland-cluster",
					Metadata: multicluster.ClusterMetadata{Region: "neverland"},
				},
			},
		}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions["copy"].Deploy).To(Equal(adminconfig.StatusFalse))
	})

	It("No decisions for policy set 99", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{
				PolicySetID: "99",
				Properties: taxonomy.AppInfo{
					Properties: serde.Properties{
						Items: map[string]interface{}{
							"stage":    "PROD",
							"priority": "high",
							"severity": "critical",
						},
					},
				},
				Cluster: multicluster.Cluster{
					Name:     "neverland-cluster",
					Metadata: multicluster.ClusterMetadata{Region: "neverland"},
				},
			},
		}
		out, err := evaluator.Evaluate(&in)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions).To(BeEmpty())
	})
})

var _ = Describe("Optimizations", func() {
	evaluator := EvaluatorWithOptimizations()
	It("SingleStrategy", func() {
		in := adminconfig.EvaluatorInput{
			Request: adminconfig.DataRequest{
				Usage: taxonomy.CopyFlow,
			},
		}
		out, err := evaluator.Evaluate(&in)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.OptimizationStrategy).To(HaveLen(1))
		Expect(string(out.OptimizationStrategy[0].Attribute)).To(Equal("storage-cost"))
	})

	It("Multiple strategies", func() {
		in := adminconfig.EvaluatorInput{
			Request: adminconfig.DataRequest{
				Usage: taxonomy.ReadFlow,
			},
		}
		out, err := evaluator.Evaluate(&in)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.OptimizationStrategy).To(HaveLen(2))
	})
})

var _ = Describe("Expiring policies", func() {
	evaluator := EvaluatorforExpiringPolicies()
	geo := "theshire"

	It("ValidPolicy", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{
				Cluster: multicluster.Cluster{
					Name:     "thegreendragon",
					Metadata: multicluster.ClusterMetadata{Region: "theshire"},
				},
				Properties: taxonomy.AppInfo{
					Properties: serde.Properties{
						Items: map[string]interface{}{
							"stage":    "PROD",
							"priority": "medium",
							"severity": "low",
						},
					},
				},
			},
		}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions["copy"].Deploy).To(Equal(adminconfig.StatusFalse))
	})

	It("ExpiredPolicy", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Metadata: &datacatalog.ResourceMetadata{Geography: geo}},
			Workload: adminconfig.WorkloadInfo{
				Cluster: multicluster.Cluster{
					Name:     "neverland-cluster",
					Metadata: multicluster.ClusterMetadata{Region: "neverland"},
				},
				Properties: taxonomy.AppInfo{
					Properties: serde.Properties{
						Items: map[string]interface{}{
							"stage":    "PROD",
							"priority": "medium",
							"severity": "low",
						},
					},
				},
			},
		}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions).To(BeEmpty())
	})
})
