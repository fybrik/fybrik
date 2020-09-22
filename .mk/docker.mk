DOCKER_PASSWORD ?= $(shell echo "${DOCKER_PASSWORD_ENCODED}" | base64 -d)
DOCKER_HOSTNAME ?= ghcr.io
DOCKER_NAMESPACE ?= the-mesh-for-data
DOCKER_NAME ?= m4d
DOCKER_TAGNAME ?= latest
DOCKER_CONTEXT ?= .
GO_INPUT_FILE ?= main.go
GO_OUTPUT_FILE ?= manager
IMG ?= ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/${DOCKER_NAME}:${DOCKER_TAGNAME}

export DOCKER_USERNAME
export DOCKER_PASSWORD
export DOCKER_HOSTNAME
export DOCKER_NAMESPACE
export DOCKER_TAGNAME

DOCKER_FILE ?= Dockerfile

.PHONY: docker-all
docker-all: docker-build docker-push

.PHONY: docker-build
docker-build:
	docker build $(DOCKER_CONTEXT) -t ${IMG} -f $(DOCKER_FILE)

.PHONY: docker-build-local
docker-build-local:
	make docker-build IMG=${DOCKER_NAME}:${DOCKER_TAGNAME}

.PHONY: docker-push
docker-push:
ifdef DOCKER_USERNAME
	@docker login \
		--username ${DOCKER_USERNAME} \
		--password ${DOCKER_PASSWORD} ${DOCKER_HOSTNAME}
endif
	docker push ${IMG}

.PHONY: docker-rmi
docker-rmi:
	docker rmi ${IMG} || true
