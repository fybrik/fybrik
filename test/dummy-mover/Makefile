ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env

include $(ROOT_DIR)/hack/make-rules/tools.mk
DOCKER_NAME ?= dummy-mover
include $(ROOT_DIR)/hack/make-rules/docker.mk

.PHONY: docker-build
docker-build:
	docker build . -t $(IMG) -f Dockerfile.dummy-mover
