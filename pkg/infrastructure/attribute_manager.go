// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package infrastructure

import (
	"encoding/json"
	"io/fs"
	"os"

	"fybrik.io/fybrik/pkg/logging"
	"fybrik.io/fybrik/pkg/model/attributes"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// A directory containing rego files that define admin config policies
const RegoPolicyDirectory string = "/tmp/adminconfig/"

// A json file containing the infrastructure information
const InfrastructureInfo string = "infrastructure.json"

// AttributeManager provides access to infrastructure attributes
type AttributeManager struct {
	Log            zerolog.Logger
	Infrastructure attributes.Infrastructure
}

func NewAttributeManager() (*AttributeManager, error) {
	content, err := readInfrastructure()
	if err != nil {
		return nil, err
	}
	return &AttributeManager{Log: logging.LogInit(logging.CONTROLLER, "FybrikApplication"), Infrastructure: content}, nil
}

func readInfrastructure() (attributes.Infrastructure, error) {
	infrastructureFile := RegoPolicyDirectory + InfrastructureInfo
	attributes := attributes.Infrastructure{Items: []attributes.InfrastructureElement{}}
	content, err := os.ReadFile(infrastructureFile)
	if errors.Is(err, fs.ErrNotExist) {
		// file does not exist - return an empty attribute list for backward compatibility
		return attributes, nil
	}
	if err != nil {
		return attributes, err
	}
	if err = json.Unmarshal(content, &attributes); err != nil {
		return attributes, errors.Wrap(err, "could not parse infrastructure json")
	}
	return attributes, nil
}

// GetAttribute returns an infrastructure attribute based on the attribute and instance names
func (m *AttributeManager) GetAttribute(name taxonomy.Attribute, instance string) *attributes.InfrastructureElement {
	for i, element := range m.Infrastructure.Items {
		if element.Attribute == name && element.Instance == instance {
			return &m.Infrastructure.Items[i]
		}
	}
	return nil
}
