#!/usr/bin/env bash

set -e

: ${KUBE_NAMESPACE:=m4d-system}
: ${PORT_TO_FORWARD:=8200}
: ${WITHOUT_VAULT:=false}
: ${WITHOUT_EGERIA:=false}
: ${WITHOUT_OPA:=false}

source hack/tools/vault_utils.sh

kubectl create ns $KUBE_NAMESPACE || true
kubectl config set-context --current --namespace=$KUBE_NAMESPACE

kubectl apply -f manager/config/prod/deployment_configmap.yaml
make cluster-prepare
kubectl create secret generic user-vault-unseal-keys --from-literal=vault-root=$(kubectl get secrets vault-unseal-keys -o jsonpath={.data.vault-root} | base64 --decode) || true
# Install third party components
$WITHOUT_EGERIA || make -C third_party/egeria deploy
$WITHOUT_OPA || make -C third_party/opa deploy

# Perform a port-forward to communicate with Vault
port_forward

# Install the manager and the connectors
# This also configures the "external" endpoint for mimicking an external Vault installation
WITHOUT_PORT_FORWARD=true WITHOUT_VAULT=false make deploy

# Configure the internal m4d endpoint in vault (the "secret" endpoint)
WITHOUT_PORT_FORWARD=true make configure-vault

make install

# Kill the port-forward
kill -9 %%
