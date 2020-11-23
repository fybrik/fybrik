ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env
#include Makefile.env

include $(ROOT_DIR)/hack/make-rules/tools.mk
DOCKER_TAGNAME ?= latest
include $(ROOT_DIR)/hack/make-rules/docker.mk

IMG_MC ?= ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/movement-controller:${DOCKER_TAGNAME}
IMG_DM ?= ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/dummy-mover:${DOCKER_TAGNAME}

.PHONY: all
all: docker-build-all docker-push-all

.PHONY: docker-build-all
docker-build-all: docker-movement-controller docker-build-dummy-mover

.PHONY: docker-push-all
docker-push-all:
	$(MAKE) docker-push IMG=$(IMG_MC)
	$(MAKE) docker-push IMG=$(IMG_DM)

.PHONY: docker-movement-controller
docker-movement-controller:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o manager ../cmd/movement-controller/main.go
	docker build . -t $(IMG_MC) -f Dockerfile.movement-controller
	rm manager

.PHONY: docker-dummy-mover
docker-dummy-mover:
	$(MAKE) docker-build-dummy-mover
	$(MAKE) docker-push-dummy-mover

.PHONY: docker-build-dummy-mover
docker-build-dummy-mover:
	docker build . -t $(IMG_DM) -f Dockerfile.dummy-mover

.PHONY: docker-push-dummy-mover
docker-push-dummy-mover:
	$(MAKE) docker-push IMG=$(IMG_DM)
