ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

WITH_OPENSHIFT ?= false

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl $(TOOLBIN)/helm
	@echo "Installing cert-manager ..."
	$(TOOLBIN)/kubectl create namespace cert-manager
	$(TOOLBIN)/helm repo add jetstack https://charts.jetstack.io
	$(TOOLBIN)/helm repo update
	$(TOOLBIN)/helm install cert-manager jetstack/cert-manager --version v1.6.2 \
		--set installCRDs=true \
		--namespace cert-manager \
		--wait --timeout 120s

.PHONY: undeploy
undeploy: $(TOOLBIN)/kubectl $(TOOLBIN)/helm
	@echo "Uinstalling cert-manager ..."
	$(TOOLBIN)/helm uninstall cert-manager --namespace cert-manager
	$(TOOLBIN)/kubectl delete namespace cert-manager
