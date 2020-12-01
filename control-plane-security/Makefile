ROOT_DIR:=../
include $(ROOT_DIR)/Makefile.env

.PHONY: deploy
deploy:
	ROOT_DIR=$(ROOT_DIR) ./deploy.sh

.PHONY: undeploy
undeploy:
	ROOT_DIR=$(ROOT_DIR) ./deploy.sh undeploy

