include Makefile.env
export DOCKER_TAGNAME ?= latest
export KUBE_NAMESPACE ?= m4d-system

.PHONY: license
license: $(TOOLBIN)/license_finder
	$(call license_go,.)

.PHONY: docker-mirror-read
docker-mirror-read:
	$(TOOLS_DIR)/docker_mirror.sh $(TOOLS_DIR)/docker_mirror.conf

.PHONY: deploy
deploy: export VALUES_FILE?=charts/m4d/values.yaml
deploy: $(TOOLBIN)/kubectl $(TOOLBIN)/helm
	$(TOOLBIN)/kubectl create namespace $(KUBE_NAMESPACE) || true
	$(TOOLBIN)/helm install m4d-crd charts/m4d-crd  \
               --namespace $(KUBE_NAMESPACE) --wait --timeout 120s
	$(TOOLBIN)/helm install m4d charts/m4d --values $(VALUES_FILE) \
               --namespace $(KUBE_NAMESPACE) --wait --timeout 120s

.PHONY: test
test:
	$(MAKE) -C manager pre-test
	go test -v ./...
	# The tests for connectors/egeria are dropped because there are none

.PHONY: run-integration-tests
run-integration-tests: export DOCKER_HOSTNAME?=localhost:5000
run-integration-tests: export DOCKER_NAMESPACE?=m4d-system
run-integration-tests: export VALUES_FILE=charts/m4d/integration-tests.values.yaml
run-integration-tests:
	$(MAKE) kind
	$(MAKE) cluster-prepare
	$(MAKE) docker-build docker-push
	$(MAKE) -C test/services docker-build docker-push
	$(MAKE) cluster-prepare-wait
	$(MAKE) deploy
	$(MAKE) configure-vault
	$(MAKE) -C modules helm
	$(MAKE) -C modules helm-uninstall # Uninstalls the deployed tests from previous command
	$(MAKE) -C pkg/helm test
	$(MAKE) -C manager run-integration-tests
	$(MAKE) -C modules test

.PHONY: run-deploy-tests
run-deploy-tests:
	$(MAKE) kind
	$(MAKE) cluster-prepare
	kubectl config set-context --current --namespace=$(KUBE_NAMESPACE)
	$(MAKE) -C third_party/opa deploy
	kubectl apply -f ./manager/config/prod/deployment_configmap.yaml
	$(MAKE) -C connectors deploy
	kubectl get pod --all-namespaces
	kubectl wait --for=condition=ready pod --all-namespaces --all --timeout=120s
	$(MAKE) configure-vault

.PHONY: cluster-prepare
cluster-prepare:
	$(MAKE) -C third_party/cert-manager deploy
	$(MAKE) -C third_party/vault deploy
	$(MAKE) -C third_party/datashim deploy

.PHONY: cluster-prepare-wait
cluster-prepare-wait:
	$(MAKE) -C third_party/datashim deploy-wait
	$(MAKE) -C third_party/vault deploy-wait

# Build only the docker images needed for integration testing
.PHONY: docker-minimal-it
docker-minimal-it:
	$(MAKE) -C manager docker-build docker-push
	$(MAKE) -C test/dummy-mover docker-build docker-push
	$(MAKE) -C test/services docker-build docker-push

.PHONY: docker-build
docker-build:
	$(MAKE) -C manager docker-build
	$(MAKE) -C connectors docker-build
	$(MAKE) -C test/dummy-mover docker-build

.PHONY: docker-push
docker-push:
	$(MAKE) -C manager docker-push
	$(MAKE) -C connectors docker-push
	$(MAKE) -C test/dummy-mover docker-push

DOCKER_PUBLIC_HOSTNAME ?= ghcr.io
DOCKER_PUBLIC_NAMESPACE ?= mesh-for-data
DOCKER_PUBLIC_TAGNAME ?= latest

DOCKER_PUBLIC_NAMES := \
	manager \
	dummy-mover \
	egr-connector \
	katalog-connector \
	opa-connector 

define do-docker-retag-and-push-public
	for name in ${DOCKER_PUBLIC_NAMES}; do \
		docker tag ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/$$name:${DOCKER_TAGNAME} ${DOCKER_PUBLIC_HOSTNAME}/${DOCKER_PUBLIC_NAMESPACE}/$$name:${DOCKER_PUBLIC_TAGNAME}; \
	done
	DOCKER_HOSTNAME=${DOCKER_PUBLIC_HOSTNAME} DOCKER_NAMESPACE=${DOCKER_PUBLIC_NAMESPACE} DOCKER_TAGNAME=${DOCKER_PUBLIC_TAGNAME} $(MAKE) docker-push
endef

.PHONY: docker-retag-and-push-public
docker-retag-and-push-public:
	$(call do-docker-retag-and-push-public)

.PHONY: helm-push-public
helm-push-public:
	DOCKER_HOSTNAME=${DOCKER_PUBLIC_HOSTNAME} DOCKER_NAMESPACE=${DOCKER_PUBLIC_NAMESPACE} DOCKER_TAGNAME=${DOCKER_PUBLIC_TAGNAME} make -C modules helm-chart-push

.PHONY: save-images
save-images:
	docker save -o images.tar ${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/manager:${DOCKER_TAGNAME} \
		${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/dummy-mover:${DOCKER_TAGNAME} \
		${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/egr-connector:${DOCKER_TAGNAME} \
		${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/katalog-connector:${DOCKER_TAGNAME} \
		${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/opa-connector:${DOCKER_TAGNAME}

include hack/make-rules/tools.mk
include hack/make-rules/verify.mk
include hack/make-rules/cluster.mk
include hack/make-rules/vault.mk
