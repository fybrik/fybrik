ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

WITH_OPENSHIFT ?= false
VALUES_FILE ?= "m4d/values.yaml"

.PHONY: kind
kind: export ROOT_DIR := .
kind:
	cd .. && $(MAKE) kind

.PHONY: cert-manager
cert-manager:
	@echo "Installing cert-manager ..."
	kubectl create namespace cert-manager
	helm repo add jetstack https://charts.jetstack.io
	helm repo update
	helm install cert-manager jetstack/cert-manager --version v1.2.0 \
		--set installCRDs=true \
		--namespace cert-manager \
		--wait --timeout 120s

.PHONY: vault
vault:
	@echo "Installing vault ..."
	kubectl create namespace m4d-system || true
	helm repo add hashicorp https://helm.releases.hashicorp.com
	helm install vault hashicorp/vault --version 0.9.1 \
                --set "global.openshift=$(WITH_OPENSHIFT)" \
                --set injector.enabled=false \
                --set "server.dev.enabled=true" \
		--values ../third_party/vault/plugin-secrets-kubernetes-reader/values.yaml \
		--namespace m4d-system \
		--wait --timeout 120s
	# Create a secret to hold Vault Token. This token is used when configuaring Vault.
	# default token in dev mode is "root".
	kubectl create secret generic vault-unseal-keys --from-literal=vault-root=root -n m4d-system || true
	# To enable Vault to read kuberenetes secrets clusterrole and clusterrolebinding should be added.
	kubectl apply -f ../third_party/vault/plugin-secrets-kubernetes-reader/clusterrole.yaml -n m4d-system || true
	kubectl apply -f ../third_party/vault/plugin-secrets-kubernetes-reader/clusterrolebinding.yaml -n m4d-system || true

.PHONY: wait-for-vault
wait-for-vault:
	kubectl wait --for=condition=ready --all pod -n m4d-system --timeout=120s

.PHONY:
vault-setup:
	$(MAKE) configure-vault

.PHONY: m4d
m4d:
	@echo "installing m4d ..."
	kubectl create namespace m4d-system || true
	helm install m4d-crd m4d-crd  \
		--namespace m4d-system --wait --timeout 120s
	helm install m4d m4d --values $(VALUES_FILE) \
		--namespace m4d-system --wait --timeout 120s

.PHONY: run-integration-tests
run-integration-tests: export ROOT_DIR := .
run-integration-tests: kind cert-manager vault vault-setup m4d
