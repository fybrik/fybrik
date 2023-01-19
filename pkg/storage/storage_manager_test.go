// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"testing"

	"github.com/onsi/gomega"

	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/storage/registrator/agent"
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

// test getProperty
func TestGetProperty(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	properties := map[string]interface{}{"s3": map[string]interface{}{
		"endpoint": "xxx",
		"bucket":   "yyy",
	}}
	val, err := agent.GetProperty(properties, taxonomy.ConnectionType("s3"), "endpoint")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(val).To(gomega.Equal("xxx"))
}
