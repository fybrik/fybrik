package multicluster

import (
	"fmt"
	"github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type ClusterManager interface {
	GetClusters() ([]Cluster, error)
	GetBlueprint(cluster string, namespace string, name string) (*v1alpha1.Blueprint, error)
	CreateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) (DeploymentDetails, error)
	UpdateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) (DeploymentDetails, error)
	DeleteBlueprint(cluster string, namespace string, name string) error
}

type ClusterMetadata struct {
	Region string
	Zone   string
}

type Cluster struct {
	Name     string
	Metadata ClusterMetadata
}

type DeploymentDetails struct {
}

type DummyMultiClusterManager struct {
	DeployedBlueprints map[string]*v1alpha1.Blueprint
}

func (m *DummyMultiClusterManager) GetClusters() ([]Cluster, error) {
	return []Cluster{
		Cluster{
			Name:     "kind-kind",
			Metadata: ClusterMetadata{},
		},
	}, nil
}

func (m *DummyMultiClusterManager) GetBlueprint(cluster string, namespace string, name string) (*v1alpha1.Blueprint, error) {
	blueprint, found := m.DeployedBlueprints[cluster]
	if found {
		return blueprint, nil
	} else {
		return nil, fmt.Errorf("Blueprint not found")
	}
}

func (m *DummyMultiClusterManager) CreateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) (DeploymentDetails, error) {
	m.DeployedBlueprints[cluster] = blueprint
	return DeploymentDetails{}, nil
}

func (m *DummyMultiClusterManager) UpdateBlueprint(cluster string, blueprint *v1alpha1.Blueprint) (DeploymentDetails, error) {
	m.DeployedBlueprints[cluster] = blueprint
	return DeploymentDetails{}, nil
}

func (m *DummyMultiClusterManager) DeleteBlueprint(cluster string, namespace string, name string) error {
	delete(m.DeployedBlueprints, cluster)
	return nil
}

// Decode json into runtime.Object, which is a pointer (such as &corev1.ConfigMapList)
func decode(json string, scheme *runtime.Scheme, object runtime.Object) error {
	decoder := serializer.NewCodecFactory(scheme).UniversalDecoder()
	err := runtime.DecodeInto(decoder, []byte(json), object)
	if err != nil {
		return err
	}
	return nil
}
