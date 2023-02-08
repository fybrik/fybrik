// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package storagemanager

// Details of the owner application
type ApplicationDetails struct {
	// Application name
	Name string `json:"name"`
	// Application namespace
	Namespace string `json:"namespace"`
	// uuid
	UUID string `json:"uuid"`
}

// Details of the new asset
// The current implementation includes only a name provided in the write flow for a new asset
type DatasetDetails struct {
	Name string `json:"name"`
}

// Configuration options
// TODO: extend IT config policies to return options for storage management
type ConfigOptions struct {
	// Delete an empty folder/bucket when the allocated storage is deleted
	DeleteEmptyFolder bool `json:"deleteEmptyFolder,omitempty"`
}

// Additional options provided for storage allocation/deletion
type Options struct {
	AppDetails        ApplicationDetails `json:"appDetails"`
	DatasetProperties DatasetDetails     `json:"datasetProperties"`
	ConfigurationOpts ConfigOptions      `json:"configurationOpts"`
}
