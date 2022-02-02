// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package adminconfig

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"fybrik.io/fybrik/pkg/model/adminrules"
	"fybrik.io/fybrik/pkg/taxonomy/validate"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/open-policy-agent/opa/util"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// A directory containing rego files that define admin config policies
const RegoPolicyDirectory string = "/tmp/adminconfig/"

// A json file containing the infrastructure information
const InfrastructureInfo string = "infrastructure.json"

// Schema
const InfrastructureSchema = "/tmp/taxonomy/adminrules.json#/definitions/Infrastructure"

// PrepareQuery prepares a query for OPA evaluation - data object and compiled modules.
// This function is called prior to FybrikApplication controller creation in main.
// Monitoring changes in rego files will be implemented in the future version.
func PrepareQuery() (rego.PreparedEvalQuery, adminrules.Infrastructure, error) {
	// read and compile rego files
	files, err := os.ReadDir(RegoPolicyDirectory)
	if err != nil {
		return rego.PreparedEvalQuery{}, adminrules.Infrastructure{}, err
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
			return rego.PreparedEvalQuery{}, adminrules.Infrastructure{}, err
		}
		modules[name] = string(module)
	}
	compiler, err := ast.CompileModules(modules)

	if err != nil {
		return rego.PreparedEvalQuery{}, adminrules.Infrastructure{}, errors.Wrap(err, "couldn't compile modules")
	}

	// sending infrastructure file as data store to OPA
	infrastructureFile := RegoPolicyDirectory + InfrastructureInfo
	jsonObj := make(map[string]interface{})
	content, err := os.ReadFile(infrastructureFile)
	if errors.Is(err, fs.ErrNotExist) {
		return rego.PreparedEvalQuery{}, adminrules.Infrastructure{}, errors.Wrap(err, "infrastructure.json file does not exist")
	}
	if err != nil {
		return rego.PreparedEvalQuery{}, adminrules.Infrastructure{}, err
	}
	err = util.UnmarshalJSON(content, &jsonObj)
	if err != nil {
		return rego.PreparedEvalQuery{}, adminrules.Infrastructure{}, errors.Wrap(err, "couldn't parse Json")
	}
	infrastructureObject := jsonObj["infrastructure"]
	if infrastructureObject == nil {
		return rego.PreparedEvalQuery{}, adminrules.Infrastructure{}, errors.New("infrastructure object is missing")
	}
	bytes, err := json.Marshal(infrastructureObject)
	if err != nil {
		return rego.PreparedEvalQuery{}, adminrules.Infrastructure{}, err
	}
	infrastructure := adminrules.Infrastructure{}
	err = json.Unmarshal(bytes, &infrastructure)
	if err != nil {
		return rego.PreparedEvalQuery{}, adminrules.Infrastructure{}, err
	}
	err = validateInfrastructure(bytes, InfrastructureSchema)
	if err != nil {
		return rego.PreparedEvalQuery{}, adminrules.Infrastructure{}, err
	}
	store := inmem.NewFromObject(jsonObj)

	rego := rego.New(
		rego.Query("data.adminconfig"),
		rego.Compiler(compiler),
		rego.Store(store),
	)
	query, err := rego.PrepareForEval(context.Background())
	return query, infrastructure, err
}

func validateInfrastructure(bytes []byte, taxonomySchema string) error {
	// validate against taxonomy
	allErrs, err := validate.TaxonomyCheck(bytes, taxonomySchema)
	if err != nil {
		return err
	}
	if len(allErrs) != 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: "app.fybrik.io", Kind: "Infrastructure"},
			"", allErrs)
	}
	return nil
}
