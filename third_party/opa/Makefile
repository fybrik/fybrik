ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl $(TOOLBIN)/kustomize
	kustomize build . | kubectl apply -f -
	

.PHONY: undeploy
undeploy: $(TOOLBIN)/kubectl $(TOOLBIN)/kustomize
	kustomize build . | kubectl delete -f -
