// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package serde_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSerde(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Serde Suite")
}
