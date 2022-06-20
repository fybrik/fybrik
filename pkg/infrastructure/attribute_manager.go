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

type MetricsDictionary map[string]taxonomy.InfrastructureMetrics

// AttributeManager provides access to infrastructure attributes
type AttributeManager struct {
	Log zerolog.Logger
	// attribute specific values
	Attributes []taxonomy.InfrastructureElement
	// metrics
	Metrics MetricsDictionary
	Mux     *sync.RWMutex
}

func NewAttributeManager() (*AttributeManager, error) {
	content, err := readInfrastructure()
	if err != nil {
		return nil, err
	}
	attributes, metrics := parseInfrastructureJSON(content)
	return &AttributeManager{
		Log:        logging.LogInit(logging.CONTROLLER, "FybrikApplication"),
		Attributes: attributes,
		Metrics:    metrics,
		Mux:        &sync.RWMutex{},
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
	attributes, metrics := parseInfrastructureJSON(content)
	m.Mux.Lock()
	m.Attributes = attributes
	m.Metrics = metrics
	m.Mux.Unlock()
}

func parseInfrastructureJSON(content infraattributes.Infrastructure) ([]taxonomy.InfrastructureElement, MetricsDictionary) {
	dict := MetricsDictionary{}
	for ind := range content.Metrics {
		dict[content.Metrics[ind].Name] = content.Metrics[ind]
	}
	return content.Attributes, dict
}

// read the infrastructure file and store attribute details in-memory
// The attribute structure is validated with respect to the generated schema (based on taxonomy)
func readInfrastructure() (infraattributes.Infrastructure, error) {
	infrastructureFile := RegoPolicyDirectory + InfrastructureInfo
	infra := infraattributes.Infrastructure{Attributes: []taxonomy.InfrastructureElement{}, Metrics: []taxonomy.InfrastructureMetrics{}}
	content, err := os.ReadFile(infrastructureFile)
	if errors.Is(err, fs.ErrNotExist) {
		// file does not exist - return an empty attribute list for backward compatibility
		return infra, nil
	}
	if err != nil {
		return infra, err
	}
	if err := validateStructure(content); err != nil {
		return infra, err
	}
	if err := json.Unmarshal(content, &infra); err != nil {
		return infra, errors.Wrap(err, "could not parse infrastructure json")
	}
	return infra, nil
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
func (m *AttributeManager) GetAttributeValue(name, instance string) (string, bool) {
	for i := range m.Attributes {
		element := &m.Attributes[i]
		if element.Name == name && element.Instance == instance {
			return element.Value, true
		}
	}
	return "", false
}

// GetInstanceType returns instance types associated with the attribute
func (m *AttributeManager) GetInstanceTypes(name string) []taxonomy.InstanceType {
	instanceTypes := []taxonomy.InstanceType{}
	for i := range m.Attributes {
		element := &m.Attributes[i]
		if element.Name == name && !hasInstanceType(element.Object, instanceTypes) {
			instanceTypes = append(instanceTypes, element.Object)
		}
	}
	return instanceTypes
}

func hasInstanceType(value taxonomy.InstanceType, values []taxonomy.InstanceType) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
