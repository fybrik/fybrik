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
		
		# copy if regions differ
		config[{"copy": decision}] {
			input.request.usage.read == true
			input.request.dataset.geography != input.workload.cluster.region
			clusters :=  [ data.clusters[i].name | data.clusters[i].metadata.region == input.request.dataset.geography ]
			decision := {"deploy": true, "restrictions": {"clusters": clusters}}
		}
		
		# copy to the workload cluster if copy is requested
		config[{"copy": decision}] {
			input.request.usage.copy == true
			clusters :=  [ data.clusters[i].name | data.clusters[i].name == input.workload.cluster.name ]
			decision := {"deploy": true, "restrictions": {"clusters": clusters}}
		}

		# do not copy in a write scenario
		config[{"copy": decision}] {
			input.request.usage.write == true
			decision := {"deploy": false}
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

})
