ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk
DOCKER_NAME ?= datauserserver
DOCKER_TAGNAME ?= latest
DOCKER_FILE = Dockerfile
DOCKER_CONTEXT = .
KUBE_NAMESPACE ?= fybrik-system
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/verify.mk


# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: fmt vet
	go run ./main.go

.PHONY: test
test: build
	./datauserserver&
	sleep 10
	go test $(TEST_OPTIONS) ./...

# Build as linux binary
.PHONY: build
build: vet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o datauserserver main.go

# Overwrite docker-build from docker.mk
.PHONY: docker-build
docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o datauserserver main.go
	docker build . -t ${IMG} -f Dockerfile
	rm datauserserver

