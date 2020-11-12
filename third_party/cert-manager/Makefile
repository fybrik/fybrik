ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/.mk/tools.mk

REMOTE:=https://github.com/jetstack/cert-manager/releases/download/v0.13.1/cert-manager.yaml
LOCAL:=$(ABSTOOLBIN)/cert-manager.yaml

$(LOCAL):
	mkdir -p $(shell dirname $(LOCAL))
	wget -x -O $(LOCAL) $(REMOTE)

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl $(LOCAL)
	$(TOOLBIN)/kubectl apply --validate=false -f $(LOCAL)

.PHONY: deploy-wait
deploy-wait: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl wait --for=condition=available -n cert-manager deployment/cert-manager-webhook --timeout=180s

.PHONY: undeploy
undeploy: $(TOOLBIN)/kubectl $(LOCAL)
	$(TOOLBIN)/kubectl delete -f $(LOCAL)
