ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env

.PHONY: all
all:
	@echo "usage: make [kind|minikube|openshift]"

include $(ROOT_DIR)/hack/make-rules/docker.mk

.PHONY: deploy
deploy:
	ROOT_DIR=$(ROOT_DIR) ./deploy.sh

.PHONY: undeploy
undeploy:
	ROOT_DIR=$(ROOT_DIR) ./deploy.sh undeploy
