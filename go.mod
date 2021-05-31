module github.com/mesh-for-data/mesh-for-data

go 1.13

require (
	cloud.google.com/go v0.72.0 // indirect
	emperror.dev/errors v0.7.0
	github.com/IBM/satcon-client-go v0.1.2-0.20210329192404-b8fa1c732712
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Microsoft/go-winio v0.4.15 // indirect
	github.com/buger/jsonparser v1.1.1
	github.com/containerd/continuity v0.0.0-20201119173150-04c754faca46 // indirect
	github.com/datashim-io/datashim/src/dataset-operator v0.0.0-20210322095623-e5d70b250696
	github.com/docker/docker v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible // indirect
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/render v1.0.1
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/spec v0.19.6 // indirect
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/go-test/deep v1.0.7 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.7
	github.com/hashicorp/hcl v1.0.1-vault // indirect
	github.com/hashicorp/vault/api v1.0.5-0.20201001211907-38d91b749c77
	github.com/hashicorp/vault/sdk v0.1.14-0.20201214222404-d8fffe05d2f4 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/lib/pq v1.8.0 // indirect
	github.com/mailru/easyjson v0.7.1-0.20191009090205-6c0755d89d1e // indirect
	github.com/mattn/go-runewidth v0.0.7 // indirect
	github.com/mitchellh/copystructure v1.1.2 // indirect
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/opencontainers/runc v1.0.0-rc9 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0 // indirect
	github.com/prometheus/common v0.19.0 // indirect
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.6.1
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0
	go.mongodb.org/mongo-driver v1.4.2 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/crypto v0.0.0-20210503195802-e9a32991a82e // indirect
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5 // indirect
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210503173045-b96a97608f20 // indirect
	google.golang.org/grpc v1.36.1
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.2.4
	k8s.io/api v0.19.7
	k8s.io/apiextensions-apiserver v0.19.7 // indirect
	k8s.io/apimachinery v0.19.7
	k8s.io/cli-runtime v0.18.6
	k8s.io/client-go v12.0.0+incompatible
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/cli-utils v0.20.4
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	github.com/google/addlicense => github.com/the-mesh-for-data/addlicense v0.0.0-20200913135744-636c44b42906
	github.com/onsi/gomega => github.com/onsi/gomega v1.10.0
	helm.sh/helm/v3 v3.2.4 => github.com/the-mesh-for-data/helm/v3 v3.2.4-hunchback2
	k8s.io/api => k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.6
	k8s.io/apiserver => k8s.io/apiserver v0.18.6
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.6
	k8s.io/client-go => k8s.io/client-go v0.18.6
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.6
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.6
	k8s.io/code-generator => k8s.io/code-generator v0.18.6
	k8s.io/component-base => k8s.io/component-base v0.18.6
	k8s.io/cri-api => k8s.io/cri-api v0.18.6
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.6
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.6
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.6
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.6
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.6
	k8s.io/kubectl => k8s.io/kubectl v0.18.6
	k8s.io/kubelet => k8s.io/kubelet v0.18.6
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.6
	k8s.io/metrics => k8s.io/metrics v0.18.6
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.6
)
