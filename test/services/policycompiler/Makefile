ROOT_DIR := ../../../
include $(ROOT_DIR)/Makefile.env
#include Makefile.env

include $(ROOT_DIR)/hack/make-rules/tools.mk
DOCKER_TAGNAME ?= latest
include $(ROOT_DIR)/hack/make-rules/docker.mk

IMG_SPCM ?= ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/serverpolicycompiler-mock:${DOCKER_TAGNAME}

.PHONY: all
all: docker-build-all docker-push-all

.PHONY: docker-build-all
docker-build-all: serverpolicycompiler-mock

.PHONY: docker-push-all
docker-push-all:
	$(MAKE) docker-push IMG=$(IMG_SPCM)

.PHONY: serverpolicycompiler-mock
serverpolicycompiler-mock:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o serverpolicycompiler serverpolicycompiler.go
	docker build . -t $(IMG_SPCM) -f Dockerfile
	rm serverpolicycompiler
