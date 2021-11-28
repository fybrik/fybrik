// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package fybriklogging

import (
	"encoding/json"

	"github.com/go-logr/logr"
)

// LogType indicates the types of log entries supported.  Logging packages typically provide separate function
// calls for the different log types.
type LogType string

const (
	ERROR LogType = "error" // Errors encountered in the code
	INFO  LogType = "info"  // Info logs are things you want to tell the user or administrator which are not errors.
)

// VerbosityLevel allows one to decide how "chatty" you want your logs to be.
type VerbosityLevel int

const (
	MANDATORY VerbosityLevel = 0
	DEBUG     VerbosityLevel = 1
	TRACE     VerbosityLevel = 2
)

// LogEntry defines the information for a single entry stored in the system logs.
// It is assumed that the timestamp is added automatically and thus not included in the log entry structure.
// Verbosity and log type are added by the log library via which the entries are sent to stdout or stderr.
type LogEntry struct {
	// Caller indicates the cluster, component and function that generated the message in the format "cluster/component/function"
	// Use the buildCallerPath function to construct this.
	// +required
	Caller string `json:"caller"`

	// Message being logged.
	// +required
	Message string `json:"message"`

	// Unique id of the kubernetes fybrikapplication (Needs to be unique over time in order to support history and not just current workloads)
	// +required
	FybrikAppGUID string `json:"fybrikAppGuid"`

	// ForUser should be true if the message should be shared with the end user in the FybrikApplication status.
	// Assumed false if not indicated.
	// +optional
	ForUser bool `json:"forUser,omitempty"`

	// ForArchive should be true if the message should be archived long term.
	// For example, contains full contents of FybrikApplication and its status and should be stored for auditing purposes.
	// Assumed false if not indicated
	// +optional
	ForArchive bool `json:"forArchive,omitempty"`

	// DataSetID is a unique identifier for the data set.
	// It should include an indicator of the catalog from which it came, as well as the catalog id.
	// +optional
	DataSetID string `json:"datasetID,omitempty"`

	// ResponseTime of the current operation in milliseconds.
	// This is an optional parameter that can use in monitoring dashboards such as Kibana.
	// +optional
	ResponseTime int64 `json:"responseTime,omitempty"`

	// Action is the current operation being called.
	// For example, “create_plotter” or “update_blueprint”.
	// +optional
	Action string `json:"action,omitempty"`
}

// ------ User Facing Log Functions ---------
// Components (control plane and data plane) that interface with the user / user workload should log interactions their USER LEVEL log entries via user level logging functions. LogFAError, LogFAStruct, LogFAUpdate
// This ensures that they are logged with ForUser and ForArchive set to true
// Log entries not relevant to users should be logged with the internal logging functions

// LogFAStruct writes to the log file a log entry containing for example, the FybrikApplication structure or a subset (spec or status) of it.
// The log entry indicates it's a user level entry and should be archived.
// This could be used by data path components as well to log user level structures.
// It is logged as INFO/MANDATORY
func LogFAStruct(log logr.Logger, arg interface{}, argName string, caller string, guid string, cluster string, function string) {
	logEntry := LogEntry{
		Caller:        buildCallerPath(cluster, caller, function),
		Message:       "",
		FybrikAppGUID: guid,
		ForUser:       true,
		ForArchive:    true,
	}
	printStructure(arg, argName, log, &logEntry, VerbosityLevel(MANDATORY))
}

// LogFAUpdate writes fybrikapplication updates as INFO/MANDATORY and ForArchive and ForUser are true
// This may be used by data plane components, as well as the control plane, to log user level updates.
func LogFAUpdate(log logr.Logger, caller string, guid string, msg string, cluster string, function string, dataset string, action string) {
	logEntry := LogEntry{
		Caller:        buildCallerPath(cluster, caller, function),
		Message:       msg,
		FybrikAppGUID: guid,
		ForUser:       true,
		ForArchive:    true,
		Action:        action,
		DataSetID:     dataset,
	}

	// Add optional fields where relevant
	if action != "" {
		logEntry.Action = action
	}
	if dataset != "" {
		logEntry.DataSetID = dataset
	}

	log.V(int(MANDATORY)).Info(logEntryToJson(&logEntry))
}

// LogFAError writes fybrikapplication.status errors as ERROR/MANDATORY and ForArchive and ForUser are true
// This may be used by data plane components, as well as the control plane, to log user level errors.
func LogFAError(log logr.Logger, err error, caller string, guid string, msg string, cluster string, function string, dataset string, action string) {
	logEntry := LogEntry{
		Caller:        buildCallerPath(cluster, caller, function),
		Message:       msg,
		FybrikAppGUID: guid,
		ForUser:       true,
		ForArchive:    true,
		Action:        action,
		DataSetID:     dataset,
	}

	// Add optional fields where relevant
	if action != "" {
		logEntry.Action = action
	}
	if dataset != "" {
		logEntry.DataSetID = dataset
	}

	log.V(int(MANDATORY)).Error(err, logEntryToJson(&logEntry))
}

