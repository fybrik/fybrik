ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/.mk/docker.mk

.PHONY: deploy
deploy:
	./deploy.sh deploy

.PHONY: undeploy
undeploy:
	./deploy.sh undeploy
