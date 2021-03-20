#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x

op=$1

source ./common.sh
source ./vault_utils.sh

# This script configures Vault server deployed in the control-plane in a single cluster setup.
# It defines Vault role for modules running in m4d-blueprints namespace to authentication against
# Vault to retrieve dataset credentials.
# To do that The following is done:
# - enable kubernetes auth path in path called kubernetes.
# - enable Vault kv secret engine to hold the dataset credentials
# - creation of Vault policy to allow to access dataset credentials in the Vault kv secret engine enabled above
# - creation of Vault role for kubernetes auth path to bind policy to identity. The identity is the service accounts
#  in m4d-blueprints namespace and the policy is the policy to allow them to access the
#  path to the dataset credentials.

: ${KUBE_NAMESPACE:=m4d-system}
: ${MODULE_NAMESPACE:="m4d-blueprints"}
: ${SECRET_PATH:=m4d/dataset-creds}
: ${ROLE:=module}
: ${WITHOUT_PORT_FORWARD:=false}

# enable_k8s_auth_for_cluster enables kubernetes auth path
# $1 - cluster name
# $2 - api server address of the cluster
enable_k8s_auth_for_cluster() {
        kubectl create ns $KUBE_NAMESPACE || true
        kubectl apply -f vault_auth_sa.yaml -n "$KUBE_NAMESPACE"
        enable_k8s_auth "$1" vault-auth "$KUBE_NAMESPACE" "$2"
}

# configure_vault enables kv secret engine to hold the dataset credentails
# and creates policy to access them
configure_vault() {
	enable_kv "$SECRET_PATH"
	create_policy "allow-all-$SECRET_PATH" "$SECRET_PATH/*"
}

# add_role adds a role to bind policy to identity.
# $1 - cluster name
add_role() {
	create_role "$ROLE" "allow-all-$SECRET_PATH" "$1" "$MODULE_NAMESPACE"
}

case "$op" in
    cleanup)
        ;;
    *)
	header_text "Configure Vault on a single cluster"
	kubectl config set-context --current --namespace=$KUBE_NAMESPACE
	export VAULT_TOKEN=$(kubectl get secrets vault-unseal-keys -n $KUBE_NAMESPACE -o jsonpath={.data.vault-root} | base64 --decode)
	$WITHOUT_PORT_FORWARD || port_forward
	export VAULT_ADDR="http://127.0.0.1:8200"
	bin/vault login "$VAULT_TOKEN"
	enable_k8s_auth_for_cluster kubernetes "https://kubernetes.default.svc:443"
	configure_vault
	add_role kubernetes
	# Kill the port-forward if nessecarry
	$WITHOUT_PORT_FORWARD || kill -9 %%
	;;
esac

