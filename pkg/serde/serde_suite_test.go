// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package serde_test

import (
	"testing"

	"github.com/onsi/gomega"

	"fybrik.io/fybrik/pkg/serde"
)

func TestSubstituteFields(t *testing.T) {
	var json = `
{
  "name": "s3",
  "s3": {
	 "bucket": "%bucket%",
	 "endpoint": "%endpoint%"
  }
}
`
	var expected = `
{
  "name": "s3",
  "s3": {
	 "bucket": "mybucket",
	 "endpoint": "myendpoint"
  }
}
`
	g := gomega.NewGomegaWithT(t)
	p := serde.Properties{}
	g.Expect(p.UnmarshalJSON([]byte(json))).To(gomega.BeNil())
	p.ReplaceTemplateWithValue("%bucket%", "mybucket")
	p.ReplaceTemplateWithValue("%endpoint%", "myendpoint")
	expectedProperties := serde.Properties{}
	g.Expect(expectedProperties.UnmarshalJSON([]byte(expected))).To(gomega.BeNil())
	g.Expect(expectedProperties).To(gomega.Equal(p))
}

func TestSubstituteNonexistingField(t *testing.T) {
	var json = `
{
	"name": "generic",
	"additionalProperties": {
		"assetProperties": [],
		"columns": [],
		"connectionProperties": {
			"name": "s3",
			"s3": {
				"bucket": "%bucket%",
				"endpoint": "%endpoint%",
				"object_key": "%object_key%"
			}
		}
	}
}
`
	var expected = `
{
	"name": "generic",
	"additionalProperties": {
		"assetProperties": [],
		"columns": [],
		"connectionProperties": {
			"name": "s3",
			"s3": {
				"bucket": "mybucket",
				"endpoint": "myendpoint",
				"object_key": "obj.csv"
			}
		}
	}
}
`

	g := gomega.NewGomegaWithT(t)
	p := serde.Properties{}
	g.Expect(p.UnmarshalJSON([]byte(json))).To(gomega.BeNil())
	p.ReplaceTemplateWithValue("%bucket%", "mybucket")
	p.ReplaceTemplateWithValue("%endpoint%", "myendpoint")
	p.ReplaceTemplateWithValue("%object_key%", "obj.csv")
	p.ReplaceTemplateWithValue("%region%", "de")
	expectedProperties := serde.Properties{}
	g.Expect(expectedProperties.UnmarshalJSON([]byte(expected))).To(gomega.BeNil())
	g.Expect(expectedProperties).To(gomega.Equal(p))
}

func TestGetProperty(t *testing.T) {
	var json1 = `
{
	"name": "generic",
	"additionalProperties": {
		"assetProperties": [],
		"columns": [],
		"connectionProperties": {
			"name": "s3",
			"s3": {
				"bucket": "%bucket%",
				"endpoint": "%endpoint%",
				"object_key": "%object_key%",
				"region": "%region%"
			}
		}
	}
}
`
	var json2 = `
{
	"name": "generic",
	"additionalProperties": {
		"assetProperties": [],
		"columns": [],
		"connectionProperties": {
			"name": "s3",
			"s3": {
				"bucket": "mybucket",
				"endpoint": "myendpoint",
				"object_key": "obj.csv"
			}
		}
	}
}
`

	g := gomega.NewGomegaWithT(t)
	p1 := serde.Properties{}
	g.Expect(p1.UnmarshalJSON([]byte(json1))).To(gomega.BeNil())
	p2 := serde.Properties{}
	g.Expect(p2.UnmarshalJSON([]byte(json2))).To(gomega.BeNil())
	g.Expect(p2.MatchPattern(p1, "%object_key%")).To(gomega.Equal("obj.csv"))
	g.Expect(p2.MatchPattern(p1, "%region%")).To(gomega.BeNil())
	g.Expect(p2.MatchPattern(p1, "%something%")).To(gomega.BeNil())
}
