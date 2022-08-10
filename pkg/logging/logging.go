// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0
package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

// Fybrik recommends using zerolog for golang logging
// Examples of how to use zerolog: https://github.com/rs/zerolog/blob/master/log_example_test.go

// zerolog levels should be used as follows:
// --------------------------------------------
// panic (zerolog.PanicLevel, 5) - Errors that prevent the component from operating correctly and handling requests
//     Ex: fybrik control plane did not deploy correctly
//	   Ex: Data plane component crashed and cannot handle requests
// fatal (zerolog.FatalLevel, 4) - Errors that prevent the component from successfully completing a particular task
//	   Avoid using fatal in the control lane. It causes an Error condition to the pod, which restarts the pod.
// error (zerolog.ErrorLevel, 3) - Errors that are not fatal nor panic, but that the user / request initiator is made
// aware of (typical production setting for stable solution)
//	   Ex: Dataset requested in fybrikapplication.spec is not allowed to be used
// 	   Ex: Query to Arrow/Flight server used to read data returns an error because of incorrect dataset ID
// warn (zerolog.WarnLevel, 2) - Errors not shared with the user / request initiator, typically from which the component
// recovers on its own
// info (zerolog.InfoLevel, 1) - High level health information that makes it clear the overall status, but without
// much detail (highest level used in production)
// debug (zerolog.DebugLevel, 0) - Additional information needed to help identify problems (typically used during testing)
// trace (zerolog.TraceLevel, -1) - For tracing step by step flow of control (typically used during development)

// Component Types
const (
	CONTROLLER string = "Controller" // A control plane controller
	MODULE     string = "Module"     // A fybrikmodule that describes a deployable or pre-deployed service
	CONNECTOR  string = "Connector"  // A component that connects an external system to the fybrik control plane - data governance,
	// policy manager, data catalog, credential manager
	SERVICE string = "Service" // A data plane service - the service itself, not the module that describes it
	SETUP   string = "Setup"   // Used by main function that initializes the control plane
	WEBHOOK string = "Webhook"
)

// Action Types
const (
	DELETE string = "Delete"
	CREATE string = "Create"
	UPDATE string = "Update"
)

// Log Entry Params - those listed in the constants below and
// FybrikApplicationUUID defined in fybrik.io/manager/utils/utils.go
// caller (file and line), error, message, timestamp - Provided by the logging mechanism
// Cluster will not be included since not all components know how to determine on which cluster they run.
// Instead it will be assumed that the logging agents will add this information as they gather the logs.
const (
	ACTION       string = "Action"       // optional
	DATASETID    string = "DataSetID"    // optional
	FORUSER      string = "ForUser"      // optional
	AUDIT        string = "Audit"        // optional
	CLUSTER      string = "Cluster"      // optional
	PLOTTER      string = "Plotter"      // optional
	BLUEPRINT    string = "Blueprint"    // optional
	NAMESPACE    string = "Namespace"    // optional
	RESPONSETIME string = "ResponseTime" // optional
)

// GetLoggingVerbosity returns the level as per https://github.com/rs/zerolog#leveled-logging
func GetLoggingVerbosity() zerolog.Level {
	retDefault := zerolog.TraceLevel
	verbosityStr, ok := os.LookupEnv("LOGGING_VERBOSITY")
	verbosityStr = strings.TrimSpace(verbosityStr)
	if !ok || verbosityStr == "" {
		return retDefault
	}
	fmt.Printf("verbosity %v\n", verbosityStr)
	verbosityInt, err := strconv.Atoi(verbosityStr)
	if err != nil {
		fmt.Printf("Trouble reading verbosity, err = %v. Found %s. Using trace as default", err, verbosityStr)
		return retDefault
	}
	return zerolog.Level(verbosityInt)
}

// GetPrettyLogging returns the indication of whether logs should be human readable or pure json
func PrettyLogging() bool {
	prettyStr, ok := os.LookupEnv("PRETTY_LOGGING")
	if !ok {
		return true
	}
	prettyBool, err := strconv.ParseBool(prettyStr)
	if err != nil {
		fmt.Println("Error parsing PRETTY_LOGGING")
		return true
	}
	return prettyBool
}

// LogInit insures that all log entries have a cluster, timestamp, caller type, file and line from which it was called.
// FybrikAppUuid is mandatory as well, but is not known when the logger is initialized.
func LogInit(callerType, callerName string) zerolog.Logger {
	// Get the logging verbosity level from the environment variable
	// It should be one of these: https://github.com/rs/zerolog#leveled-logging
	verbosity := GetLoggingVerbosity()

	var log zerolog.Logger
	zerolog.SetGlobalLevel(verbosity)

	// Initialize the logger with the parameters known at the time of its initiation
	// All entries include timestamp and caller that generated them

	// If PRETTY_LOGGING
	// Include the filename and line if we're debugging
	if PrettyLogging() {
		log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Str(callerType, callerName).Caller().Logger()
	} else {
		// UNIX Time is faster and smaller than most timestamps
		//		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		log = zerolog.New(os.Stdout).With().Timestamp().Str(callerType, callerName).Caller().Logger()
	}

	log.Debug().Msg("Logging verbosity level is " + log.GetLevel().String())
	return log
}

// LogStructure prints out the provided structure to the log in json format.
func LogStructure(argName string, argStruct interface{}, log *zerolog.Logger, verbosity zerolog.Level, forUser, audit bool) {
	if log.GetLevel() > verbosity {
		return
	}
	var jsonStruct []byte
	var err error
	if PrettyLogging() {
		jsonStruct, err = json.MarshalIndent(argStruct, "", "\t")
	} else {
		jsonStruct, err = json.Marshal(argStruct)
	}

	if err != nil {
		msg := "Failed converting " + argName + " to json: "
		log.WithLevel(verbosity).CallerSkipFrame(1).Bool(FORUSER, forUser).Bool(AUDIT, audit).Msg(msg + fmt.Sprintf("%v", argStruct))
	} else {
		// Log the info making sure that the calling function is listed as the caller
		log.WithLevel(verbosity).CallerSkipFrame(1).Bool(FORUSER, forUser).Bool(AUDIT, audit).Msg(argName + ": " + string(jsonStruct))
	}
}
