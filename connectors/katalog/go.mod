module github.com/ibm/the-mesh-for-data/connectors/katalog

go 1.13

require (
	github.com/ibm/the-mesh-for-data v0.0.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	google.golang.org/grpc v1.28.1
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
)

replace github.com/ibm/the-mesh-for-data v0.0.0 => ../..
