ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env

KUBE_NAMESPACE ?= m4d-system

.PHONY: charts-secret-provider
charts-secret-provider:
	cd base && kustomize edit set image secret-provider=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/secret-provider:${DOCKER_TAGNAME}
	cd base && kustomize edit set namespace ${KUBE_NAMESPACE}
	kustomize build base > $(ROOT_DIR)/charts/m4d/templates/secret-provider.yaml
