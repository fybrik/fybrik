ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl
	./deploy.sh deploy

.PHONY: undeploy
undeploy: $(TOOLBIN)/kubectl
	./deploy.sh undeploy
