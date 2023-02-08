// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"testing"

	"github.com/onsi/gomega"
)

// test getProperty
func TestGetProperty(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	properties := map[string]interface{}{"s3": map[string]interface{}{
		"endpoint": "xxx",
		"bucket":   "yyy",
	}}
	val, err := GetProperty(properties, "s3", "endpoint")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(val).To(gomega.Equal("xxx"))
}
