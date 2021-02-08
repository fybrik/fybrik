.PHONY: all
all: helm-all

ROOT_DIR := ../..
DOCKER_NAME := m4d-implicit-copy-batch
DOCKER_TAGNAME := 0.1.0
HELM_VALUES := --values values.yaml.sample

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/helm.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk

helm-crd-hook: crd-batch
.PHONY: crd-batch
crd-batch:
	$(ABSTOOLBIN)/kubectl apply -f $(ROOT_DIR)/manager/config/crd/bases/motion.m4d.ibm.com_batchtransfers.yaml
