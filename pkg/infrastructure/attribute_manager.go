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

// AttributeManager provides access to infrastructure attributes
type AttributeManager struct {
	Log            zerolog.Logger
	Infrastructure infraattributes.Infrastructure
	Mux            *sync.RWMutex
}

func NewAttributeManager() (*AttributeManager, error) {
	content, err := readInfrastructure()
	if err != nil {
		return nil, err
	}
	return &AttributeManager{
		Log:            logging.LogInit(logging.CONTROLLER, "FybrikApplication"),
		Infrastructure: content,
		Mux:            &sync.RWMutex{},
	}, nil
}

func (m *AttributeManager) OnError(err error) {
	m.Log.Error().Err(err).Msg("Error reading infrastructure attributes")
}

func (m *AttributeManager) GetOptions() monitor.FileMonitorOptions {
	return monitor.FileMonitorOptions{Path: RegoPolicyDirectory, Extension: ".json"}
}

func (m *AttributeManager) OnNotify() {
	content, err := readInfrastructure()
	if err != nil {
		m.OnError(err)
	}
	m.Mux.Lock()
	m.Infrastructure = content
	m.Mux.Unlock()
}

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

// GetAttribute returns an infrastructure attribute based on the attribute and instance names
func (m *AttributeManager) GetAttribute(name taxonomy.Attribute, instance string) *taxonomy.InfrastructureElement {
	for i := range m.Infrastructure.Items {
		if m.Infrastructure.Items[i].Attribute == name && m.Infrastructure.Items[i].Instance == instance {
			return &m.Infrastructure.Items[i]
		}
	}
	return nil
}

// GetInstanceType returns the instance type associated with the attribute
// TODO: validate that there is only one instance type associated with the given attribute
func (m *AttributeManager) GetInstanceType(name taxonomy.Attribute) *taxonomy.InstanceType {
	for i := range m.Infrastructure.Items {
		if m.Infrastructure.Items[i].Attribute == name {
			return &m.Infrastructure.Items[i].Object
		}
	}
	return nil
}
