ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk

.PHONY: deploy
deploy:
	ROOT_DIR=$(ROOT_DIR) ./deploy.sh

.PHONY: undeploy
undeploy:
	ROOT_DIR=$(ROOT_DIR) ./deploy.sh undeploy
