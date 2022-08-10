ROOT_DIR := ../..
DOCKER_NAME = katalog-connector

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk
include $(ROOT_DIR)/hack/make-rules/version.mk

.PHONY: all
all: docker-build docker-push

.PHONY: docker-build
docker-build: source-build
	docker build . -t ${IMG}
	rm -rf bin

.PHONY: source-build
source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build $(LDFLAGS) -o bin/katalog main.go

.PHONY: run
run:
	go run main.go run

.PHONY: test
test:
	go test -v ./...
