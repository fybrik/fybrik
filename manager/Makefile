ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

.PHONY: all
all: manager

DOCKER_NAME ?= manager
KUBE_NAMESPACE ?= fybrik-system
CONTROLLER_NAMESPACE ?= ${KUBE_NAMESPACE}

GIT_COMMIT := $(shell git rev-list -1 HEAD)

include $(ROOT_DIR)/hack/make-rules/tools.mk
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/verify.mk

# Build manager binary
.PHONY: manager
manager:
	go build -ldflags "-X main.gitCommit=$(GIT_COMMIT)" -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run:
	go run -ldflags "-X main.gitCommit=$(GIT_COMMIT)" ./main.go --enable-all-controllers --metrics-bind-addr=0

.PHONY: source-build
source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags "-X main.gitCommit=$(GIT_COMMIT)" -o manager main.go

# Overwrite docker-build from docker.mk
.PHONY: docker-build
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


.PHONY: run-notebook-readflow-tests
run-notebook-readflow-tests: export DOCKER_HOSTNAME?=localhost:5000
run-notebook-readflow-tests: export DOCKER_NAMESPACE?=fybrik-system
run-notebook-readflow-tests: export DOCKER_TAGNAME?=master
run-notebook-readflow-tests: export USE_MOCKUP_CONNECTOR?=true
run-notebook-readflow-tests: wait_for_manager
	cd testdata/notebook/read-flow && ./setup.sh
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... -v -run TestS3NotebookReadFlow -count 1

.PHONY: run-notebook-writeflow-tests
run-notebook-writeflow-tests: export DOCKER_HOSTNAME?=localhost:5000
run-notebook-writeflow-tests: export DOCKER_NAMESPACE?=fybrik-system
run-notebook-writeflow-tests: export DOCKER_TAGNAME?=master
run-notebook-writeflow-tests: export USE_MOCKUP_CONNECTOR?=true
run-notebook-writeflow-tests: wait_for_manager
	cd testdata/notebook/write-flow && ./setup.sh
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... -v -run TestS3NotebookWriteFlow -count 1

DEBUG := ./debug.out

$(DEBUG): main.go */*/*.go
	go build -o $(DEBUG) main.go
