package razee

import (
	"github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/onsi/gomega"
	"testing"

)

const (
	clusterName string = "kind-kind"
	blueprintName string = "blueprint-ysqxbhyihk" // Using the blueprint in manager/testdata/blueprint.yaml
	blueprintNS string = "default"
)

func TestGetBlueprints(t *testing.T) {
	println("orgId: ", orgId)
	println("clusterId: ", clusterId)
	clusterManager := NewRazeeManager(razeeTestURL, loginTestUser, razeeTestPassword, orgId)
	actualBlueprint, err := clusterManager.GetBlueprint(clusterName, blueprintNS, blueprintName)
	if err != nil {
		t.Error(err)
	}
	//t.Log("blueprint.spec: ", blueprint.Spec)
	expectedSpecTemplates := []v1alpha1.ComponentTemplate{
		{
			Name: "implicit-copy-db2wh-to-s3-latest",
			Kind: "M4DModule",
			Resources: []string{
				"kind-registry:5000/m4d-system/m4d-db2wh:0.1.0",
			},
		},
	}
	// Blueprint is a complex object, it's enough to verify they agree on some of it
	t.Log("actualBlueprint.Spec.Templates: ", actualBlueprint.Spec.Templates)
	g := gomega.NewGomegaWithT(t)
	g.Expect(actualBlueprint.Spec.Templates).To(gomega.Equal(expectedSpecTemplates))
}

func TestGetClusters(t *testing.T) {
	clusterManager := NewRazeeManager(razeeTestURL, loginTestUser, razeeTestPassword, orgId)
	clusters, err := clusterManager.GetClusters()
	if err != nil {
		t.Error(err)
	}
	for i, cluster := range clusters {
		t.Log("cluster number: ", i, " cluster metadata: ", cluster)
	}
}