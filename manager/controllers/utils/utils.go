// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sort"

	app "github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	dc "github.com/mesh-for-data/mesh-for-data/pkg/connectors/protobuf"
)

// GetDataFormat returns the existing data format
func GetDataFormat(info *dc.DatasetDetails) (app.DataFormatType, error) {
	switch info.DataFormat {
	case "parquet":
		return app.Parquet, nil
	case "table":
		return app.Table, nil
	case "csv":
		return app.CSV, nil
	case "json":
		return app.JSON, nil
	case "avro":
		return app.AVRO, nil
	case "binary":
		return app.Binary, nil
	case "arrow":
		return app.Arrow, nil
	}
	return app.Binary, fmt.Errorf("unknown format %s", info.DataFormat)
}

// GetProtocol returns the existing data protocol
func GetProtocol(info *dc.DatasetDetails) (app.IFProtocol, error) {
	switch info.DataStore.Type {
	case dc.DataStore_S3:
		return app.S3, nil
	case dc.DataStore_KAFKA:
		return app.Kafka, nil
	case dc.DataStore_DB2:
		return app.JdbcDb2, nil
	}
	return app.S3, errors.New("unknown protocol")
}

// IsTransformation returns true if the data transformation is required
func IsTransformation(actionName string) bool {
	return (actionName != "Allow") // TODO FIX THIS
}

// IsAction returns true if any action is required
func IsAction(actionName string) bool {
	return (actionName != "Allow") // TODO FIX THIS
}

// IsDenied returns true if the data access is denied
func IsDenied(actionName string) bool {
	return (actionName == "Deny") // TODO FIX THIS
}

// StructToMap converts a struct to a map using JSON marshal
func StructToMap(data interface{}) (map[string]interface{}, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	mapData := make(map[string]interface{})
	err = json.Unmarshal(dataBytes, &mapData)
	if err != nil {
		return nil, err
	}
	return mapData, nil
}

// Hash generates a name based on the unique identifier
func Hash(value string, hashLength int) string {
	data := sha256.Sum256([]byte(value))
	hashedStr := hex.EncodeToString(data[:])
	if hashLength >= len(hashedStr) {
		return hashedStr
	}
	return hashedStr[:hashLength]
}

// CreateDataSetIdentifier constructs an identifier for a dataset
// For a JSON string, a concatenation of values is used when keys are sorted alphabetically
func CreateDataSetIdentifier(datasetID string) string {
	jsonMap := make(map[string]string)
	if err := json.Unmarshal([]byte(datasetID), &jsonMap); err != nil {
		return datasetID // not a JSON representation - return the received string
	}
	keys := []string{}
	for key := range jsonMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	id := ""
	for _, key := range keys {
		id += key + "-" + jsonMap[key] + "-"
	}
	return id[:len(id)-1]
}

// Generating release name from step
func GetReleaseName(applicationName string, namespace string, step app.FlowStep) string {
	return GetReleaseNameByStepName(applicationName, namespace, step.Name)
}

// Generate release name from step name
func GetReleaseNameByStepName(applicationName string, namespace string, stepName string) string {
	fullName := applicationName + "-" + namespace + "-" + stepName
	return HelmConformName(fullName)
}

// Generate fqdn for a module
func GenerateModuleEndpointFQDN(releaseName string, blueprintNamespace string) string {
	return releaseName + "." + blueprintNamespace + ".svc.cluster.local"
}

// Some k8s objects only allow for a length of 63 characters.
// This method shortens the name keeping a prefix and using the last 5 characters of the
// new name for the hash of the postfix.
func K8sConformName(name string) string {
	return ShortenedName(name, 63, 5)
}

// Helm has stricter restrictions than K8s and restricts release names to 53 characters
func HelmConformName(name string) string {
	return ShortenedName(name, 53, 5)
}

// Create a name for a step in a blueprint.
// Since this is part of the name of a release, this should be done in a central location to make testing easier
func CreateStepName(moduleName string, assetID string) string {
	return moduleName + "-" + Hash(assetID, 10)
}

// This function shortens a name to the maximum length given and uses rest of the string that is too long
// as hash that gets added to the valid name.
func ShortenedName(name string, maxLength int, hashLength int) string {
	if len(name) > maxLength {
		// The new name is in the form prefix-suffix
		// The prefix is the prefix of the original name (so it's human readable)
		// The suffix is a deterministic hash of the suffix of the original name
		// Overall, the new name is deterministic given the original name
		cutOffIndex := maxLength - hashLength - 1
		prefix := name[:cutOffIndex]
		suffix := Hash(name[cutOffIndex:], hashLength)
		return prefix + "-" + suffix
	}
	return name
}

func ListeningAddress(port int) string {
	address := fmt.Sprintf(":%d", port)
	if runtime.GOOS == "darwin" {
		address = "localhost" + address
	}
	return address
}

// SupportsInterface returns true iff the protocol/format list contains the given protocol/format interface
func SupportsInterface(array []*app.InterfaceDetails, element *app.InterfaceDetails) bool {
	for _, item := range array {
		if item.DataFormat == element.DataFormat && item.Protocol == element.Protocol {
			return true
		}
	}
	return false
}

// SupportsFlow checks whether the given flow element can be found inside the array of flows
func SupportsFlow(array []app.ModuleFlow, element app.ModuleFlow) bool {
	for _, flow := range array {
		if flow == element {
			return true
		}
	}
	return false
}
