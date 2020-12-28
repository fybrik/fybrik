ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

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
	helm install cert-manager jetstack/cert-manager --version v1.1.0 \
		--set installCRDs=true \
		--namespace cert-manager \
		--wait --timeout 120s

.PHONY: vault
vault:
	@echo "Installing vault ..."
	helm repo add hashicorp https://helm.releases.hashicorp.com
	helm install vault hashicorp/vault --version 0.8.0 \
                --set "global.openshift=false" \
                --set "server.dev.enabled=true" \
		--wait --timeout 120s

.PHONY:
vault-setup:
	cd ../secret-provider && $(MAKE) configure-vault

.PHONY: m4d
m4d:
	@echo "installing m4d ..."
	kubectl create namespace m4d-system || true
	helm install m4d-crd m4d-crd \
		--namespace m4d-system --wait --timeout 120s
	helm install m4d m4d \
		--namespace m4d-system --wait --timeout 120s

.PHONY: run-integration-tests
run-integration-tests: export ROOT_DIR := .
run-integration-tests: kind cert-manager vault vault-setup m4d
