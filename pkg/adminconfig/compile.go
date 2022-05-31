// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/pkg/errors"

	"fybrik.io/fybrik/manager/controllers/utils"
)

// RegoPolicyDirectory is a directory containing rego files that
// define admin config policies
var RegoPolicyDirectory = utils.DataRootDir + "/adminconfig/"

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
		var module []byte
		module, err = os.ReadFile(filepath.Clean(fileName))
		if err != nil {
			return rego.PreparedEvalQuery{}, err
		}
		modules[name] = string(module)
	}
	compiler, err := ast.CompileModules(modules)

	if err != nil {
		return rego.PreparedEvalQuery{}, errors.Wrap(err, "couldn't compile modules")
	}

	rg := rego.New(
		rego.Query("data.adminconfig"),
		rego.Compiler(compiler),
	)
	return rg.PrepareForEval(context.Background())
}
