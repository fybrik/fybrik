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
	mkdir solver-tools
	mkdir solver-tools/bin
	mkdir solver-tools/lib
	cp ${ABSTOOLBIN}/fzn-or-tools solver-tools/bin/solver
	cp ${ABSTOOLBIN}/../lib/libfz.so solver-tools/lib/
	cp ${ABSTOOLBIN}/../lib/libortools.so solver-tools/lib/
	docker build . -t ${IMG} -f Dockerfile
	rm -rf solver-tools
