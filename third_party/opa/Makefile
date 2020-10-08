ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/.mk/tools.mk

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl $(TOOLBIN)/kustomize
	kustomize build data-and-policies | kubectl apply -f -
	

.PHONY: undeploy
undeploy: $(TOOLBIN)/kubectl $(TOOLBIN)/kustomize
	kustomize build data-and-policies | kubectl delete -f -
