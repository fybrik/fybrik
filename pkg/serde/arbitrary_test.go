package serde_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mesh-for-data/mesh-for-data/pkg/serde"
)

type ExampleArbitrary struct {
	Text string `json:"text"`
}

var _ = Describe("Arbitrary", func() {
	example := ExampleArbitrary{Text: "abc"}
	arbitrary := serde.NewArbitrary(example)

	It("marshals correctly", func() {
		raw, err := json.Marshal(arbitrary)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(raw)).To(Equal(`{"text":"abc"}`))
	})

	It("unmarshals correctly", func() {
		target := &serde.Arbitrary{}
		err := json.Unmarshal([]byte(`{"text":"abc"}`), target)
		Expect(err).ToNot(HaveOccurred())
		Expect(target.Data).To(Equal(map[string]interface{}{
			"text": "abc",
		}))
	})

	It("unmarshals into a concrete type", func() {
		target := &serde.Arbitrary{}
		err := json.Unmarshal([]byte(`{"text":"abc"}`), target)
		Expect(err).ToNot(HaveOccurred())

		result := &ExampleArbitrary{}
		err = target.Into(result)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.Text).To(Equal("abc"))
	})

	It("rountrip", func() {
		// marshal
		raw, err := json.Marshal(arbitrary)
		Expect(err).ToNot(HaveOccurred())

		// unmarshal
		target := &serde.Arbitrary{}
		err = json.Unmarshal(raw, target)
		Expect(err).ToNot(HaveOccurred())

		// To concrete type
		result := &ExampleArbitrary{}
		err = target.Into(result)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.Text).To(Equal("abc"))
	})

})
