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
	"strings"

	app "github.com/ibm/the-mesh-for-data/manager/apis/app/v1alpha1"
	dc "github.com/ibm/the-mesh-for-data/pkg/connectors/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
)

// DetermineCause attempts to find the reason for an error using its grpc code
func DetermineCause(err error, prefix string) string {
	errStatus, _ := status.FromError(err)
	switch errStatus.Code() {
	case codes.InvalidArgument:
		return "InvalidArgument"
	case codes.PermissionDenied:
		return prefix + "PermissionDenied"
	default:
		return prefix + "CommunicationError"
	}
}

//GetDataFormat returns the existing data format
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
	return app.Binary, errors.New("Unknown format " + info.DataFormat)
}

//GetProtocol returns the existing data protocol
func GetProtocol(info *dc.DatasetDetails) (app.IFProtocol, error) {
	switch info.DataStore.Type {
	case dc.DataStore_S3:
		return app.S3, nil
	case dc.DataStore_KAFKA:
		return app.Kafka, nil
	case dc.DataStore_DB2:
		return app.JdbcDb2, nil
	}
	return app.S3, errors.New("Unknown protocol ")
}

// IsTransformation returns true if the data transformation is required
func IsTransformation(actionName string) bool {
	return (actionName != "Allow") //TODO FIX THIS
}

// IsAction returns true if any action is required
func IsAction(actionName string) bool {
	return (actionName != "Allow") //TODO FIX THIS
}

// IsDenied returns true if the data access is denied
func IsDenied(actionName string) bool {
	return (actionName == "Deny") //TODO FIX THIS
}

// GetAttribute parses a JSON string and returns the required attribute value
func GetAttribute(attribute string, jsonStr string) string {
	obj := make(map[string]string)
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return ""
	}
	return obj[attribute]
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
		id += key + "/" + jsonMap[key] + "/"
	}
	return id[:len(id)-1]
}

// CreateAppIdentifier constructs an identifier for a m4d application: namespace/name.
func CreateAppIdentifier(application *app.M4DApplication) string {
	return application.Namespace + "/" + application.Name
}

func ListeningAddress(port int) string {
	address := fmt.Sprintf(":%d", port)
	if runtime.GOOS == "darwin" {
		address = "localhost" + address
	}
	return address
}

// GetCondition is a helper function to retrieve the relevant condition. Returns nil if not found.
func GetCondition(status *app.M4DApplicationStatus, cType app.ConditionType, reason string) *app.Condition {
	for _, cond := range status.Conditions {
		if cond.Type == cType && cond.Reason == reason {
			return &cond
		}
	}
	return nil
}

// HasCondition returns true if there is an error condition
func HasCondition(status *app.M4DApplicationStatus, cType app.ConditionType) bool {
	for _, cond := range status.Conditions {
		if cond.Type == cType {
			return true
		}
	}
	return false
}

// UpdateCondition updates a condition or adds a new one
func UpdateCondition(status *app.M4DApplicationStatus, cType app.ConditionType, reason string, message string) {
	for ind, cond := range status.Conditions {
		if cond.Type == cType && cond.Reason == reason {
			// A condition already exists: aggregate the error message in order to report multiple errors to the user
			if !strings.Contains(cond.Message, message) {
				// avoid duplicate errors
				// TODO: add a more detailed description to the error message indicating from what dataset/resource it comes from
				status.Conditions[ind].Message += " \n" + message
			}
			return
		}
	}
	status.Conditions = append(status.Conditions, app.Condition{Type: cType, Status: corev1.ConditionTrue, Reason: reason, Message: message})
}

// ActivateCondition sets the required condition details and marks its status as True
func ActivateCondition(appContext *app.M4DApplication, cType app.ConditionType, reason string, message string) {
	UpdateCondition(&appContext.Status, cType, reason, message)
	// specific actions that need to be taken
	if cType == app.ErrorCondition {
		// add failure condition
		if !HasCondition(&appContext.Status, app.FailureCondition) {
			appContext.Status.Conditions = append(appContext.Status.Conditions, app.Condition{
				Type: app.FailureCondition, Status: corev1.ConditionTrue, Reason: "Error", Message: "An error has occurred during blueprint construction."})
		}
	}
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
