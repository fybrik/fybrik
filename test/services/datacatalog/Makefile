ROOT_DIR := ../../../
include $(ROOT_DIR)/Makefile.env

DOCKER_NAME ?= data-catalog-mock
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk

.PHONY: docker-build
docker-build: source-build
	docker build . -t ${IMG} -f Dockerfile --build-arg tag=${BASE_IMAGE_TAG}
	rm datacatalog

.PHONY: source-build
source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o datacatalog datacatalogstub.go
