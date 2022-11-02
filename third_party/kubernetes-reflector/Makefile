ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

.PHONY: deploy
deploy: $(TOOLBIN)/helm
	@echo "Installing vault ..."
	$(TOOLBIN)/helm repo add emberstack https://emberstack.github.io/helm-charts
	$(TOOLBIN)/helm repo update
	$(TOOLBIN)/helm upgrade --install reflector emberstack/reflector -n default

.PHONY: deploy-wait
deploy-wait: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl wait --for=condition=ready --all pod -n default --timeout=120s
