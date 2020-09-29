module github.com/ibm/the-mesh-for-data/test/services/pilot/catalog-credential-mock

require github.com/ibm/the-mesh-for-data v0.0.0

replace github.com/ibm/the-mesh-for-data v0.0.0 => ../../../..

replace github.com/ibm/the-mesh-for-data/connectors/vault v0.0.0 => ../../../../connectors/vault

go 1.13

require (
	github.com/golang/protobuf v1.4.2
	github.com/hashicorp/vault/api v1.0.4
	github.com/ibm/the-mesh-for-data/connectors/vault v0.0.0
	google.golang.org/grpc v1.28.1
)
