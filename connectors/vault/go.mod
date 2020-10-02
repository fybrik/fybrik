module github.com/ibm/the-mesh-for-data/connectors/vault

require github.com/ibm/the-mesh-for-data v0.0.0

replace github.com/ibm/the-mesh-for-data v0.0.0 => ../..

go 1.13

require (
	//https://github.com/kubernetes/client-go/issues/628
	github.com/Azure/go-autorest v12.2.0+incompatible
	github.com/golang/protobuf v1.4.2
	//https://github.com/hashicorp/vault/issues/9575
	github.com/hashicorp/vault v1.5.0
	github.com/hashicorp/vault/api v1.0.5-0.20200630205458-1a16f3c699c6
	github.com/stretchr/testify v1.6.1
	//github.com/hashicorp/vault/api v1.0.4
	google.golang.org/grpc v1.29.1
	gotest.tools/v3 v3.0.2 // indirect
)

//https://github.com/hashicorp/vault/issues/9575
replace github.com/hashicorp/vault/api => github.com/hashicorp/vault/api v0.0.0-20200718022110-340cc2fa263f
