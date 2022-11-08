include Makefile.env
export DOCKER_TAGNAME ?= 0.0.0
export KUBE_NAMESPACE ?= fybrik-system
export DATA_DIR ?= /tmp
# the latest backward compatible CRD version
export LATEST_BACKWARD_SUPPORTED_CRD_VERSION ?= 0.7.0
export FYBRIK_CHARTS ?= https://fybrik.github.io/charts
# Install latest development version
export DEPLOY_FYBRIK_DEV_VERSION ?= 1
# Install tls certificates used for testing
export DEPLOY_TLS_TEST_CERTS ?= 0
# Copy CA certificate to the system certificates store in the control-plane pods.
# Used for testing tls connection between the control-plane componenets.
export COPY_TEST_CACERTS ?= 0
# If true, run a script to configure Vault. If set to False then
# Vault is configured in commands executed during Vault helm chart
# deployment.
export RUN_VAULT_CONFIGURATION_SCRIPT ?= 1
# Deploy openmetadata catalog in tests
export DEPLOY_OPENMETADATA ?= 0

.PHONY: all
all: generate manifests generate-docs verify

.PHONY: license
license: $(TOOLBIN)/license_finder
	$(call license_go,.)

.PHONY: generate
generate: $(TOOLBIN)/controller-gen $(TOOLBIN)/json-schema-generator
	$(TOOLBIN)/json-schema-generator -r ./manager/apis/app/v1beta1/ -o charts/fybrik/files/taxonomy/
	$(TOOLBIN)/json-schema-generator -r ./pkg/model/... -o charts/fybrik/files/taxonomy/
	$(TOOLBIN)/controller-gen object:headerFile=./hack/boilerplate.go.txt,year=$(shell date +%Y) paths="./..."

.PHONY: generate-docs
generate-docs:
	$(MAKE) -C site generate

.PHONY: manifests
manifests: $(TOOLBIN)/controller-gen $(TOOLBIN)/yq
	$(TOOLBIN)/controller-gen --version
	$(TOOLBIN)/controller-gen crd output:crd:artifacts:config=charts/fybrik-crd/templates/ paths=./manager/apis/...
	$(TOOLBIN)/yq -i eval 'del(.metadata.creationTimestamp)' charts/fybrik-crd/templates/app.fybrik.io_blueprints.yaml
	$(TOOLBIN)/yq -i eval 'del(.metadata.creationTimestamp)' charts/fybrik-crd/templates/app.fybrik.io_fybrikapplications.yaml
	$(TOOLBIN)/yq -i eval 'del(.metadata.creationTimestamp)' charts/fybrik-crd/templates/app.fybrik.io_fybrikmodules.yaml
	$(TOOLBIN)/yq -i eval 'del(.metadata.creationTimestamp)' charts/fybrik-crd/templates/app.fybrik.io_fybrikstorageaccounts.yaml
	$(TOOLBIN)/yq -i eval 'del(.metadata.creationTimestamp)' charts/fybrik-crd/templates/app.fybrik.io_plotters.yaml
	$(TOOLBIN)/controller-gen crd output:crd:artifacts:config=charts/fybrik-crd/charts/asset-crd/templates/ paths=./connectors/katalog/pkg/apis/katalog/...
	$(TOOLBIN)/yq -i eval 'del(.metadata.creationTimestamp)' charts/fybrik-crd/charts/asset-crd/templates/katalog.fybrik.io_assets.yaml
	$(TOOLBIN)/controller-gen webhook paths=./manager/apis/... output:stdout | \
		$(TOOLBIN)/yq eval 'del(.metadata.creationTimestamp)' - | \
		$(TOOLBIN)/yq eval '.metadata.annotations."cert-manager.io/inject-ca-from" |= "{{ .Release.Namespace }}/serving-cert"' - | \
		$(TOOLBIN)/yq eval '.metadata.annotations."certmanager.k8s.io/inject-ca-from" |= "{{ .Release.Namespace }}/serving-cert"' - | \
		$(TOOLBIN)/yq eval '(.metadata.name | select(. == "mutating-webhook-configuration")) = "{{ .Release.Namespace }}-mutating-webhook"' - | \
		$(TOOLBIN)/yq eval '(.metadata.name | select(. == "validating-webhook-configuration")) = "{{ .Release.Namespace }}-validating-webhook"' - | \
		$(TOOLBIN)/yq eval '(.webhooks.[].clientConfig.service.namespace) = "{{ .Release.Namespace }}"' - > charts/fybrik/files/webhook-configs.yaml

