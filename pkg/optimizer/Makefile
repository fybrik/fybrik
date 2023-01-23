ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env

DOCKER_NAME ?= optimizer
KUBE_NAMESPACE ?= fybrik-system
CONTROLLER_NAMESPACE ?= ${KUBE_NAMESPACE}

include $(ROOT_DIR)/hack/make-rules/tools.mk
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/verify.mk

# Overwrite docker-build from docker.mk
.PHONY: docker-build
docker-build: 	
	mkdir -p solver-tools
	cp $(TOOLS_DIR)/install_or_tools.sh $(TOOLS_DIR)/common.sh $(TOOLS_DIR)/requirements.sh solver-tools
	docker build . -t ${IMG} -f Dockerfile
	rm -rf solver-tools
