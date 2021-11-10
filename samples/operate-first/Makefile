ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
export RELEASE_NAME=fybrik
export CRD_RELEASE_NAME=fybrik-crd
export VAULT_RELEASE_NAME=fybrik-vault
export NAMESPACE=fybrik-system
export Fybrik_CHART_PATH=$(ROOT_DIR)/charts/fybrik
export CRD_CHART_PATH=$(ROOT_DIR)/charts/fybrik-crd
export VAULT_CHART_PATH=$(ROOT_DIR)/charts/vault

.PHONY: all
all: clean helm-template generate

.PHONY: all
clean: clean
	rm -r */ || true
	rm *.yaml || true

.PHONY: helm-template
helm-template: $(TOOLBIN)/yq $(TOOLBIN)/helm
	$(TOOLBIN)/helm template $(CRD_RELEASE_NAME) -n $(NAMESPACE) $(CRD_CHART_PATH) > fybrik-crd.yaml
	$(TOOLBIN)/yq -i e '.plugins.vaultPluginSecretsKubernetesReader.clusterScope |= false | .plugins.vaultPluginSecretsKubernetesReader.namespaces |= . + ["fybrik-applications"] | .plugins.vaultPluginSecretsKubernetesReader.namespaces |= . + ["fybrik-system"]' $(VAULT_CHART_PATH)/values.yaml
	$(TOOLBIN)/helm template $(VAULT_RELEASE_NAME) --set "vault.injector.enabled=false" --set "vault.server.dev.enabled=true" --values $(VAULT_CHART_PATH)/env/dev/vault-single-cluster-values.yaml -n $(NAMESPACE) $(VAULT_CHART_PATH) | $(TOOLBIN)/yq e 'select(.kind=="MutatingWebhookConfiguration" or .kind=="ClusterRole" or .kind=="ClusterRoleBinding")' -> fybrik-vault.yaml

.PHONY: generate
generate:
	python -m pip install -r requirements.txt
	python filterClusterScoped.py
