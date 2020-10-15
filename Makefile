include Makefile.env

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

.PHONY: e2e
e2e:
	# TODO(roee88): temporarily removed until can be set against local registry
	# $(MAKE) -C pkg/helm test
	$(MAKE) -C manager e2e

# Splitting cluster-prepare into deploy and deploy-wait
# Allows to differentiate between YAMLs to be deployed
# and waiting for the state to be ready

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
	$(MAKE) -C build all
	$(MAKE) -C manager docker-all
	$(MAKE) -C secret-provider docker-all
	$(MAKE) -C connectors docker-all

# Build only the docker images needed for integration testing
.PHONY: docker-minimal-it
docker-minimal-it:
	$(MAKE) -C build docker-dummy-mover
	$(MAKE) -C manager docker-all
	$(MAKE) -C secret-provider docker-all
	$(MAKE) -C test/services docker

.PHONY: docker-build
docker-build:
	$(MAKE) -C build docker-build-all
	$(MAKE) -C manager docker-build
	$(MAKE) -C secret-provider docker-build
	$(MAKE) -C connectors docker-build

.PHONY: docker-push
docker-push:
	$(MAKE) -C build docker-push-all
	$(MAKE) -C manager docker-push
	$(MAKE) -C secret-provider docker-push
	$(MAKE) -C connectors docker-push

# Build docker images locally. (Can be used in combination with minikube and deploy-local)
.PHONY: docker-build-local
docker-build-local:
	$(MAKE) -C manager docker-build-local
	$(MAKE) -C secret-provider docker-build-local

.PHONY: helm
helm:
	$(MAKE) -C modules helm

.PHONY: docker-retag-images
docker-retag-images:
	docker tag ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/manager:latest ghcr.io/the-mesh-for-data/manager:latest
	docker tag ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/secret-provider:latest ghcr.io/the-mesh-for-data/secret-provider:latest
	docker tag ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/egr-connector:latest ghcr.io/the-mesh-for-data/egr-connector:latest
	docker tag ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/movement-controller:latest ghcr.io/the-mesh-for-data/movement-controller:latest
	docker tag ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/dummy-mover:latest ghcr.io/the-mesh-for-data/dummy-mover:latest
	docker tag ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/opa-connector:latest ghcr.io/the-mesh-for-data/opa-connector:latest
	docker tag ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/vault-connector:latest ghcr.io/the-mesh-for-data/vault-connector:latest
	docker tag ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/serverpolicycompiler-mock:latest ghcr.io/the-mesh-for-data/serverpolicycompiler-mock:latest
	docker tag ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/data-catalog-mock:latest ghcr.io/the-mesh-for-data/data-catalog-mock:latest

.PHONY: docker-push-public
docker-push-public:
	docker push ghcr.io/the-mesh-for-data/manager:latest
	docker push ghcr.io/the-mesh-for-data/secret-provider:latest
	docker push ghcr.io/the-mesh-for-data/egr-connector:latest
	docker push ghcr.io/the-mesh-for-data/movement-controller:latest
	docker push ghcr.io/the-mesh-for-data/dummy-mover:latest
	docker push ghcr.io/the-mesh-for-data/opa-connector:latest
	docker push ghcr.io/the-mesh-for-data/vault-connector:latest
	docker push ghcr.io/the-mesh-for-data/serverpolicycompiler-mock:latest
	docker push ghcr.io/the-mesh-for-data/data-catalog-mock:latest

include .mk/ibmcloud.mk
include .mk/tools.mk
include .mk/verify.mk
include .mk/cluster.mk
include .mk/helm.mk
