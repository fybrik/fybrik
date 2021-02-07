#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

: ${PORT_TO_FORWARD:=8200}

# $1 role name
# $2 Vault token
# $3 policy name
# $4 path to auth method
# $5 vault address
# $6 bound_service_account_namespaces
create_role() {
        # Configure a role for the secret-provider
        echo "creating role $1 in k8s auth"
        curl \
        --header "X-Vault-Token: $2" \
        --request POST \
        --data '{"bound_service_account_names": "*", "bound_service_account_namespaces": "'$6'", "policies": "'$1'", "ttl": "24h"}' \
        "$5"/v1/auth/$4/role/"$1"
}

# $1 - the path of the secret
# $2 - the secret, as json
# $3 - vault-token
# $4 - vault address
push_secret() {
        curl \
        -H "X-Vault-Token: $3" \
        -H "Content-Type: application/json" \
        -X POST \
        -d "$2" \
        "$4"/v1/"$1"
}

enable_kv() {
        # Enable kv engine to write secrets to vault
        echo "enabling kv engine for endpoint: $1"

        # Enable the secret endpoint
        curl \
        --header "X-Vault-Token: $2" \
        --request POST \
        --data '{"type": "kv", "options": {"version": "1"} }' \
        "$3"/v1/sys/mounts/"$1"

        # Equivalent using the CLI:
        # vault secrets enable -path=$2 -version=1 kv
}

# $1 - vault-token
# $2 - cluster path in vault
# $3 - vault sa secret name
# $4 - vault sa secret namespace
# $5 - vault address
# $6 - kube host
enable_k8s_auth() {
        # Enable k8s service account token auth
        echo "enabling k8s auth $1 $2 $3 $4"
        curl \
        --header "X-Vault-Token: $1" \
        --request POST \
        --data '{"type": "kubernetes"}' \
        "$5"/v1/sys/auth/"$2"

        # Equivalent using the CLI:
        # vault auth enable <auth path>

        TOKEN_REVIEW_JWT=$(kubectl get secret $3 -n $4 -o jsonpath="{.data.token}" | base64 --decode)
        KUBE_CA_CERT=$(jq -n --arg cert "$(kubectl get secret $3 -n $4 -o jsonpath="{.data['ca\.crt']}" | base64 --decode)" '$cert')

        # Configure the k8s sa auth
        echo "configuring k8s auth"
        curl \
        --header "X-Vault-Token: $1" \
        --request POST \
        --data '{"token_reviewer_jwt":"'"$TOKEN_REVIEW_JWT"'", "kubernetes_ca_cert":'"$KUBE_CA_CERT"', "kubernetes_host":"'"$6"'"}' \
        "$5"/v1/auth/"$2"/config
}


# $1 - policy name
# $2 - path
# $3 - vault-token
# $4 - vault address
create_policy() {
        echo "creating policy $1, to access the secrets in: $2 $3 "
        curl \
        --header "X-Vault-Token: $3" \
        --request PUT \
        --data '{"policy": "path \"'$2'\" { capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"] }"}' \
        "$4"/v1/sys/policy/"$1"
        # Equivalent using the CLI:
        # vault policy write "$1" - <<EOF
}

# Do port-forwarding
port_forward() {
        # Port forward, so we could access vault
        echo "Creating a port-forward from $PORT_TO_FORWARD to 8200 for Vault"
        kubectl port-forward -n $KUBE_NAMESPACE service/vault "$PORT_TO_FORWARD":8200 &
        while ! nc -z localhost "$PORT_TO_FORWARD"; do echo "Waiting for the port-forward from $PORT_TO_FORWARD to 8200 to take effect"; sleep 1; done
}
