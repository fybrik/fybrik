#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x

op=$1

source ./common.sh
source ./vault_utils.sh

: ${KUBE_NAMESPACE:=m4d-system}
: ${INGRESS_ADDRESS:=localhost:80}
: ${KIND_CLUSTER_KUBE_HOST:=https://kind-control-plane:6443}
: ${CONTROL_CLUSTER_KUBE_HOST:=https://control-control-plane:6443}
: ${MODULE_NAMESPACE:="m4d-blueprints"}
: ${SECRET_PATH:=m4d/dataset-creds}
: ${ROLE:=module}


# $1 - cluster name
# $2 - vault root token
# $3 - vault address
# $4 - kube host of the cluster
enable_k8s_auth_for_cluster() {
        kubectl config use-context kind-"$1"
        kubectl create ns $KUBE_NAMESPACE || true
        kubectl apply -f vault_auth_sa.yaml -n "$KUBE_NAMESPACE"
        enable_k8s_auth "$2" "$1" vault-auth "$KUBE_NAMESPACE" "$3" "$4"
}

# $1 - vault root token
# $2 - vault address
configure_vault() {
	enable_kv "$SECRET_PATH" "$1" "$2"
	create_policy "allow-all-$SECRET_PATH" "$SECRET_PATH/*" "$1" "$2"
}


# $1 - cluster name
# $2 - vault root token
# $3 - vault address
add_role() {
	create_role "$ROLE" "$2" "allow-all-$SECRET_PATH" "$1" "$3" "$MODULE_NAMESPACE"
}

case "$op" in
    cleanup)
        ;;
    multi)
        header_text "Configure Vault on kind multi-cluster"
        kubectl config use-context kind-control --namespace=$KUBE_NAMESPACE
        export VAULT_TOKEN=$(kubectl get secrets vault-unseal-keys -n $KUBE_NAMESPACE -o jsonpath={.data.vault-root} | base64 --decode)
        enable_k8s_auth_for_cluster control "$VAULT_TOKEN" "$INGRESS_ADDRESS" "$CONTROL_CLUSTER_KUBE_HOST"
        enable_k8s_auth_for_cluster kind "$VAULT_TOKEN" "$INGRESS_ADDRESS" "$KIND_CLUSTER_KUBE_HOST"
	configure_vault "$VAULT_TOKEN" "$INGRESS_ADDRESS"
	add_role kind "$VAULT_TOKEN" "$INGRESS_ADDRESS"
        add_role control "$VAULT_TOKEN" "$INGRESS_ADDRESS"
        ;;
    *)
        header_text "Configure Vault on kind cluster"
        kubectl config use-context kind-kind --namespace=$KUBE_NAMESPACE
        export VAULT_TOKEN=$(kubectl get secrets vault-unseal-keys -n $KUBE_NAMESPACE -o jsonpath={.data.vault-root} | base64 --decode)
        port_forward
	enable_k8s_auth_for_cluster kind "$VAULT_TOKEN"  'http://127.0.0.1:8200' "$KIND_CLUSTER_KUBE_HOST"
	configure_vault "$VAULT_TOKEN" 'http://127.0.0.1:8200'
	add_role kind "$VAULT_TOKEN" 'http://127.0.0.1:8200'
        # Kill the port-forward if nessecarry
        kill -9 %%
        ;;
esac

