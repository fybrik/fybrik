package taxonomy

import "fybrik.io/fybrik/pkg/serde"

// +kubebuilder:pruning:PreserveUnknownFields
type AppInfo struct {
	serde.Properties `json:"-"`
}