.PHONY: docker-mirror-read
docker-mirror-read:
	$(TOOLS_DIR)/docker_mirror.sh $(TOOLS_DIR)/docker_mirror.conf

.PHONY: deploy-fybrik
deploy-fybrik: export VALUES_FILE?=charts/fybrik/values.yaml
deploy-fybrik: $(TOOLBIN)/kubectl $(TOOLBIN)/helm
deploy-fybrik:
	$(TOOLBIN)/kubectl create namespace $(KUBE_NAMESPACE) || true
ifeq ($(DEPLOY_FYBRIK_DEV_VERSION),1)
	$(TOOLBIN)/helm install fybrik-crd charts/fybrik-crd  \
               --namespace $(KUBE_NAMESPACE) --wait --timeout 120s
	$(TOOLBIN)/helm install fybrik charts/fybrik --values $(VALUES_FILE) $(HELM_SETTINGS) \
               --namespace $(KUBE_NAMESPACE) --wait --timeout 120s
else
	$(TOOLBIN)/helm repo add fybrik-charts $(FYBRIK_CHARTS)
	$(TOOLBIN)/helm repo update
	$(TOOLBIN)/helm install fybrik-crd fybrik-charts/fybrik-crd  \
               --namespace $(KUBE_NAMESPACE) --version $(LATEST_BACKWARD_SUPPORTED_CRD_VERSION) --wait --timeout 120s
	$(TOOLBIN)/helm install fybrik charts/fybrik --values $(VALUES_FILE) $(HELM_SETTINGS) \
               --namespace $(KUBE_NAMESPACE) --wait --timeout 120s
endif

