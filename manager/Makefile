ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env
#include Makefile.env

.PHONY: all
all: manager

include $(ROOT_DIR)/.mk/tools.mk
DOCKER_NAME ?= manager
DOCKER_TAGNAME ?= latest
DOCKER_FILE = Dockerfile
DOCKER_CONTEXT = .
GO_OUTPUT_FILE = manager
KUBE_NAMESPACE ?= m4d-system
include $(ROOT_DIR)/.mk/docker.mk
include $(ROOT_DIR)/.mk/verify.mk
include $(ROOT_DIR)/.mk/operator.mk

# Build manager binary as linux binary
build-linux: generate vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Overwrite docker-build from docker.mk
docker-build: generate
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o manager main.go
	docker build . -t ${IMG} -f Dockerfile
	rm manager

# Deploy only movement-controller in the configured Kubernetes cluster in ~/.kube/config
deploy_mc: manifests $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl create namespace ${KUBE_NAMESPACE} || true
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/movement-controller:${DOCKER_TAGNAME}
	$(TOOLBIN)/kustomize build --load_restrictor none config/movement-controller | $(TOOLBIN)/kubectl apply -f -

undeploy_mc: manifests $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/movement-controller:${DOCKER_TAGNAME}
	$(TOOLBIN)/kustomize build --load_restrictor none config/movement-controller | $(TOOLBIN)/kubectl delete -f -

deploy_it: docker-secret
	$(TOOLBIN)/kubectl create namespace ${KUBE_NAMESPACE} || true
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/manager:${DOCKER_TAGNAME}
	cd config/integration-tests && $(ABSTOOLBIN)/kustomize edit set image data-catalog-mock=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/data-catalog-mock:${DOCKER_TAGNAME}
	cd config/integration-tests && $(ABSTOOLBIN)/kustomize edit set image policycompiler=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/serverpolicycompiler-mock:${DOCKER_TAGNAME}
	$(TOOLBIN)/kustomize build --load_restrictor none config/integration-tests | $(TOOLBIN)/kubectl apply -f -

undeploy_it:
	cd config/manager && $(ABSTOOLBIN)/kustomize edit set image controller=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/manager:${DOCKER_TAGNAME}
	$(TOOLBIN)/kustomize build --load_restrictor none config/integration-tests | $(TOOLBIN)/kubectl delete -f -

.PHONY: wait_for_manager
wait_for_manager:
	$(TOOLBIN)/kubectl wait --for=condition=available -n ${KUBE_NAMESPACE} deployment/m4d-controller-manager --timeout=120s

.PHONY: run-integration-tests
run-integration-tests: wait_for_manager
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... -v -run TestMotionAPIs -count 1
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... -v -run TestAPIs -count 1

.PHONY: main.deps
main.deps: generate fmt vet manifests

# Download CA certificates from manager image and add the CA certificate that was generated for the test.
# This file is then uploaded to K8s as a secret and injected into the manager during deployment.
# This way the image of the manger is not altered and still only contains the CA certificates from the base image.
.PHONY: build-ca-certs
build-ca-certs: $(TOOLBIN)/kubectl
	rm $(ROOT_DIR)/hack/registry/ca-certificates.crt || true
	# Hack to fetch CA certificates of base image
	bash -c "docker cp $$(docker create --rm $(DOCKER_HOSTNAME)/$(DOCKER_NAMESPACE)/$(DOCKER_NAME):$(DOCKER_TAGNAME)):/etc/ssl/certs/ca-certificates.crt $(ROOT_DIR)/hack/registry/ca-certificates.crt"
	cat $(ROOT_DIR)/hack/registry/ca.crt >> $(ROOT_DIR)/hack/registry/ca-certificates.crt

.PHONY: docker-secret
docker-secret: $(TOOLBIN)/kubectl
	make build-ca-certs
	$(ABSTOOLBIN)/kubectl create secret generic docker-credentials \
		--from-literal=hostname=$(DOCKER_HOSTNAME) \
		--from-literal=namespace=$(DOCKER_NAMESPACE) \
		--from-literal=tagname=$(DOCKER_TAGNAME) \
		--from-literal=username=$(DOCKER_USERNAME) \
		--from-literal=password=$(DOCKER_PASSWORD) \
		--namespace ${KUBE_NAMESPACE} \
		|| true
	$(ABSTOOLBIN)/kubectl create secret generic registry-ca-cert \
		--from-file=../hack/registry/ca-certificates.crt \
		--namespace ${KUBE_NAMESPACE} \
		|| true

DEBUG := ./debug.out

$(DEBUG): main.go */*/*.go
	go build -o $(DEBUG) main.go
