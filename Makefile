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

.PHONY: cluster-prepare
cluster-prepare:
	$(MAKE) -C third_party/cert-manager deploy
	$(MAKE) -C third_party/registry deploy
	$(MAKE) -C third_party/vault deploy

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

include .mk/ibmcloud.mk
include .mk/tools.mk
include .mk/verify.mk
include .mk/cluster.mk
include .mk/helm.mk
