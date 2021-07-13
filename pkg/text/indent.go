// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package text

import (
	"strings"
)

// Indent indents a block of text with an indent string
// source: https://play.golang.org/p/nV1_VLau7C
func Indent(text, indent string) string {
	if text[len(text)-1:] == "\n" {
		result := ""
		for _, j := range strings.Split(text[:len(text)-1], "\n") {
			result += indent + j + "\n"
		}
		return result
	}
	result := ""
	for _, j := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		result += indent + j + "\n"
	}
	return result[:len(result)-1]
}
