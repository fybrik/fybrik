// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v12

// ObservedState represents a part of the generated Blueprint/Plotter resource status that allows update of FybrikApplication status
type ObservedState struct {
	// Ready represents that the modules have been orchestrated successfully and the data is ready for usage
	Ready bool `json:"ready,omitempty"`
	// Error indicates that there has been an error to orchestrate the modules and provides the error message
	Error string `json:"error,omitempty"`
}
