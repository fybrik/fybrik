ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/.mk/tools.mk

YAML:=https://github.com/jetstack/cert-manager/releases/download/v0.13.1/cert-manager.yaml

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl apply --validate=false -f $(YAML)
	$(TOOLBIN)/kubectl wait --for=condition=available -n cert-manager deployment/cert-manager-webhook --timeout=180s

.PHONY: undeploy
undeploy: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl delete -f $(YAML)
