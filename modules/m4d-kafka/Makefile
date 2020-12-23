.PHONY: all
all: helm-all

ROOT_DIR := ../..
DOCKER_NAME := m4d-kafka
DOCKER_TAGNAME := 0.1.0
HELM_VALUES := \
	--values values.yaml.sample \
	--set image=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/mover:${DOCKER_TAGNAME}

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/helm.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk

helm-crd-hook: crd-stream
.PHONY: crd-stream
crd-stream:
	$(ABSTOOLBIN)/kubectl apply -f $(ROOT_DIR)/manager/config/crd/bases/motion.m4d.ibm.com_streamtransfers.yaml
