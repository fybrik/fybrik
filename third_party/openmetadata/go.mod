module fybrik.io/openmetadata-connector-prepare

go 1.18

require (
	fybrik.io/fybrik v1.1.0
	fybrik.io/openmetadata-connector v0.0.0
)

require (
	fybrik.io/openmetadata-connector/datacatalog-go v0.0.0 // indirect
	fybrik.io/openmetadata-connector/datacatalog-go-client v0.0.0 // indirect
	fybrik.io/openmetadata-connector/datacatalog-go-models v0.0.0 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/rs/zerolog v1.26.0 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

replace (
	fybrik.io/openmetadata-connector => github.com/fybrik/openmetadata-connector v0.0.0-20220919120052-5769dc293e89
	fybrik.io/openmetadata-connector/datacatalog-go => github.com/fybrik/openmetadata-connector/auto-generated/api v0.0.0-20220919120052-5769dc293e89
	fybrik.io/openmetadata-connector/datacatalog-go-client => github.com/fybrik/openmetadata-connector/auto-generated/client v0.0.0-20220919120052-5769dc293e89
	fybrik.io/openmetadata-connector/datacatalog-go-models => github.com/fybrik/openmetadata-connector/auto-generated/models v0.0.0-20220919120052-5769dc293e89
)
