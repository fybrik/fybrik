ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

KUBE_NAMESPACE ?= fybrik-system
WITH_OPENSHIFT ?= false

VAULT_VALUES_FILE ?= charts/vault/env/dev/plugin-secrets-values.yaml

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl $(TOOLBIN)/helm
	@echo "Installing vault ..."
	$(TOOLBIN)/kubectl create namespace $(KUBE_NAMESPACE) || true
	$(TOOLBIN)/helm repo add hashicorp https://helm.releases.hashicorp.com
	$(TOOLBIN)/helm dependency update $(ROOT_DIR)/charts/vault
	$(TOOLBIN)/helm install vault $(ROOT_DIR)/charts/vault \
		--set "vault.global.openshift=$(WITH_OPENSHIFT)" \
                --set "vault.injector.enabled=false" \
		--values $(ROOT_DIR)/$(VAULT_VALUES_FILE) \
		--namespace $(KUBE_NAMESPACE) \
		--wait --timeout 120s

.PHONY: deploy-wait
deploy-wait: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl wait --for=condition=ready --all pod -n $(KUBE_NAMESPACE) --timeout=120s
