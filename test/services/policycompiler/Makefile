ROOT_DIR := ../../../
include $(ROOT_DIR)/Makefile.env

DOCKER_NAME ?= serverpolicycompiler-mock
include $(ROOT_DIR)/hack/make-rules/docker.mk

.PHONY: docker-build
docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o serverpolicycompiler serverpolicycompiler.go
	docker build . -t $(IMG)
	rm serverpolicycompiler
