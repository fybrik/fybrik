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

.PHONY: config
config:
	kubectl create namespace m4d-system || true
	kubectl create secret generic docker-credentials \
		--from-literal=hostname=$(DOCKER_HOSTNAME) \
		--from-literal=namespace=$(DOCKER_NAMESPACE) \
		--from-literal=tagname=$(DOCKER_TAGNAME) \
		--from-literal=username=$(DOCKER_USERNAME) \
		--from-literal=password=$(DOCKER_PASSWORD) \
		--namespace m4d-system
	kubectl apply --namespace m4d-system -f ../manager/config/prod/deployment_configmap.yaml
	kubectl create --namespace m4d-system secret generic user-vault-unseal-keys \
		--from-literal=user-vault-root=''
	kubectl create --namespace m4d-system secret generic vault-unseal-keys \
		--from-literal=vault-root=''
	kubectl create secret generic user-vault-unseal-keys \
		--from-literal=user-vault-root=''
	kubectl create secret generic vault-unseal-keys \
		--from-literal=vault-root=''
	
.PHONY: m4d
m4d:
	@echo "installing m4d ..."
	helm install m4d-crd m4d-crd \
		--namespace m4d-system --wait --timeout 120s
	helm install m4d m4d \
		--namespace m4d-system --wait --timeout 120s

.PHONY: verify
verify:
	kubectl get pod --all-namespaces
	kubectl wait --for=condition=ready pod --all-namespaces --all --timeout=120s

.PHONY: run-deploy-tests
run-deploy-tests: export ROOT_DIR := .
run-deploy-tests: kind cert-manager vault config m4d verify
