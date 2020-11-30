ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
#include Makefile.env

include $(ROOT_DIR)/hack/make-rules/tools.mk
DOCKER_NAME ?= dummy-mover
DOCKER_TAGNAME ?= latest
DOCKER_FILE = Dockerfile.dummy-mover
DOCKER_CONTEXT = .
GO_OUTPUT_FILE = manager
KUBE_NAMESPACE ?= m4d-system
include $(ROOT_DIR)/hack/make-rules/docker.mk

.PHONY: docker-build
docker-build:
	docker build . -t $(IMG) -f Dockerfile.dummy-mover
