ROOT_DIR := ../../../
include $(ROOT_DIR)/Makefile.env

DOCKER_NAME ?= serverpolicymanager-mock
include $(ROOT_DIR)/hack/make-rules/docker.mk

.PHONY: docker-build
docker-build: source-build
	docker build . -t $(IMG)
	rm serverpolicymanager

.PHONY: source-build
source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o serverpolicymanager serverpolicymanager.go
