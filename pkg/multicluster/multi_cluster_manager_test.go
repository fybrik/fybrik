package multicluster

import (
	"fmt"
	"github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)


type DummyMultiClusterManager struct {
	DeployedBlueprints map[string]*v1alpha1.Blueprint
}

func (m *DummyMultiClusterManager) GetClusters() ([]Cluster, error) {
	return []Cluster{
		{
			Name:     "kind-kind",
			Metadata: ClusterMetadata{},
		},
	}, nil
}

func (m *DummyMultiClusterManager) GetBlueprint(cluster string, namespace string, name string) (*v1alpha1.Blueprint, error) {
	blueprint, found := m.DeployedBlueprints[cluster]
	if found {
		return blueprint, nil
	}
	return nil, fmt.Errorf("Blueprint not found")
}

func (m *DummyMultiClusterManager) CreateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) error {
	m.DeployedBlueprints[cluster] = blueprint
	return nil
}

func (m *DummyMultiClusterManager) UpdateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) error {
	m.DeployedBlueprints[cluster] = blueprint
	return nil
}

func (m *DummyMultiClusterManager) DeleteBlueprint(cluster string, namespace string, name string) error {
	delete(m.DeployedBlueprints, cluster)
	return nil
}

func TestDecodeJsonToRuntimeObject(t *testing.T) {
	var json = `
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
	 "name": "d1",
	 "namespace": "default"
  },
  "spec": {
    "replicas": 2,
    "template": {
	  "spec": {
	    "containers": [
		  {
		    "name": "container",
            "image": "image"
		  }
	    ]
	  }
    }
  }
}
`
	g := gomega.NewGomegaWithT(t)
	actualDeployment := apps.Deployment{}
	scheme := runtime.NewScheme()
	// utilruntime.Must(apps.AddToScheme(scheme))
	if err := decode(json, scheme, &actualDeployment); err != nil {
		println("failed decoding")
		t.Error(err)
	}
	var replicaNumber int32 = 2

	expectedDeployment := apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "d1",
			Namespace: "default",
		},
		Spec: apps.DeploymentSpec{
			Replicas: &replicaNumber,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "container",
							Image: "image",
						},
					},
				},
			},
		},
	}
	g.Expect(expectedDeployment).To(gomega.Equal(actualDeployment))
}
