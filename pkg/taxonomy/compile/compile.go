// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package compile

import (
	"github.com/mohae/deepcopy"

	taxonomyio "fybrik.io/fybrik/pkg/taxonomy/io"
	"fybrik.io/fybrik/pkg/taxonomy/model"
)

// Files generates a taxonomy document from a base file and zero or more layer files
func Files(baseDocPath string, layerDocPaths []string, opts ...Option) (*model.Document, error) {
	// load base document
	base, err := taxonomyio.ReadDocumentFromFile(baseDocPath)
	if err != nil {
		return nil, err
	}

	// load layer documents
	layers := make([]*model.Document, 0, len(layerDocPaths))
	for _, path := range layerDocPaths {
		doc, err := taxonomyio.ReadDocumentFromFile(path)
		if err != nil {
			return nil, err
		}
		layers = append(layers, doc)
	}

	// merge all documents
	return Documents(base, layers, opts...)
}

// Documents generates a taxonomy document from a base document and zero or more layer documents
func Documents(base *model.Document, layers []*model.Document, opts ...Option) (*model.Document, error) {
	options := compileOptions{
		codegenTarget: false,
	}
	for _, opt := range opts {
		opt(&options)
	}

	// merge layers on top of base
	baseCopy := deepcopy.Copy(base).(*model.Document)
	documents := append([]*model.Document{baseCopy}, layers...)
	merged, err := mergeDefinitions(documents...)
	if err != nil {
		return nil, err
	}
	merged.Schema = base.Schema
	merged.SchemaVersion = base.SchemaVersion

	// transform into a structural schema
	return transform(base, merged, options.codegenTarget)
}

type compileOptions struct {
	codegenTarget bool
}

// Option is the type for Compile options
type Option func(*compileOptions)

// WithCodeGenerationTarget option to enable generating an output that is more suitable for code generation tools
func WithCodeGenerationTarget(enabled bool) Option {
	return func(h *compileOptions) {
		h.codegenTarget = enabled
	}
}
