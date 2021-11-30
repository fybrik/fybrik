// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"fmt"
	"os"

	"fybrik.io/fybrik/manager/controllers/utils"
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
//	   Ex: fybrikapplication controller cannot generate a plotter
//	   Ex: Arrow/Flight server used to read data cannot access data store
// error (zerolog.ErrorLevel, 3) - Errors that are not fatal nor panic, but that the user / request initiator is made aware of (typical production setting for stable solution)
//	   Ex: Dataset requested in fybrikapplication.spec is not allowed to be used
// 	   Ex: Query to Arrow/Flight server used to read data returns an error because of incorrect dataset ID
// warn (zerolog.WarnLevel, 2) - Errors not shared with the user / request initiator, typically from which the component recovers on its own
// info (zerolog.InfoLevel, 1) - High level health information that makes it clear the overall status, but without much detail (highest level used in production)
// debug (zerolog.DebugLevel, 0) - Additional information needed to help identify problems (typically used during testing)
// trace (zerolog.TraceLevel, -1) - For tracing step by step flow of control (typically used during development)

// Component Types
const (
	CONTROLLER string = "Controller" // A control plane controller
	MODULE     string = "Module"     // A fybrikmodule that describes a deployable or pre-deployed service
	CONNECTOR  string = "Connector"  // A component that connects an external system to the fybrik control plane - data governance policy manager, data catalog, credential manager
)

// Action Types
const (
	DELETE string = "Delete"
	CREATE string = "Create"
	UPDATE string = "Update"
)

// Log Entry Params - beyond timestamp, caller, err, and msg provided via zerologger
const (
	ACTION        string = "Action"        // optional
	DATASETID     string = "DataSetID"     // optional
	FYBRIKAPPUUID string = "FybrikAppUUID" // mandatory
	FORUSER       string = "ForUser"       // optional
	AUDIT         string = "Audit"         // optional
	CLUSTER       string = "Cluster"       // mandatory
)

// LogInit insures that all log entries have a cluster, timestamp, caller type, file and line from which it was called.
// FYBRIKAPPUUID is mandatory as well, but is not known when the logger is initialized.
func LogInit(callerType string, callerName string, cluster string) zerolog.Logger {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Get the logging verbosity level from the environment variable
	// It should be one of these: https://github.com/rs/zerolog#leveled-logging
	verbosityStr := utils.GetLoggingVerbosity()
	if len(verbosityStr) > 0 {
		var verbosity zerolog.Level
		fmt.Sscan(verbosityStr, verbosity)
		zerolog.SetGlobalLevel(verbosity)
	} else {
		// No environment variable set, so use Info as default
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Initialize the logger with the parameters known at the time of its initiation
	// All entries include timestamp and caller that generated them
	return zerolog.New(os.Stdout).With().Timestamp().Caller().Str(callerType, callerName).Str(CLUSTER, cluster).Logger()
}
