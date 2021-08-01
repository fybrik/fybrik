#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

# This script configures Vault which is deployed on the control-plane namespace (usually fybrik-system)
# in multi clusters setup which includes two kind clusters.
# It defines Vault role for modules running in fybrik-blueprints namespace to authentication against
# Vault to retrieve dataset credentials.
# To create the two kind clusters the create_kind.sh script should be used as follows:
# ./create_kind.sh multi

set -x

op=$1

source ./common.sh
source ./vault_utils.sh

: ${KUBE_NAMESPACE:=fybrik-system}
: ${INGRESS_ADDRESS:=http://localhost:80}
: ${KIND_CLUSTER_KUBE_HOST:=https://kind-control-plane:6443}
: ${CONTROL_CLUSTER_KUBE_HOST:=https://control-control-plane:6443}
: ${MODULE_NAMESPACE:="fybrik-blueprints"}
: ${ROLE:=module}
# Add policy and role for modules running in fybrik-blueprints namespace to
# use the vault-plugin-secrets-kubernetes-reader plugin enabled in Vault
# path kubernetes-secrets.
: ${PLUGIN_PATH:=kubernetes-secrets}
: ${DATA_PROVIDER_USERNAME:=data_provider}
: ${DATA_PROVIDER_PASSWORD:=password}

# $1 - cluster name
# $2 - kube host of the cluster
enable_k8s_auth_for_cluster() {
        kubectl config use-context kind-"$1"
        kubectl create ns $KUBE_NAMESPACE || true
        kubectl apply -f vault_auth_sa.yaml -n "$KUBE_NAMESPACE"
        enable_k8s_auth "$1" vault-auth "$KUBE_NAMESPACE" "$2"
}

configure_vault() {
        create_policy_with_plugin_path "allow-all-dataset-creds" "$PLUGIN_PATH/*"
}


# $1 - cluster name
add_role() {
	create_role "$ROLE" "allow-all-dataset-creds" "$1" "$MODULE_NAMESPACE"
}

add_userpass_auth_method() {
	kubectl config use-context kind-control
	enable_userpass_auth "$DATA_PROVIDER_USERNAME" "$DATA_PROVIDER_PASSWORD" "allow-all-dataset-creds"
}

case "$op" in
    *)
        header_text "Configure Vault on kind multi-cluster"
        kubectl config use-context kind-control --namespace=$KUBE_NAMESPACE
        export VAULT_ADDR=$INGRESS_ADDRESS
        export VAULT_TOKEN=$(kubectl get secrets vault-credentials -n $KUBE_NAMESPACE -o jsonpath={.data.VAULT_TOKEN} | base64 --decode)
        bin/vault login "$VAULT_TOKEN"
        enable_k8s_auth_for_cluster control "$CONTROL_CLUSTER_KUBE_HOST"
        enable_k8s_auth_for_cluster kind "$KIND_CLUSTER_KUBE_HOST"
        configure_vault
        add_role kind
        add_role control
	add_userpass_auth_method
        # Switch to control cluster after configuration
        kubectl config use-context kind-control

        ;;
esac
