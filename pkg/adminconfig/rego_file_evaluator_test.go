// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig_test

import (
	"context"
	"testing"

	adminconfig "fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/adminrules"
	"fybrik.io/fybrik/pkg/model/datacatalog"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/multicluster"
	"fybrik.io/fybrik/pkg/serde"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

func EvaluatorWithInvalidRules() *adminconfig.RegoPolicyEvaluator {
	module := `
		package adminconfig

		config[{"capability": "read", "decision": decision}] {
			input.workload.properties.rule == "test-deployment"
			policy := {"ID": "invlaid-status"}
			decision := {"policy": policy, "deploy": "anything"}
		}

		config[{"capability": "read", "decision": decision}] {
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

	rego := rego.New(
		rego.Query("data.test"),
		rego.Compiler(compiler),
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

	//nolint:dupl
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
		Expect(out.ConfigDecisions["copy"].Deploy).To(Equal(adminrules.StatusFalse))
	})

	//nolint:dupl
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

	//nolint:dupl
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
		Expect(out.ConfigDecisions["copy"].Deploy).To(Equal(adminrules.StatusFalse))
	})

	//nolint:dupl
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
