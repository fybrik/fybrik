ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

KUBE_NAMESPACE ?= m4d-system
WITH_OPENSHIFT ?= false

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl $(TOOLBIN)/helm
	@echo "Installing vault ..."
	$(TOOLBIN)/kubectl create namespace $(KUBE_NAMESPACE) || true
	$(TOOLBIN)/helm repo add hashicorp https://helm.releases.hashicorp.com
	$(TOOLBIN)/helm dependency update $(ROOT_DIR)/charts/vault
	$(TOOLBIN)/helm install vault $(ROOT_DIR)/charts/vault \
		--set "vault.global.openshift=$(WITH_OPENSHIFT)" \
                --set "vault.injector.enabled=false" \
                --set "vault.server.dev.enabled=true" \
		--values $(ROOT_DIR)/charts/vault/env/dev/plugin-secrets-kubernetes-reader-values.yaml \
		--namespace $(KUBE_NAMESPACE) \
		--wait --timeout 120s

.PHONY: wait-for-vault
wait-for-vault: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl wait --for=condition=ready --all pod -n $(KUBE_NAMESPACE) --timeout=120s