// ----- Non-User Logging Functions --------
// Non-user facing INFO/TRACE and INFO/DEBUG will have ForArchive and ForUser set to false
// Errors for which there is a recovery should be ERROR/DEBUG and ForArchive and ForUser set to false
// Errors for which there is no recovery but not shared with the user (fybrikapplication.status or a data plane component's status) should be ERROR/MANDATORY. ForArchive true and ForUser set to false

// LogDebugInfo prints debugging information.  Verbosity should be either DEBUG or TRACE.
// ForUser and ForArchive are not included in the entry because they are assumed to be false if not present
func LogDebugInfo(log logr.Logger, verbosity VerbosityLevel, caller string, guid string, msg string, cluster string, function string, dataset string, action string, responseTime int64) {
	logEntry := LogEntry{
		Caller:        buildCallerPath(cluster, caller, function),
		Message:       msg,
		FybrikAppGUID: guid,
	}

	// Add optional fields where relevant
	if action != "" {
		logEntry.Action = action
	}
	if dataset != "" {
		logEntry.DataSetID = dataset
	}
	if responseTime != 0 {
		logEntry.ResponseTime = responseTime
	}

	log.V(int(verbosity)).Info(logEntryToJson(&logEntry))
}

// LogError prints errors that may or may not have been shared via fybrikapplication.status
// If recovered is true then it is logged as DEBUG, otherwise as MANDATORY
// ForUser and ForArchive are not included in the entry because they are assumed to be false if not present
func LogError(log logr.Logger, err error, recovered bool, caller string, guid string, msg string, cluster string, function string, dataset string, action string, responseTime int64) {
	logEntry := LogEntry{
		Caller:        buildCallerPath(cluster, caller, function),
		Message:       msg,
		FybrikAppGUID: guid,
	}

	// Add optional fields where relevant
	if action != "" {
		logEntry.Action = action
	}
	if dataset != "" {
		logEntry.DataSetID = dataset
	}
	if responseTime != 0 {
		logEntry.ResponseTime = responseTime
	}

	var verbosity VerbosityLevel
	if recovered == false {
		verbosity = MANDATORY
	} else {
		verbosity = DEBUG
	}
	log.V(int(verbosity)).Error(err, logEntryToJson(&logEntry))
}

// LogDebugStruct writes to the log file a log entry containing a structure for the purpose of logging or tracing.
// LogFAStruct should be used for user level structures since they are archived and can be shared with the fybrikapplication creator.
// Verbosity should usually be either debug or trace
func LogDebugStruct(log logr.Logger, verbosity VerbosityLevel, arg interface{}, argName string, caller string, guid string, cluster string, function string) {
	logEntry := LogEntry{
		Caller:        buildCallerPath(cluster, caller, function),
		Message:       "",
		FybrikAppGUID: guid,
	}
	printStructure(arg, argName, log, &logEntry, verbosity)
}

// ----- INTERNAL FUNCTIONS -------
// Called only by functions in this package

// LogEntryToJson converts the LogEntry structure to json and then to a string, which can be logged.
// If the json creation fails a simple string entry is returned indicated the json error.
func logEntryToJson(entry *LogEntry) string {
	// Convert the structure to json
	jsonStruct, err := json.Marshal(entry)

	if err != nil {
		msg := "Error parsing log entry structure to be logged. Log entry contained message: " + entry.Message
		return msg
	}

	return string(jsonStruct)
}

// printStructure prints the structure in a textual format
func printStructure(argStruct interface{}, argName string, log logr.Logger, entry *LogEntry, verbosity VerbosityLevel) {

	tempMap := make(map[string]interface{}, 1)
	tempMap[argName] = argStruct

	// Convert the structure to json
	jsonStruct, err := json.Marshal(tempMap)

	// Got an error parsing the structure to be logged.  Log the error.
	if err != nil {
		entry.Message = `{"Error parsing the " + argName + " structure to be logged.")`
		entry.Caller = entry.Caller + "/PrintStructure"

		log.V(int(DEBUG)).Error(err, logEntryToJson(entry))
		return
	}

	// Log it
	//	entry.Message = `{` + argName + `:` + string(jsonStruct) + `}`
	entry.Message = string(jsonStruct)
	log.V(int(verbosity)).Info(logEntryToJson(entry))
}

// buildCallerPath concatentates the cluster, component and function into one string, because the log entry puts them in one field
func buildCallerPath(cluster string, component string, function string) string {
	return cluster + "/" + component + "/" + function
}
