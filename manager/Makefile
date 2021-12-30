ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

.PHONY: all
all: manager

include $(ROOT_DIR)/hack/make-rules/tools.mk
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/verify.mk

DOCKER_NAME ?= manager
GO_OUTPUT_FILE = manager
KUBE_NAMESPACE ?= fybrik-system
CONTROLLER_NAMESPACE ?= ${KUBE_NAMESPACE}

# Build manager binary
manager:
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run:
	go run ./main.go --enable-all-controllers --metrics-bind-addr=0

source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o manager main.go

# Overwrite docker-build from docker.mk
docker-build: source-build
	docker build . -t ${IMG} -f Dockerfile
	rm manager

.PHONY: wait_for_manager
wait_for_manager: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl wait --for=condition=available -n ${CONTROLLER_NAMESPACE} deployment/manager --timeout=120s

.PHONY: run-integration-tests
run-integration-tests: export DOCKER_HOSTNAME?=localhost:5000
run-integration-tests: export DOCKER_NAMESPACE?=fybrik-system
run-integration-tests: export DOCKER_TAGNAME?=master
run-integration-tests: export USE_MOCKUP_CONNECTOR?=true
run-integration-tests: wait_for_manager
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... -v -run TestAPIs -count 1

.PHONY: run-notebook-tests
run-notebook-tests: export DOCKER_HOSTNAME?=localhost:5000
run-notebook-tests: export DOCKER_NAMESPACE?=fybrik-system
run-notebook-tests: export DOCKER_TAGNAME?=master
run-notebook-tests: export USE_MOCKUP_CONNECTOR?=true
run-notebook-tests: wait_for_manager
	cd testdata/notebook && ./setup.sh
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... -v -run TestS3Notebook -count 1

DEBUG := ./debug.out

$(DEBUG): main.go */*/*.go
	go build -o $(DEBUG) main.go
