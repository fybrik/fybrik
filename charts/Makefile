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
	helm dependency update vault
	helm install vault vault \
                --set "vault.global.openshift=$(WITH_OPENSHIFT)" \
                --set "vault.injector.enabled=false" \
                --set "vault.server.dev.enabled=true" \
		--values vault/env/dev/plugin-secrets-kubernetes-reader-values.yaml \
		--namespace m4d-system \
		--wait --timeout 300s

.PHONY: wait-for-vault
wait-for-vault:
	docker ps
	kubectl get pods -o wide -n m4d-system
	kubectl get nodes -o wide
	kubectl describe nodes
	kubectl describe pods -A
	kubectl wait --for=condition=ready --all pod -n m4d-system --timeout=360s	kubectl wait --for=condition=ready --all pod -n m4d-system --timeout=120s

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
