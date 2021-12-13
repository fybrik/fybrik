// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	taxonomymodels "fybrik.io/fybrik/pkg/taxonomy/model/policymanager/base"
)

// +kubebuilder:validation:Type=object
// +kubebuilder:pruning:PreserveUnknownFields
type SupportedAction struct {
	taxonomymodels.Action
}

func (action *SupportedAction) UnmarshalJSON(data []byte) error {
	return action.Action.UnmarshalJSON(data)
}

func (action *SupportedAction) MarshalJSON() ([]byte, error) {
	return action.Action.MarshalJSON()
}
