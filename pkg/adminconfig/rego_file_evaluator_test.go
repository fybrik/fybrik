// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig_test

import (
	"context"
	"testing"

	"fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/app/assetmetadata"
	adminconfig "fybrik.io/fybrik/pkg/adminconfig"
	"fybrik.io/fybrik/pkg/multicluster"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func NewInfrastructure() *adminconfig.Infrastructure {
	clusters := []multicluster.Cluster{
		{Name: "clusterA", Metadata: multicluster.ClusterMetadata{Region: "R1"}},
		{Name: "clusterB", Metadata: multicluster.ClusterMetadata{Region: "R1"}},
		{Name: "clusterC", Metadata: multicluster.ClusterMetadata{Region: "R2"}},
		{Name: "clusterD", Metadata: multicluster.ClusterMetadata{Region: "R3"}},
	}
	return &adminconfig.Infrastructure{Clusters: clusters}
}

func NewEvaluator() *adminconfig.RegoPolicyEvaluator {
	data := NewInfrastructure()
	var json map[string]interface{}
	bytes, err := yaml.Marshal(data)
	Expect(err).ToNot(HaveOccurred())
	Expect(yaml.Unmarshal(bytes, &json)).ToNot(HaveOccurred())
	// Manually create the storage layer. inmem.NewFromObject returns an
	// in-memory store containing the supplied data.
	store := inmem.NewFromObject(json)
	module := `
		package adminconfig

		# do not copy in the same region unless requested so
		config[{"copy": decision}] {
			policy := {"policySetID": "1", "ID": "copy-1"}
			input.request.usage.read == true
			input.request.usage.copy == false
			input.request.dataset.geography == input.workload.cluster.metadata.region
			decision := {"policy": policy, "deploy": false}
		}

		# copy if regions differ
		config[{"copy": decision}] {
			input.request.usage.read == true
			input.request.dataset.geography != input.workload.cluster.metadata.region
			clusters :=  [ data.clusters[i].name | data.clusters[i].metadata.region == input.request.dataset.geography ]
			policy := {"policySetID": "1", "ID": "copy-2"}
			decision := {"policy": policy, "deploy": true, "restrictions": {"clusters": clusters}}
		}
		
		# copy to all clusters except clusterD if copy is required
		config[{"copy": decision}] {
			input.request.usage.copy == true
			clusters :=  [ data.clusters[i].name | data.clusters[i].name != "clusterD" ]
			policy := {"policySetID": "1", "ID": "copy-3"}
			decision := {"policy": policy, "deploy": true, "restrictions": {"clusters": clusters}}
		}

		# do not copy in a write scenario
		config[{"copy": decision}] {
			input.request.usage.write == true
			policy := {"policySetID": "2", "ID": "copy-4"}
			decision := {"policy": policy, "deploy": false}
		}

	`
	// Compile the module. The keys are used as identifiers in error messages.
	compiler, err := ast.CompileModules(map[string]string{
		"example.rego": module,
	})
	Expect(err).ToNot(HaveOccurred())

	rego := rego.New(
		rego.Query("data.adminconfig.config"),
		rego.Store(store),
		rego.Compiler(compiler),
	)
	query, err := rego.PrepareForEval(context.Background())
	Expect(err).ToNot(HaveOccurred())
	return &adminconfig.RegoPolicyEvaluator{Data: data, ReadyForEval: true, Query: query}
}

func TestRegoFileEvaluator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Policy Evaluator Suite")
}

var _ = Describe("Evaluate a policy", func() {

	evaluator := NewEvaluator()
	It("Conflict", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: true, v1alpha1.CopyFlow: true},
			Metadata: &assetmetadata.DataDetails{Geography: "R3"}},
			Workload: adminconfig.WorkloadInfo{Cluster: multicluster.Cluster{Name: "clusterA", Metadata: multicluster.ClusterMetadata{Region: "R1"}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(false))
	})

	It("ValidSolution", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: false, v1alpha1.CopyFlow: false},
			Metadata: &assetmetadata.DataDetails{Geography: "R1"}},
			Workload: adminconfig.WorkloadInfo{Cluster: multicluster.Cluster{Name: "clusterA", Metadata: multicluster.ClusterMetadata{Region: "R1"}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions[v1alpha1.Copy].Deploy).To(Equal(corev1.ConditionFalse))
	})

	It("MergeClusters", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: false, v1alpha1.CopyFlow: true},
			Metadata: &assetmetadata.DataDetails{Geography: "R1"}},
			Workload: adminconfig.WorkloadInfo{Cluster: multicluster.Cluster{Name: "clusterA", Metadata: multicluster.ClusterMetadata{Region: "R1"}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions[v1alpha1.Copy].DeploymentRestrictions.Clusters).To(ContainElements("clusterA", "clusterB"))
	})

	It("No conflict for policy set 1", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: true, v1alpha1.CopyFlow: true},
			Metadata: &assetmetadata.DataDetails{Geography: "R2"}},
			Workload: adminconfig.WorkloadInfo{
				PolicySetID: "1",
				Cluster:     multicluster.Cluster{Name: "clusterB", Metadata: multicluster.ClusterMetadata{Region: "R1"}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).ToNot(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions[v1alpha1.Copy].Deploy).To(Equal(corev1.ConditionTrue))
		Expect(out.ConfigDecisions[v1alpha1.Copy].DeploymentRestrictions.Clusters).To(ContainElements("clusterC"))
	})

	It("No decisions for policy set 99", func() {
		in := adminconfig.EvaluatorInput{Request: adminconfig.DataRequest{
			Usage:    map[v1alpha1.DataFlow]bool{v1alpha1.ReadFlow: true, v1alpha1.WriteFlow: true, v1alpha1.CopyFlow: true},
			Metadata: &assetmetadata.DataDetails{Geography: "R1"}},
			Workload: adminconfig.WorkloadInfo{
				PolicySetID: "99",
				Cluster:     multicluster.Cluster{Name: "clusterC", Metadata: multicluster.ClusterMetadata{Region: "R2"}}}}
		out, err := evaluator.Evaluate(&in)
		Expect(err).NotTo(HaveOccurred())
		Expect(out.Valid).To(Equal(true))
		Expect(out.ConfigDecisions[v1alpha1.Copy].Deploy).To(Equal(corev1.ConditionUnknown))
		Expect(out.ConfigDecisions[v1alpha1.Copy].DeploymentRestrictions.Clusters).To(HaveLen(4))
	})
})
