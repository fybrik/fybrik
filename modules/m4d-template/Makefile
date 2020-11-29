.PHONY: all
all: helm-all

ROOT_DIR := ../..
DOCKER_NAME := m4d-template
DOCKER_TAGNAME := 0.1.0

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/helm.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk
