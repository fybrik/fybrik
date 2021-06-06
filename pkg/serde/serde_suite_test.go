package serde_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSerde(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Serde Suite")
}
