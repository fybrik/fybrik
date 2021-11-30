module fybrik.io/fybrik

go 1.16

require (
	emperror.dev/errors v0.7.0
	fybrik.io/openapi2crd v0.4.0
	github.com/IBM/satcon-client-go v0.2.1-0.20211027144622-4f54f37377a3
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/apache/arrow/go/arrow v0.0.0-20210907151234-f40856a768f2
	github.com/aws/aws-sdk-go v1.40.37
	github.com/buger/jsonparser v1.1.1
	github.com/containerd/containerd v1.4.11 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/gdexlab/go-render v1.0.1
	github.com/getkin/kin-openapi v0.66.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/gin-gonic/gin v1.7.3
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/go-chi/render v1.0.1
	github.com/go-logr/logr v0.4.0
	github.com/go-test/deep v1.0.7 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.0
	github.com/hashicorp/vault/api v1.2.0
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/mpvl/unique v0.0.0-20150818121801-cbe035fff7de
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.14.0
	github.com/opencontainers/runc v1.0.0-rc95 // indirect
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron v1.2.0
	github.com/rogpeppe/go-internal v1.6.0 // indirect
	github.com/rs/zerolog v1.26.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/pretty v1.0.1
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602
	google.golang.org/genproto v0.0.0-20210707164411-8c882eb9abba // indirect
	google.golang.org/grpc v1.39.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.6.2
	k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/cli-runtime v0.21.0
	k8s.io/client-go v0.21.3
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/cli-utils v0.19.2
	sigs.k8s.io/controller-runtime v0.9.5
	sigs.k8s.io/yaml v1.2.0
)

// This replace is for https://github.com/advisories/GHSA-w73w-5m7g-f7qc
replace github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt v3.2.1+incompatible

replace helm.sh/helm/v3 v3.6.2 => github.com/fybrik/helm/v3 v3.6.2-fybrik-update
