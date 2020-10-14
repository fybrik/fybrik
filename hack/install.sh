#!/usr/bin/env bash

set -e

: ${KUBE_NAMESPACE:=m4d-system}
: ${PORT_TO_FORWARD:=8200}

source third_party/vault/vault-util.sh

kubectl create ns $KUBE_NAMESPACE || true

kubectl apply -f manager/config/prod/deployment_configmap.yaml -n $KUBE_NAMESPACE

make cluster-prepare

# Install third party components
make -C third_party/vault deploy
make -C third_party/egeria deploy
make -C third_party/opa deploy

# Waiting for the vault deployment to become ready
make -C third_party/vault wait_for_vault

# Perform a port-forward to communicate with Vault
port_forward

# Install the manager, the connectors and the secret-provider
# This also configures the "external" endpoint for mimicking an external Vault installation
WITHOUT_PORT_FORWARD=true WITHOUT_VAULT=false make deploy

# Configure the internal m4d endpoint in vault (the "secret" endpoint)
WITHOUT_PORT_FORWARD=true make -C secret-provider configure-vault

# Kill the port-forward
kill -9 %%
