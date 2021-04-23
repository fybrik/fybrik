package local

import (
	"testing"

	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ multicluster.ClusterManager = &ClusterManager{}

func TestLocalClusterManager(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	s := scheme.Scheme
	objs := []runtime.Object{
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-metadata",
				Namespace: "m4d-system",
			},
			Data: map[string]string{
				"ClusterName": "remote-cluster",
				"Region":      "Region-1",
				"Zone":        "Zone-1",
			},
		},
	}
	cl := fake.NewFakeClientWithScheme(s, objs...)
	namespace := "m4d-system"
	cm, err := NewManager(cl, namespace)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(cm).NotTo(gomega.BeNil())
	var actualClusters []multicluster.Cluster
	if actualClusters, err = cm.GetClusters(); err != nil {
		t.Errorf("unexpected error in GetClusters: %v", err)
	}

	expectedClusters := []multicluster.Cluster{
		{
			Name: "remote-cluster",
			Metadata: multicluster.ClusterMetadata{
				Region: "Region-1",
				Zone:   "Zone-1",
			},
		},
	}
	g.Expect(expectedClusters).To(gomega.Equal(actualClusters))
}
