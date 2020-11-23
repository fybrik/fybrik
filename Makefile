include Makefile.env
export DOCKER_TAGNAME ?= latest

.PHONY: license
license: $(TOOLBIN)/license_finder
	$(call license_go,.)
	$(call license_python,secret-provider)

.PHONY: build
build:
	$(MAKE) -C pkg/policy-compiler build
	$(MAKE) -C manager manager

.PHONY: test
test:
	$(MAKE) -C pkg/policy-compiler test
	$(MAKE) -C manager test

.PHONY: cluster-prepare
cluster-prepare:
	$(MAKE) -C third_party/cert-manager deploy
	$(MAKE) -C third_party/registry deploy
	$(MAKE) -C third_party/vault deploy

.PHONY: cluster-prepare-wait
cluster-prepare-wait:
	$(MAKE) -C third_party/cert-manager deploy-wait
	$(MAKE) -C third_party/vault deploy-wait

.PHONY: install
install:
	$(MAKE) -C manager install

.PHONY: deploy
deploy:
	$(MAKE) -C secret-provider deploy
	$(MAKE) -C manager deploy
	$(MAKE) -C connectors deploy

# Deploys the manager using local images
.PHONY: deploy-local
deploy-local:
	$(MAKE) -C manager deploy-local

.PHONY: undeploy
undeploy:
	$(MAKE) -C secret-provider undeploy
	$(MAKE) -C manager undeploy
	$(MAKE) -C connectors undeploy

.PHONY: docker
docker:
	$(MAKE) -C manager docker-all
	$(MAKE) -C secret-provider docker-all
	$(MAKE) -C connectors docker-all
	$(MAKE) -C test/dummy-mover docker-all

# Build only the docker images needed for integration testing
.PHONY: docker-minimal-it
docker-minimal-it:
	$(MAKE) -C manager docker-all
	$(MAKE) -C secret-provider docker-all
	$(MAKE) -C test/dummy-mover docker-all
	$(MAKE) -C test/services docker

.PHONY: docker-build
docker-build:
	$(MAKE) -C test/dummy-mover docker-all
	$(MAKE) -C manager docker-build
	$(MAKE) -C secret-provider docker-build
	$(MAKE) -C connectors docker-build
	$(MAKE) -C test/dummy-mover docker-build-dummy-mover

.PHONY: docker-push
docker-push:
	$(MAKE) -C build docker-push-all
	$(MAKE) -C manager docker-push
	$(MAKE) -C secret-provider docker-push
	$(MAKE) -C connectors docker-push
	$(MAKE) -C test/dummy-mover docker-push-dummy-mover

# Build docker images locally. (Can be used in combination with minikube and deploy-local)
.PHONY: docker-build-local
docker-build-local:
	$(MAKE) -C manager docker-build-local
	$(MAKE) -C secret-provider docker-build-local

.PHONY: helm
helm:
	$(MAKE) -C modules helm

DOCKER_PUBLIC_HOSTNAME ?= ghcr.io
DOCKER_PUBLIC_NAMESPACE ?= the-mesh-for-data
DOCKER_PUBLIC_NAMES := \
	manager \
	secret-provider \
	egr-connector \
	dummy-mover \
	opa-connector \
	vault-connector
 
define do-docker-retag-and-push-public
	for name in ${DOCKER_PUBLIC_NAMES}; do \
		docker tag ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/$$name:${DOCKER_TAGNAME} ${DOCKER_PUBLIC_HOSTNAME}/${DOCKER_PUBLIC_NAMESPACE}/$$name:$1; \
	done
	DOCKER_HOSTNAME=${DOCKER_PUBLIC_HOSTNAME} DOCKER_NAMESPACE=${DOCKER_PUBLIC_NAMESPACE} DOCKER_TAGNAME=$1 $(MAKE) docker-push
endef

.PHONY: docker-retag-and-push-public
docker-retag-and-push-public:
	$(call do-docker-retag-and-push-public,latest)
ifneq (${TRAVIS_TAG},)
	$(call do-docker-retag-and-push-public,${TRAVIS_TAG})
endif

include hack/make-rules/tools.mk
include hack/make-rules/verify.mk
include hack/make-rules/cluster.mk
include hack/make-rules/helm.mk
