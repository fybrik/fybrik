ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env
#include Makefile.env

.PHONY: all
all: manager

include $(ROOT_DIR)/hack/make-rules/tools.mk
DOCKER_NAME ?= manager
GO_OUTPUT_FILE = manager
KUBE_NAMESPACE ?= m4d-system
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/verify.mk

config/default/kustomization.yaml: config/default/kustomization.yaml.in
	cp $< $@

# Run tests
test: generate manifests $(TOOLBIN)/etcd $(TOOLBIN)/kube-apiserver
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: generate
generate: $(TOOLBIN)/controller-gen
	$(TOOLBIN)/controller-gen object:headerFile=$(ROOT_DIR)/hack/boilerplate.go.txt,year=$(shell date +%Y) paths="./..."

# Generate code
.PHONY: manifests
manifests: $(TOOLBIN)/controller-gen $(TOOLBIN)/yq
	$(TOOLBIN)/controller-gen crd:trivialVersions=true webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	$(TOOLBIN)/controller-gen crd:trivialVersions=true output:crd:artifacts:config=$(ROOT_DIR)/charts/m4d-crd/templates/ paths=./apis/...
	$(TOOLBIN)/controller-gen webhook paths=./apis/... output:stdout | \
		$(TOOLBIN)/yq eval '.metadata.annotations."cert-manager.io/inject-ca-from" |= "{{ .Release.Namespace }}/serving-cert"' - | \
		$(TOOLBIN)/yq eval '(.metadata.name | select(. == "mutating-webhook-configuration")) = "{{ .Release.Namespace }}-mutating-webhook"' - | \
		$(TOOLBIN)/yq eval '(.metadata.name | select(. == "validating-webhook-configuration")) = "{{ .Release.Namespace }}-validating-webhook"' - | \
		$(TOOLBIN)/yq eval '(.webhooks.[].clientConfig.service.namespace) = "{{ .Release.Namespace }}"' - > $(ROOT_DIR)/charts/m4d/files/webhook-configs.yaml

# Install CRDs into a cluster
.PHONY: install
install: $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl manifests
	$(TOOLBIN)/kustomize build config/crd | $(TOOLBIN)/kubectl apply -f -

# Uninstall CRDs from a cluster
.PHONY: uninstall
uninstall: $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl manifests
	$(TOOLBIN)/kustomize build config/crd | $(TOOLBIN)/kubectl delete -f -

# Install core into charts/m4d
.PHONY: charts
charts:
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${IMG}
	$(TOOLBIN)/kustomize build config/default > $(ROOT_DIR)/charts/m4d/templates/manager.yaml

.PHONY: deploy-crd
deploy-crd:
	kustomize build config/crd | kubectl apply -f -

.PHONY: undeploy-crd
undeploy-crd:
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl config/default/kustomization.yaml
	$(TOOLBIN)/kubectl create namespace m4d-system || true
	$(TOOLBIN)/kubectl create namespace m4d-blueprints || true
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${IMG}
	$(TOOLBIN)/kustomize build config/default | $(TOOLBIN)/kubectl apply -f -

# Delete controller in the configured Kubernetes cluster in ~/.kube/config
undeploy: $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl manifests config/default/kustomization.yaml
	$(TOOLBIN)/kustomize build config/default | $(TOOLBIN)/kubectl delete -f -

# Overwrite docker-build from docker.mk
docker-build: generate
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o manager main.go
	docker build . -t ${IMG} -f Dockerfile
	rm manager

# Deploy only movement-controller in the configured Kubernetes cluster in ~/.kube/config
deploy_mc: manifests $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl config/default/kustomization.yaml
	$(TOOLBIN)/kubectl create namespace ${KUBE_NAMESPACE} || true
	$(TOOLBIN)/kubectl create namespace m4d-blueprints || true
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/manager:${DOCKER_TAGNAME}
	$(TOOLBIN)/kustomize build --load_restrictor none config/movement-controller | $(TOOLBIN)/kubectl apply -f -

undeploy_mc: manifests $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl config/default/kustomization.yaml
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/manager:${DOCKER_TAGNAME}
	$(TOOLBIN)/kustomize build --load_restrictor none config/movement-controller | $(TOOLBIN)/kubectl delete -f -

deploy_it: manifests $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl config/default/kustomization.yaml 
	$(TOOLBIN)/kubectl create namespace ${KUBE_NAMESPACE} || true
	$(TOOLBIN)/kubectl create namespace m4d-blueprints || true
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/manager:${DOCKER_TAGNAME}
	cd config/integration-tests && $(ABSTOOLBIN)/kustomize edit set image data-catalog-mock=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/data-catalog-mock:${DOCKER_TAGNAME}
	cd config/integration-tests && $(ABSTOOLBIN)/kustomize edit set image policycompiler=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/serverpolicycompiler-mock:${DOCKER_TAGNAME}
	$(TOOLBIN)/kustomize build --load_restrictor none config/integration-tests | $(TOOLBIN)/kubectl apply -f -

deploy_multi_it: manifests $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl config/default/kustomization.yaml
	$(TOOLBIN)/kubectl create namespace ${KUBE_NAMESPACE} || true
	$(TOOLBIN)/kubectl create namespace m4d-blueprints || true
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/manager:${DOCKER_TAGNAME}
	cd config/integration-tests && $(ABSTOOLBIN)/kustomize edit set image data-catalog-mock=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/data-catalog-mock:${DOCKER_TAGNAME}
	cd config/integration-tests && $(ABSTOOLBIN)/kustomize edit set image policycompiler=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/serverpolicycompiler-mock:${DOCKER_TAGNAME}
	$(TOOLBIN)/kustomize build --load_restrictor none config/multi-integration-tests | $(TOOLBIN)/kubectl apply -f -

undeploy_it: config/default/kustomization.yaml
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/manager:${DOCKER_TAGNAME}
	$(TOOLBIN)/kustomize build --load_restrictor none config/integration-tests | $(TOOLBIN)/kubectl delete -f -

deploy_control_plane_security: $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl config/default/kustomization.yaml
	$(TOOLBIN)/kubectl create namespace m4d-system || true
	$(TOOLBIN)/kubectl label namespaces m4d-system istio-injection=enabled --overwrite || true
	$(TOOLBIN)/kustomize build config/control-plane-security | $(TOOLBIN)/kubectl apply -f -

undeploy_control_plane_security: $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl config/default/kustomization.yaml
	$(TOOLBIN)/kubectl label namespace m4d-system istio-injection-
	$(TOOLBIN)/kustomize build --load_restrictor none config/control-plane-security | $(TOOLBIN)/kubectl delete -f -

.PHONY: wait_for_manager
wait_for_manager: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl wait --for=condition=available -n ${KUBE_NAMESPACE} deployment/m4d-controller-manager --timeout=120s

.PHONY: run-integration-tests
run-integration-tests: export DOCKER_HOSTNAME?=localhost:5000
run-integration-tests: export DOCKER_NAMESPACE?=m4d-system
run-integration-tests: wait_for_manager
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... -v -run TestMotionAPIs -count 1
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... -v -run TestAPIs -count 1

.PHONY: main.deps
main.deps: generate fmt vet manifests

DEBUG := ./debug.out

$(DEBUG): main.go */*/*.go
	go build -o $(DEBUG) main.go
