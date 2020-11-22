module github.com/ibm/the-mesh-for-data

go 1.13

require (
	emperror.dev/errors v0.7.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/render v1.0.1
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.5
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/vault/api v1.0.4
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron v1.2.0
	github.com/stretchr/testify v1.6.1
	google.golang.org/grpc v1.28.1
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.2.4
	istio.io/api v0.0.0-20200723170824-3c2193e74947 // indirect
	istio.io/client-go v0.0.0-20200128004641-c87542c7dc1d
	istio.io/gogo-genproto v0.0.0-20191009201739-17d570f95998 // indirect
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/cli-runtime v0.18.4
	k8s.io/client-go v0.18.6
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/cli-utils v0.20.4
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/yaml v1.2.0
)

replace helm.sh/helm/v3 v3.2.4 => github.com/the-mesh-for-data/helm/v3 v3.2.4-hunchback2

replace github.com/onsi/gomega => github.com/onsi/gomega v1.10.0

replace github.com/google/addlicense => github.com/the-mesh-for-data/addlicense v0.0.0-20200913135744-636c44b42906
