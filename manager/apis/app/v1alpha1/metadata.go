// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"encoding/json"

	taxonomymodels "fybrik.io/fybrik/pkg/taxonomy/model/datacatalog/base"
)

// +kubebuilder:validation:Type=object
// +kubebuilder:pruning:PreserveUnknownFields
type AssetMetadata struct {
	taxonomymodels.Resource
}

func (metadata *AssetMetadata) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &metadata.Resource)
}

func (metadata *AssetMetadata) MarshalJSON() ([]byte, error) {
	return metadata.Resource.MarshalJSON()
}
