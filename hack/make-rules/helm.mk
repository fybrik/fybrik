# This script contains helm version 3.7 commands for pushing and pulling charts to OCI registry
# as described in https://github.com/helm/community/blob/main/hips/hip-0006.md
# To use it the following env vars should be defined:

# CHART_NAME the chart name as is appear in Chart.yaml
# HELM_RELEASE the helm release-name of the chart
# HELM_TAG  the OCI reference tag (and also the chart version). Must be SemVer
# CHART_LOCAL_PATH path to the chart directory
# DOCKER_HOSTNAME the docker registry hostname
# DOCKER_NAMESPACE docker registry namespace
# DOCKER_USERNAME docker registry  username


HELM_VALUES ?= \
	--set hello=world1

TEMP := /tmp
CHART_LOCAL_PATH ?= ../${DOCKER_NAME}
CHART_NAME ?= ${DOCKER_NAME}
HELM_RELEASE ?= rel1-${DOCKER_NAME}
HELM_TAG ?= 0.0.0

CHART_REGISTRY_PATH := oci://${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}

export HELM_EXPERIMENTAL_OCI=1
export GODEBUG=x509ignoreCN=0

.PHONY: helm-login
helm-login: $(TOOLBIN)/helm
ifneq (${DOCKER_PASSWORD},)
	$(ABSTOOLBIN)/helm registry login -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}" ${DOCKER_HOSTNAME}
endif

.PHONY: helm-verify
helm-verify: $(TOOLBIN)/helm
	$(ABSTOOLBIN)/helm lint ${CHART_LOCAL_PATH}
	$(ABSTOOLBIN)/helm install --dry-run ${HELM_RELEASE} ${CHART_LOCAL_PATH} ${HELM_VALUES}

.PHONY: helm-uninstall
helm-uninstall: $(TOOLBIN)/helm
	$(ABSTOOLBIN)/helm uninstall ${HELM_RELEASE} || true

.PHONY: helm-install
helm-install: $(TOOLBIN)/helm
	$(ABSTOOLBIN)/helm install ${HELM_RELEASE} ${CHART_LOCAL_PATH} ${HELM_VALUES}


# example for helm chart push:
# helm package fybrik-template -d /tmp/ --version 0.7.0
# helm push /tmp/fybrik-template-0.7.0.tgz oci://localhost:5000/fybrik-system/
.PHONY: helm-chart-push
helm-chart-push: helm-login
	$(ABSTOOLBIN)/helm package ${CHART_LOCAL_PATH} --version=${HELM_TAG} --destination=${TEMP}
	$(ABSTOOLBIN)/helm push ${TEMP}/${CHART_NAME}-${HELM_TAG}.tgz ${CHART_REGISTRY_PATH}
	rm -rf ${TEMP}/${CHART_NAME}-${HELM_TAG}.tgz

.PHONY: helm-chart-pull
helm-chart-pull: helm-login
	$(ABSTOOLBIN)/helm pull ${CHART_REGISTRY_PATH}/${CHART_NAME} --version ${HELM_TAG}

.PHONY: helm-list
helm-list: $(TOOLBIN)/helm
	$(ABSTOOLBIN)/helm list

.PHONY: helm-chart-install
helm-chart-install: $(TOOLBIN)/helm
	$(ABSTOOLBIN)/helm install ${HELM_RELEASE} ${CHART_REGISTRY_PATH}/${CHART_NAME} --version ${HELM_TAG} ${HELM_VALUES}
	$(ABSTOOLBIN)/helm list

.PHONY: helm-template
helm-template: $(TOOLBIN)/helm
	$(ABSTOOLBIN)/helm template ${HELM_RELEASE} ${CHART_REGISTRY_PATH} --version ${HELM_TAG} ${HELM_VALUES}

.PHONY: helm-debug
helm-debug: $(TOOLBIN)/helm
	$(ABSTOOLBIN)/helm template ${HELM_RELEASE} ${CHART_REGISTRY_PATH} ${HELM_VALUES} --version ${HELM_TAG} --debug

.PHONY: helm-actions
helm-actions: $(TOOLBIN)/helm
	$(ABSTOOLBIN)/helm show values --version ${HELM_TAG}  ${CHART_REGISTRY_PATH} | yq -y -r .actions

.PHONY: helm-all
helm-all: helm-verify helm-chart-push helm-chart-pull helm-uninstall helm-chart-install

