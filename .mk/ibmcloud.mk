DOCKER_PASSWORD ?= $(shell echo "${DOCKER_PASSWORD_ENCODED}" | base64 -d)
DOCKER_HOSTNAME ?= registry-1.docker.io
DOCKER_NAMESPACE ?= m4d
DOCKER_NAME ?= m4d
DOCKER_TAGNAME ?= latest
IMG ?= ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/${DOCKER_NAME}:${DOCKER_TAGNAME}

IBMCLOUD:=/usr/local/bin/ibmcloud

.PHONY: ibmcloud-login
ibmcloud-login: $(IBMCLOUD)
	$(IBMCLOUD) login --apikey ${DOCKER_PASSWORD}

.PHONY: ibmcloud-image-list
ibmcloud-image-list: $(IBMCLOUD) ibmcloud-login
	$(IBMCLOUD) cr image-list --restrict ${DOCKER_NAMESPACE}

.PHONY: ibmcloud-image-rm
ibmcloud-image-rm: $(IBMCLOUD) ibmcloud-login
	$(IBMCLOUD) cr image-rm ${IMG} || true
