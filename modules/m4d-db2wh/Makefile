.PHONY: all
all: helm-all

ROOT_DIR := ../..
DOCKER_NAME := m4d-db2wh
DOCKER_TAGNAME := 0.1.0
HELM_VALUES := \
	--values values.yaml.sample \
	--set image.name=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/mover-dummy:${DOCKER_TAGNAME}

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/.mk/docker.mk
include $(ROOT_DIR)/.mk/helm.mk
include $(ROOT_DIR)/.mk/tools.mk
include $(ROOT_DIR)/.mk/ibmcloud.mk

helm-crd-hook: crd-batch
.PHONY: crd-batch
crd-batch:
	$(ABSTOOLBIN)/kubectl apply -f $(ROOT_DIR)/manager/config/crd/bases/motion.m4d.ibm.com_batchtransfers.yaml
