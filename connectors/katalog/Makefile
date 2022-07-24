ROOT_DIR := ../..
DOCKER_NAME = katalog-connector

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk

GIT_COMMIT := $(shell git rev-list -1 HEAD)

.PHONY: all
all: docker-build docker-push

.PHONY: docker-build
docker-build: source-build
	docker build . -t ${IMG}
	rm -rf bin

.PHONY: source-build
source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags "-X main.gitCommit=$(GIT_COMMIT)" -o bin/katalog main.go

.PHONY: run
run:
	go run main.go run

.PHONY: test
test:
	go test -v ./...
