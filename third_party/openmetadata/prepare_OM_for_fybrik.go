package main

import (
	logging "fybrik.io/fybrik/pkg/logging"
	openapi_connector_core "fybrik.io/openmetadata-connector/pkg/openmetadata-connector-core"
)

func main() {
	conf := map[interface{}]interface{}{
		"openmetadata_endpoint": "http://localhost:8585/api",
	}
	logger := logging.LogInit(logging.CONNECTOR, "OpenMetadata Prepare for Fybrik")
	openapi_connector_core.NewOpenMetadataAPIService(conf, &logger)
}
