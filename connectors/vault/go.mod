module github.com/ibm/the-mesh-for-data/connectors/vault

require github.com/ibm/the-mesh-for-data v0.0.0

replace github.com/ibm/the-mesh-for-data v0.0.0 => ../..

go 1.13

require (
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.2
	github.com/hashicorp/vault v1.6.0
	github.com/hashicorp/vault/api v1.0.5-0.20201001211907-38d91b749c77
	github.com/stretchr/testify v1.6.1
	google.golang.org/grpc v1.29.1
)