.PHONY: pre-test
pre-test: generate manifests $(TOOLBIN)/etcd $(TOOLBIN)/kube-apiserver $(TOOLBIN)/fzn-or-tools
	mkdir -p $(DATA_DIR)/taxonomy
	mkdir -p $(DATA_DIR)/adminconfig
	cp charts/fybrik/files/taxonomy/*.json $(DATA_DIR)/taxonomy/
	cp charts/fybrik/files/adminconfig/* $(DATA_DIR)/adminconfig/
	cp samples/adminconfig/* $(DATA_DIR)/adminconfig/
	mkdir -p manager/testdata/unittests/basetaxonomy
	mkdir -p manager/testdata/unittests/sampletaxonomy
	cp charts/fybrik/files/taxonomy/*.json manager/testdata/unittests/basetaxonomy
	cp charts/fybrik/files/taxonomy/*.json manager/testdata/unittests/sampletaxonomy
	go run main.go taxonomy compile -o manager/testdata/unittests/sampletaxonomy/taxonomy.json \
  	-b charts/fybrik/files/taxonomy/taxonomy.json \
		$(shell find samples/taxonomy/example -type f -name '*.yaml')
	cp manager/testdata/unittests/sampletaxonomy/taxonomy.json $(DATA_DIR)/taxonomy/taxonomy.json

.PHONY: test
test: export MODULES_NAMESPACE?=fybrik-blueprints
test: export CONTROLLER_NAMESPACE?=fybrik-system
test: export CSP_PATH=$(ABSTOOLBIN)/fzn-or-tools
test: pre-test
	go test -v ./...
	USE_CSP=true go test -v ./manager/controllers/app -count 1

.PHONY: run-integration-tests
run-integration-tests: export VALUES_FILE=charts/fybrik/integration-tests.values.yaml
run-integration-tests: export HELM_SETTINGS=--set "manager.solver.enabled=true"
run-integration-tests:
	$(MAKE) setup-cluster
	$(MAKE) -C modules helm
	$(MAKE) -C modules helm-uninstall # Uninstalls the deployed tests from previous command
	$(MAKE) -C pkg/helm test
	$(MAKE) -C samples/rest-server test
	$(MAKE) -C manager run-integration-tests

.PHONY: run-notebook-readflow-tests
run-notebook-readflow-tests: export VALUES_FILE=charts/fybrik/notebook-test-readflow.values.yaml
run-notebook-readflow-tests:
	$(MAKE) setup-cluster
	$(MAKE) -C manager run-notebook-readflow-tests

.PHONY: run-notebook-readflow-tests-om
run-notebook-readflow-tests-om: export HELM_SETTINGS=--set "coordinator.catalog=openmetadata"
run-notebook-readflow-tests-om: export VALUES_FILE=charts/fybrik/notebook-test-readflow.values.yaml
run-notebook-readflow-tests-om: export DEPLOY_OPENMETADATA=1
run-notebook-readflow-tests-om: export CATALOGED_ASSET="openmetadata-s3.default.demo.\"PS_20174392719_1491204439457_log.csv\""
run-notebook-readflow-tests-om:
	$(MAKE) setup-cluster
	$(MAKE) -C manager run-notebook-readflow-tests

.PHONY: run-notebook-readflow-tls-tests
run-notebook-readflow-tls-tests: export VALUES_FILE=charts/fybrik/notebook-test-readflow.tls.values.yaml
run-notebook-readflow-tls-tests: export DEPLOY_TLS_TEST_CERTS=1
run-notebook-readflow-tls-tests: export VAULT_VALUES_FILE=charts/vault/env/ha/vault-single-cluster-values-tls.yaml
run-notebook-readflow-tls-tests: export RUN_VAULT_CONFIGURATION_SCRIPT=0
run-notebook-readflow-tls-tests: export PATCH_FYBRIK_MODULE=1
run-notebook-readflow-tls-tests:
	$(MAKE) setup-cluster
	$(MAKE) -C manager run-notebook-readflow-tests

.PHONY: run-notebook-readflow-tls-system-cacerts-tests
run-notebook-readflow-tls-system-cacerts-tests: export VALUES_FILE=charts/fybrik/notebook-test-readflow.tls-system-cacerts.yaml
run-notebook-readflow-tls-system-cacerts-tests: export FROM_IMAGE=registry.access.redhat.com/ubi8/ubi:8.6
run-notebook-readflow-tls-system-cacerts-tests: export DEPLOY_TLS_TEST_CERTS=1
run-notebook-readflow-tls-system-cacerts-tests: export COPY_TEST_CACERTS=1
run-notebook-readflow-tls-system-cacerts-tests:
	$(MAKE) setup-cluster
	$(MAKE) -C manager run-notebook-readflow-tests

.PHONY: run-notebook-readflow-bc-tests
run-notebook-readflow-bc-tests: export VALUES_FILE=charts/fybrik/notebook-test-readflow.values.yaml
run-notebook-readflow-bc-tests: export DEPLOY_FYBRIK_DEV_VERSION=0
run-notebook-readflow-bc-tests:
	$(MAKE) setup-cluster
	$(MAKE) -C manager run-notebook-readflow-tests

.PHONY: run-notebook-writeflow-tests
run-notebook-writeflow-tests: export VALUES_FILE=charts/fybrik/notebook-test-writeflow.values.yaml
run-notebook-writeflow-tests:
	$(MAKE) setup-cluster
	$(MAKE) -C manager run-notebook-writeflow-tests

.PHONY: run-namescope-integration-tests
run-namescope-integration-tests: export HELM_SETTINGS=--set "clusterScoped=false" --set "applicationNamespace=default"
run-namescope-integration-tests: export VALUES_FILE=charts/fybrik/integration-tests.values.yaml
run-namescope-integration-tests:
	$(MAKE) setup-cluster
	$(MAKE) -C manager run-integration-tests

.PHONY: setup-cluster
setup-cluster: export DOCKER_HOSTNAME?=localhost:5000
setup-cluster: export DOCKER_NAMESPACE?=fybrik-system
setup-cluster:
	$(MAKE) kind
	$(MAKE) cluster-prepare
	$(MAKE) docker-build docker-push
	$(MAKE) -C test/services docker-build docker-push
	$(MAKE) cluster-prepare-wait
	$(MAKE) -C charts test
	$(MAKE) deploy-fybrik
ifeq ($(COPY_TEST_CACERTS), 1)
	cd manager/testdata/notebook/read-flow-tls && ./copy-cacert-to-pods.sh
endif
ifeq ($(RUN_VAULT_CONFIGURATION_SCRIPT),1)
	$(MAKE) configure-vault
endif

.PHONY: cluster-prepare
cluster-prepare:
	$(MAKE) -C third_party/cert-manager deploy
ifeq ($(DEPLOY_TLS_TEST_CERTS),1)
	$(MAKE) -C third_party/kubernetes-reflector deploy
	cd manager/testdata/notebook/read-flow-tls && ./setup-certs.sh
endif
ifeq ($(DEPLOY_OPENMETADATA), 1)
	$(MAKE) -C third_party/openmetadata prepare-openmetadata-for-fybrik
endif
	$(MAKE) -C third_party/vault deploy
	$(MAKE) -C third_party/datashim deploy

.PHONY: cluster-prepare-wait
cluster-prepare-wait:
	$(MAKE) -C third_party/datashim deploy-wait
	$(MAKE) -C third_party/vault deploy-wait

# Build only the docker images needed for integration testing
.PHONY: docker-minimal-it
docker-minimal-it:
	$(MAKE) -C manager docker-build docker-push
	$(MAKE) -C test/services docker-build docker-push

.PHONY: docker-build
docker-build:
	$(MAKE) -C manager docker-build
	$(MAKE) -C connectors docker-build

.PHONY: docker-push
docker-push:
	$(MAKE) -C manager docker-push
	$(MAKE) -C connectors docker-push

DOCKER_PUBLIC_HOSTNAME ?= ghcr.io
DOCKER_PUBLIC_NAMESPACE ?= fybrik
DOCKER_PUBLIC_TAGNAME ?= master

DOCKER_PUBLIC_NAMES := \
	manager \
	katalog-connector \
	opa-connector

define do-docker-retag-and-push-public
	for name in ${DOCKER_PUBLIC_NAMES}; do \
		docker tag ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/$$name:${DOCKER_TAGNAME} ${DOCKER_PUBLIC_HOSTNAME}/${DOCKER_PUBLIC_NAMESPACE}/$$name:${DOCKER_PUBLIC_TAGNAME}; \
	done
	DOCKER_HOSTNAME=${DOCKER_PUBLIC_HOSTNAME} DOCKER_NAMESPACE=${DOCKER_PUBLIC_NAMESPACE} DOCKER_TAGNAME=${DOCKER_PUBLIC_TAGNAME} $(MAKE) docker-push
endef

.PHONY: docker-retag-and-push-public
docker-retag-and-push-public:
	$(call do-docker-retag-and-push-public)

.PHONY: helm-push-public
helm-push-public:
	DOCKER_HOSTNAME=${DOCKER_PUBLIC_HOSTNAME} DOCKER_NAMESPACE=${DOCKER_PUBLIC_NAMESPACE} make -C modules helm-chart-push

.PHONY: save-images
save-images:
	docker save -o images.tar ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/manager:${DOCKER_TAGNAME} \
		${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/katalog-connector:${DOCKER_TAGNAME} \
		${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/opa-connector:${DOCKER_TAGNAME}

include hack/make-rules/tools.mk
include hack/make-rules/verify.mk
include hack/make-rules/cluster.mk
include hack/make-rules/vault.mk
