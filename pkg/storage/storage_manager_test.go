// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"testing"

	"github.com/onsi/gomega"

	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/storage/apis/app/v1beta2"
)

// test that the implementation agents have been registered successfully
func TestSupportedConnections(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	connections := GetSupportedConnectionTypes()
	g.Expect(connections).To(gomega.HaveLen(2))
	g.Expect(connections).To(gomega.ContainElement(taxonomy.ConnectionType("s3")))
	g.Expect(connections).To(gomega.ContainElement(taxonomy.ConnectionType("mysql")))
}

// test that the correct implemetation has been selected, and the generated object has the correct connection type
func TestAllocation(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	conn, err := AllocateStorage(&v1beta2.FybrikStorageAccountSpec{Type: "s3"}, nil, nil)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(conn.Name).To(gomega.Equal(taxonomy.ConnectionType("s3")))
}
