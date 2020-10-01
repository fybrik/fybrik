#!/usr/bin/env bash

ROOT_DIR=../../
#cd "${0%/*}"

set -e

: ${KUBE_NAMESPACE:=m4d-system}

source secret-provider/deploy/vault-util.sh

kubectl create ns $KUBE_NAMESPACE|| true

kubectl config set-context --current --namespace=$KUBE_NAMESPACE

kubectl apply -f manager/config/prod/deployment_configmap.yaml

# pushd ${ROOT_DIR}
make cluster-prepare

# Install third party components
make -C third_party/vault deploy
make -C third_party/egeria deploy
make -C third_party/opa deploy


# Waiting for the vault deployment to become ready
# We're using old-school while b/c we can't waint on object that haven't been created, and we can't know for sure that the statefulset had been created so far
while [[ $(kubectl get -n $KUBE_NAMESPACE pods -l statefulset.kubernetes.io/pod-name=vault-0 -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]];
do
    echo "waiting for vault pod to become ready" 
    sleep 5
done

# Perform a port-forward to communicate with Vault
port_forward

# Install the manager, the connectors and the secret-provider
# This also configures the "external" endpoint for mimicking an external Vault installation
WITHOUT_PORT_FORWARD=true WITHOUT_VAULT=false make deploy

# Configure the secret-end point in vault
export AUTH_METHOD=K8S
WITHOUT_PORT_FORWARD=true make -C secret-provider configure-vault

# Kill the port-forward
kill -9 %%
