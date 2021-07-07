// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomyio

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/mesh-for-data/mesh-for-data/pkg/taxonomy/model"
	"sigs.k8s.io/yaml"
)

// WriteDocumentToFile writes a document model to a JSON or YAML file.
// The format is auto detected by the filename suffix with a fallback to JSON.
func WriteDocumentToFile(doc *model.Document, outPath string) error {
	var err error
	var encoded []byte
	if strings.HasSuffix(outPath, ".yaml") || strings.HasSuffix(outPath, ".yml") {
		encoded, err = yaml.Marshal(doc)
	} else {
		encoded, err = json.Marshal(doc)
	}
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Clean(outPath), encoded, 0644)
}
