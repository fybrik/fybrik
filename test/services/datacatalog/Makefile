ROOT_DIR := ../../../
include $(ROOT_DIR)/Makefile.env
#include Makefile.env

include $(ROOT_DIR)/hack/make-rules/tools.mk
DOCKER_TAGNAME ?= latest
include $(ROOT_DIR)/hack/make-rules/docker.mk

IMG_DCM ?= ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/data-catalog-mock:${DOCKER_TAGNAME}

.PHONY: all
all: docker-build-all docker-push-all

.PHONY: docker-build-all
docker-build-all: data-catalog-mock

.PHONY: docker-push-all
docker-push-all:
	$(MAKE) docker-push IMG=$(IMG_DCM)

.PHONY: data-catalog-mock
data-catalog-mock:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o datacatalog datacatalogstub.go
	docker build . -t $(IMG_DCM) -f Dockerfile
	rm datacatalog
