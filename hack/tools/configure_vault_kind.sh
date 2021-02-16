#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x

op=$1

source ./common.sh
source ./vault_utils.sh

: ${KUBE_NAMESPACE:=m4d-system}
: ${INGRESS_ADDRESS:=http://localhost:80}
: ${KIND_CLUSTER_KUBE_HOST:=https://kind-control-plane:6443}
: ${CONTROL_CLUSTER_KUBE_HOST:=https://control-control-plane:6443}
: ${MODULE_NAMESPACE:="m4d-blueprints"}
: ${SECRET_PATH:=m4d/dataset-creds}
: ${ROLE:=module}


# $1 - cluster name
# $2 - kube host of the cluster
enable_k8s_auth_for_cluster() {
        kubectl config use-context kind-"$1"
        kubectl create ns $KUBE_NAMESPACE || true
        kubectl apply -f vault_auth_sa.yaml -n "$KUBE_NAMESPACE"
        enable_k8s_auth "$1" vault-auth "$KUBE_NAMESPACE" "$2"
}

configure_vault() {
	enable_kv "$SECRET_PATH"
	create_policy "allow-all-$SECRET_PATH" "$SECRET_PATH/*"
}


# $1 - cluster name
add_role() {
	create_role "$ROLE" "allow-all-$SECRET_PATH" "$1" "$MODULE_NAMESPACE"
}

case "$op" in
    cleanup)
        ;;
    multi)
        header_text "Configure Vault on kind multi-cluster"
        kubectl config use-context kind-control --namespace=$KUBE_NAMESPACE
        export VAULT_ADDR=$INGRESS_ADDRESS
        export VAULT_TOKEN=$(kubectl get secrets vault-unseal-keys -n $KUBE_NAMESPACE -o jsonpath={.data.vault-root} | base64 --decode)
        bin/vault login "$VAULT_TOKEN"
        enable_k8s_auth_for_cluster control "$CONTROL_CLUSTER_KUBE_HOST"
        enable_k8s_auth_for_cluster kind "$KIND_CLUSTER_KUBE_HOST"
        configure_vault
        add_role kind
        add_role control
        ;;
    *)
        header_text "Configure Vault on kind cluster"
        kubectl config use-context kind-kind --namespace=$KUBE_NAMESPACE
        export VAULT_TOKEN=$(kubectl get secrets vault-unseal-keys -n $KUBE_NAMESPACE -o jsonpath={.data.vault-root} | base64 --decode)
        port_forward
        export VAULT_ADDR="http://127.0.0.1:8200"
        bin/vault login "$VAULT_TOKEN"
	enable_k8s_auth_for_cluster kind "$KIND_CLUSTER_KUBE_HOST"
	configure_vault
	add_role kind
        # Kill the port-forward if nessecarry
        kill -9 %%
        ;;
esac

