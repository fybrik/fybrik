ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk

KUBE_NAMESPACE ?= m4d-system

.PHONY: deploy
deploy:
	./deploy.sh deploy

.PHONY: undeploy
undeploy:
	./deploy.sh undeploy

.PHONY: deploy-wait
deploy-wait:
	./deploy.sh deploy-wait
