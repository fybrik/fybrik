ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

.PHONY: all
all: source-build

DOCKER_NAME ?= manager
KUBE_NAMESPACE ?= fybrik-system
CONTROLLER_NAMESPACE ?= ${KUBE_NAMESPACE}
PRETTY_LOGGING ?= true
LOGGING_VERBOSITY?= -1

include $(ROOT_DIR)/hack/make-rules/tools.mk
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/verify.mk
include $(ROOT_DIR)/hack/make-rules/version.mk

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run:
	go run ./main.go --enable-all-controllers --metrics-bind-addr=0 --health-probe-addr=127.0.0.1:8088

.PHONY: source-build
source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build $(LDFLAGS) -o manager main.go

# Overwrite docker-build from docker.mk
.PHONY: docker-build
docker-build: source-build
	docker build . -t ${IMG} -f Dockerfile --build-arg tag=${BASE_IMAGE_TAG}
	rm manager

.PHONY: wait_for_manager
wait_for_manager: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl wait --for=condition=available -n ${CONTROLLER_NAMESPACE} deployment/manager --timeout=120s

.PHONY: prep-test
prep-test: export DOCKER_HOSTNAME?=localhost:5000
prep-test: export DOCKER_NAMESPACE?=fybrik-system
prep-test: export DOCKER_TAGNAME?=master
prep-test: export USE_MOCKUP_CONNECTOR?=true
prep-test: wait_for_manager

.PHONY: run-integration-tests
run-integration-tests: prep-test
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... $(TEST_OPTIONS) -run TestAPIs -count 1

.PHONY: run-notebook-readflow-tests
run-notebook-readflow-tests: prep-test
	cd testdata/notebook/read-flow && ./setup.sh
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... $(TEST_OPTIONS) -run TestS3NotebookReadFlow -count 1

.PHONY: run-network-policy-readflow-tests
run-network-policy-readflow-tests: prep-test
	cd testdata/notebook/read-flow && ./setup.sh
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... $(TEST_OPTIONS) -run TestNetworkPolicyReadFlow -count 1

.PHONY: run-notebook-writeflow-tests
run-notebook-writeflow-tests: prep-test
	cd testdata/notebook/write-flow && ./setup.sh
	NO_SIMULATED_PROGRESS=true USE_EXISTING_CONTROLLER=true USE_EXISTING_CLUSTER=true go test ./... $(TEST_OPTIONS) -run TestS3NotebookWriteFlow -count 1

DEBUG := ./debug.out

$(DEBUG): main.go */*/*.go
	go build -o $(DEBUG) main.go
