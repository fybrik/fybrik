ROOT_DIR := ../../../..
DOCKER_NAME = manager-client

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk

KUBE_NAMESPACE ?= m4d-system

docker-all: docker-build docker-push 

.PHONY: build
build: $(TOOLBIN)/protoc $(TOOLBIN)/protoc-gen-go
	rm -f protobuf/*
	PATH=$(ABSTOOLBIN) protoc -I protos/ protos/*.proto --go_opt=paths=source_relative --go_out=plugins=grpc:$(ROOT_DIR)/pkg/connectors/protobuf

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o manager-client manager-client2.go
	docker build . -t $(IMG)
	rm manager-client

-include ${ROOT_DIR}/.env
.EXPORT_ALL_VARIABLES:


.PHONY: deploy
deploy: $(TOOLBIN)/kustomize
	./deploy.sh

.PHONY: undeploy
undeploy:
	./deploy.sh undeploy
