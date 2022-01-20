// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/open-policy-agent/opa/util"
	"github.com/pkg/errors"
)

// A directory containing rego files that define admin config policies
const RegoPolicyDirectory string = "/tmp/adminconfig/"

// A json file containing the infrastructure information
const InfrastructureInfo string = "infrastructure.json"

// PrepareQuery prepares a query for OPA evaluation - data object and compiled modules.
// This function is called prior to FybrikApplication controller creation in main.
// Monitoring changes in rego files will be implemented in the future version.
func PrepareQuery() (rego.PreparedEvalQuery, error) {
	// read and compile rego files
	files, err := os.ReadDir(RegoPolicyDirectory)
	if err != nil {
		return rego.PreparedEvalQuery{}, err
	}
	modules := map[string]string{}
	for _, info := range files {
		name := info.Name()
		if !strings.HasSuffix(name, ".rego") {
			continue
		}
		fileName := filepath.Join(RegoPolicyDirectory, name)
		module, err := os.ReadFile(filepath.Clean(fileName))
		if err != nil {
			return rego.PreparedEvalQuery{}, err
		}
		modules[name] = string(module)
	}
	compiler, err := ast.CompileModules(modules)

	if err != nil {
		return rego.PreparedEvalQuery{}, errors.Wrap(err, "couldn't compile modules")
	}

	// sending infrastructure file as data store to OPA
	infrastructureFile := RegoPolicyDirectory + InfrastructureInfo
	json := make(map[string]interface{})
	content, err := os.ReadFile(infrastructureFile)
	if !errors.Is(err, fs.ErrNotExist) {
		// file exists - create a data store, otherwise an empty data store will be used
		if err != nil {
			return rego.PreparedEvalQuery{}, err
		}
		err = util.UnmarshalJSON(content, &json)
		if err != nil {
			return rego.PreparedEvalQuery{}, errors.Wrap(err, "couldn't parse Json")
		}
	}
	store := inmem.NewFromObject(json)

	rego := rego.New(
		rego.Query("data.adminconfig"),
		rego.Compiler(compiler),
		rego.Store(store),
	)
	return rego.PrepareForEval(context.Background())
}
