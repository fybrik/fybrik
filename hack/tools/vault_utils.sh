#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

: ${PORT_TO_FORWARD:=8200}

# Enable userpass auth method
# $1 user name
# $2 password
# $3 policy
enable_userpass_auth() {
	bin/vault auth enable userpass || true
	bin/vault write auth/userpass/users/"$1" password="$2" policies="$3"
}


# $1 role name
# $2 policy name
# $3 path to auth method
# $4 bound_service_account_namespaces
create_role() {
        echo "creating role $1 in k8s auth"
        bin/vault write auth/"$3"/role/"$1" \
        bound_service_account_names="*" \
        bound_service_account_namespaces="$4" \
        policies=$2 \
        ttl=24h

}

# $1 kv engine path
enable_kv() {
        # Enable kv engine to write secrets to vault
        echo "enabling kv engine for endpoint: $1"
        bin/vault secrets enable -path=$1 -version=1 kv || true
}

# $1 - auth path name in bin/vault
# $2 - bin/vault sa secret name
# $3 - bin/vault sa secret namespace
# $4 - kube host
enable_k8s_auth() {
        # Enable k8s service account token auth
        echo "enabling k8s auth $1 $2 $3 $4"
        bin/vault auth enable -path="$1" kubernetes || true

        TOKEN_REVIEW_JWT=$(kubectl get secret $2 -n $3 -o jsonpath="{.data.token}" | base64 --decode)
        mkdir -p ./tmp
        kubectl get secret $2 -n $3 -o jsonpath="{.data['ca\.crt']}" | base64 --decode > tmp/ca.crt

        # Configure the k8s sa auth
        # TODO: Add issuer for kubernetes versions greater than 1.20
        echo "configuring k8s auth"
        bin/vault write auth/"$1"/config \
        token_reviewer_jwt="$TOKEN_REVIEW_JWT" \
        kubernetes_host="$4" \
        kubernetes_ca_cert=@tmp/ca.crt
        rm -rf ./tmp
}


# $1 - policy name
# $2 - path
create_policy() {
	echo "creating policy $1, to access the secrets in: $2"
        bin/vault policy write "$1" - <<EOF
        path "$2" {
        capabilities = ["create", "read", "update", "delete", "list"]
        }
EOF
}

# Create a policy to allow access to Vault plugin path as well as
# for the path where dataset credentials resides. This is temporary
# until dataset credentials path become obselete.
# $1 - policy name
# $2 - plugin path
create_policy_with_plugin_path() {
        echo "creating policy $1, to access the secrets in: $2"
        bin/vault policy write "$1" - <<EOF
        path "$2" {
        capabilities = ["create", "read", "update", "delete", "list"]
        }
EOF
}


# Do port-forwarding
port_forward() {
        # Port forward, so we could access vault
        echo "Creating a port-forward from $PORT_TO_FORWARD to 8200 for Vault"
        kubectl port-forward -n $KUBE_NAMESPACE service/vault "$PORT_TO_FORWARD":8200 &
        while ! nc -z localhost "$PORT_TO_FORWARD"; do echo "Waiting for the port-forward from $PORT_TO_FORWARD to 8200 to take effect"; sleep 1; done
}
