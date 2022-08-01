// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package dummy

import (
	"errors"
	"testing"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	app "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/pkg/multicluster"
)

var _ multicluster.ClusterManager = &MockClusterManager{}

func TestDummyMultiClusterManager(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	blueprint := &app.Blueprint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "n",
			Namespace: "ns",
		},
	}
	manager := MockClusterManager{
		DeployedBlueprints: make(map[string]*app.Blueprint),
	}

	// Test listing clusters
	clusters, err := manager.GetClusters()
	g.Expect(clusters).To(gomega.Equal([]multicluster.Cluster{{Name: "kind-kind", Metadata: multicluster.ClusterMetadata{}}}))
	g.Expect(err).To(gomega.BeNil())

	// Test creating a blueprint
	err = manager.CreateBlueprint("kind-kind", blueprint)
	g.Expect(err).To(gomega.BeNil())

	// Test retrieving the before created blueprint
	getBlueprint, err := manager.GetBlueprint("kind-kind", "ns", "n")
	g.Expect(getBlueprint).To(gomega.Equal(blueprint))
	g.Expect(err).To(gomega.BeNil())

	blueprint2 := &app.Blueprint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "n2",
			Namespace: "ns2",
		},
	}

	// Test updating blueprint
	err = manager.UpdateBlueprint("kind-kind", blueprint2)
	g.Expect(err).To(gomega.BeNil())

	// Test retrieving the before updated blueprint
	getBlueprint, err = manager.GetBlueprint("kind-kind", "ns", "n")
	g.Expect(getBlueprint).To(gomega.Equal(blueprint2))
	g.Expect(getBlueprint.Name).To(gomega.Equal("n2"))
	g.Expect(err).To(gomega.BeNil())

	// Test removing the blueprint
	err = manager.DeleteBlueprint("kind-kind", "ns", "n")
	g.Expect(err).To(gomega.BeNil())

	// Test removing a non-existing blueprint (just a no-op)
	err = manager.DeleteBlueprint("kind-kind", "ns", "n")
	g.Expect(err).To(gomega.BeNil())

	// Test retrieving a non-existing blueprint
	getBlueprint, err = manager.GetBlueprint("kind-kind", "ns", "n")
	g.Expect(getBlueprint).To(gomega.BeNil())
	g.Expect(err).To(gomega.Not(gomega.BeNil()))
	g.Expect(err).To(gomega.Equal(errors.New("blueprint not found")))
}
