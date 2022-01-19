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
	"fybrik.io/fybrik/pkg/multicluster"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	corev1 "k8s.io/api/core/v1"
)

func NewEvaluator() *adminconfig.RegoPolicyEvaluator {
	module := `
		package adminconfig

		version := "0.0.0"

		# read scenario, same location
		config[{"test": decision}] {
			policy := {"policySetID": "1", "ID": "test-1"}
			input.request.usage.read == true
			input.request.usage.copy == false
			input.request.dataset.geography == input.workload.cluster.metadata.region
			decision := {"policy": policy, "deploy": false}
		}

		# read scenario, different locations
		config[{"test": decision}] {
			input.request.usage.read == true
			input.request.dataset.geography != input.workload.cluster.metadata.region
			clusters :=  { "name": [ "clusterB", "clusterD", "clusterC" ] }
			modules := {"scope": ["asset"]}
			policy := {"policySetID": "1", "ID": "test-2"}
			decision := {"policy": policy, "deploy": true, "restrictions": {"clusters": clusters, "modules": modules}}
		}
		
		# copy scenario
		config[{"test": decision}] {
			input.request.usage.copy == true
			clusters :=  { "name": [ "clusterA", "clusterB", "clusterC" ] }
			modules := {"type": ["service","plugin","config"]}
			policy := {"policySetID": "1", "ID": "test-3"}
			decision := {"policy": policy, "deploy": true, "restrictions": {"clusters": clusters, "modules": modules}}
		}

		# write scenario
		config[{"test": decision}] {
			input.request.usage.write == true
			policy := {"policySetID": "2", "ID": "test-4"}
			decision := {"policy": policy, "deploy": false}
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

func TestRegoFileEvaluator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Policy Evaluator Suite")
}

var _ = Describe("Evaluate a policy", func() {
	evaluator := NewEvaluator()
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
		Expect(out.ConfigDecisions["test"].Deploy).To(Equal(corev1.ConditionFalse))
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
		Expect(out.ConfigDecisions["test"].Deploy).To(Equal(corev1.ConditionFalse))
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
