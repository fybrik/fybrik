export DOCKER_USERNAME ?=
export DOCKER_PASSWORD ?=
export DOCKER_HOSTNAME ?= ghcr.io
export DOCKER_NAMESPACE ?= fybrik
export DOCKER_TAGNAME ?= latest

DOCKER_NAME ?= fybrik

IMG ?= ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/${DOCKER_NAME}:${DOCKER_TAGNAME}

.PHONY: docker-push
docker-push:
ifneq (${DOCKER_PASSWORD},)
	@docker login \
		--username ${DOCKER_USERNAME} \
		--password ${DOCKER_PASSWORD} ${DOCKER_HOSTNAME}
endif
	docker push ${IMG}

.PHONY: docker-rmi
docker-rmi:
	docker rmi ${IMG} || true
