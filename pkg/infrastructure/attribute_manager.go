// Copyright 2022 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package infrastructure

import (
	"encoding/json"
	"io/fs"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"fybrik.io/fybrik/pkg/logging"
	infraattributes "fybrik.io/fybrik/pkg/model/attributes"
	"fybrik.io/fybrik/pkg/model/taxonomy"
	"fybrik.io/fybrik/pkg/monitor"
	"fybrik.io/fybrik/pkg/taxonomy/validate"
)

// A directory containing rego files that define admin config policies
const RegoPolicyDirectory string = "/tmp/adminconfig/"

// A json file containing the infrastructure information
const InfrastructureInfo string = "infrastructure.json"

const ValidationPath string = "/tmp/taxonomy/infraattributes.json#/definitions/Infrastructure"

type Dictionary map[taxonomy.Attribute]taxonomy.AttributeDefinition

// AttributeManager provides access to infrastructure attributes
type AttributeManager struct {
	Log zerolog.Logger
	// attribute specific values
	Attributes []taxonomy.InfrastructureElement
	// attribute definitions
	Definitions Dictionary
	Mux         *sync.RWMutex
}

func NewAttributeManager() (*AttributeManager, error) {
	content, err := readInfrastructure()
	if err != nil {
		return nil, err
	}
	attributes, definitions := parseInfrastructureJSON(content)
	return &AttributeManager{
		Log:         logging.LogInit(logging.CONTROLLER, "FybrikApplication"),
		Attributes:  attributes,
		Definitions: definitions,
		Mux:         &sync.RWMutex{},
	}, nil
}

// notification from the file system monitor about an error while getting access to the infrastructure file
func (m *AttributeManager) OnError(err error) {
	m.Log.Error().Err(err).Msg("Error reading infrastructure attributes")
}

// Options for file monitor including the monitored directory and the relevant file extension
func (m *AttributeManager) GetOptions() monitor.FileMonitorOptions {
	return monitor.FileMonitorOptions{Path: RegoPolicyDirectory, Extension: ".json"}
}

// notification from the file monitor on change in the infrastructure json file
func (m *AttributeManager) OnNotify() {
	content, err := readInfrastructure()
	if err != nil {
		m.OnError(err)
	}
	attributes, definitions := parseInfrastructureJSON(content)
	m.Mux.Lock()
	m.Attributes = attributes
	m.Definitions = definitions
	m.Mux.Unlock()
}

func parseInfrastructureJSON(content infraattributes.Infrastructure) ([]taxonomy.InfrastructureElement,
	map[taxonomy.Attribute]taxonomy.AttributeDefinition) {
	dict := map[taxonomy.Attribute]taxonomy.AttributeDefinition{}
	for ind := range content.Definitions {
		dict[content.Definitions[ind].Attribute] = content.Definitions[ind]
	}
	return content.Items, dict
}

// read the infrastructure file and store attribute details in-memory
// The attribute structure is validated with respect to the generated schema (based on taxonomy)
func readInfrastructure() (infraattributes.Infrastructure, error) {
	infrastructureFile := RegoPolicyDirectory + InfrastructureInfo
	attributes := infraattributes.Infrastructure{Items: []taxonomy.InfrastructureElement{}}
	content, err := os.ReadFile(infrastructureFile)
	if errors.Is(err, fs.ErrNotExist) {
		// file does not exist - return an empty attribute list for backward compatibility
		return attributes, nil
	}
	if err != nil {
		return attributes, err
	}
	if err := validateStructure(content); err != nil {
		return attributes, err
	}
	if err := json.Unmarshal(content, &attributes); err != nil {
		return attributes, errors.Wrap(err, "could not parse infrastructure json")
	}
	return attributes, nil
}

func validateStructure(bytes []byte) error {
	allErrs, err := validate.TaxonomyCheck(bytes, ValidationPath)
	if err != nil {
		return err
	}
	if len(allErrs) != 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: "app.fybrik.io", Kind: "infrastructure"}, "", allErrs)
	}
	return nil
}

// GetAttributeValue returns the value of an infrastructure attribute based on the attribute and instance names
func (m *AttributeManager) GetAttributeValue(name taxonomy.Attribute, instance string) (string, bool) {
	for i := range m.Attributes {
		element := &m.Attributes[i]
		if element.Attribute == name && element.Instance == instance {
			return element.Value, true
		}
	}
	return "", false
}

// GetInstanceType returns the instance type associated with the attribute
// TODO: validate that there is only one instance type associated with the given attribute
func (m *AttributeManager) GetInstanceType(name taxonomy.Attribute) *taxonomy.InstanceType {
	if def, found := m.Definitions[name]; found {
		return &def.Object
	}
	return nil
}
