// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/pkg/model/taxonomy"
)

// HelmValues are the values passed to modules during orchestration of the data plane
type HelmValues struct {
	// Asset specific arguments such as data stores and transformations
	fapp.ModuleArguments `json:",inline"`
	// Application context
	Context taxonomy.AppInfo `json:"context,omitempty"`
	// Application and debug labels
	Labels map[string]string `json:"labels"`
	// Application unique identifier
	UUID string `json:"uuid"`
}
