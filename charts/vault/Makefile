SHELL := /bin/bash # Use bash syntax
ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk

KUBE_NAMESPACE ?= m4d-system
WITHOUT_OPENSHIFT ?= true

.PHONY: install
install:
	kubectl create namespace $(KUBE_NAMESPACE) 2>/dev/null || true
	helm repo add hashicorp https://helm.releases.hashicorp.com
        ifeq ($(WITHOUT_OPENSHIFT), true)
	$(info installing Vault without openshift)
	helm install vault hashicorp/vault --values values/plugin-secrets-kubernetes-reader.yaml
        else
	helm install vault hashicorp/vault --values values/plugin-secrets-kubernetes-reader.yaml --values values/openshift.yaml
        endif
	kubectl create secret generic vault-unseal-keys --from-literal=vault-root=root -n $(KUBE_NAMESPACE) || true
	kubectl apply -f clusterrole.yaml -n $(KUBE_NAMESPACE) || true
	kubectl apply -f clusterrolebinding.yaml -n $(KUBE_NAMESPACE) || true

.PHONY: uninstall
uninstall:
	kubectl delete secret vault-unseal-keys -n $(KUBE_NAMESPACE)
	kubectl delete -f clusterrole.yaml -n $(KUBE_NAMESPACE)
	kubectl delete -f clusterrolebinding.yaml -n $(KUBE_NAMESPACE) || true
	helm uninstall vault -n $(KUBE_NAMESPACE)

.PHONY: wait-for-vault
wait-for-vault:
	@while [[ -z `kubectl get pods vault-0 -o json | jq 'select(.status.phase=="Running") | true'` ]]; do echo "Waiting for vault pod" ; sleep 2; done
