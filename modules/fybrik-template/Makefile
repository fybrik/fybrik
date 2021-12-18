.PHONY: all
all: helm-all

ROOT_DIR := ../..
DOCKER_NAME := fybrik-template
CHART_LOCAL_PATH := ../${DOCKER_NAME}
DOCKER_TAGNAME ?= 0.0.0

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/helm.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk
