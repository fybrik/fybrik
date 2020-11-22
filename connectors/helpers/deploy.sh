#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e

: ${KUBE_NAMESPACE:=m4d-system}
: ${WITHOUT_VAULT=true}
: ${WITHOUT_ISTIO=false}
: ${ROOT_DIR=../..}
: ${PORT_TO_FORWARD=8200}

POLICY_DIR=$ROOT_DIR/pkg/policy-compiler
source $POLICY_DIR/policy-compiler.env
source $ROOT_DIR/third_party/vault/vault-util.sh

vault_delete() {
        kubectl delete secret user-vault-unseal-keys || true
}

vault_create() {
	echo "USER_VAULT_PATH: $USER_VAULT_PATH"
        #export VAULT_ADDR=http://127.0.0.1:$PORT_TO_FORWARD
        export VAULT_TOKEN=$(kubectl get secrets vault-unseal-keys -n $KUBE_NAMESPACE -o jsonpath={.data.vault-root} | base64 --decode)
        
        export SECRET_PATH=$USER_VAULT_PATH
        export AUTH_METHOD=USERPASS
        configure_path
        push_secret $USER_VAULT_PATH/87ffdca3-8b5d-4f77-99f9-0cb1fba1f73f/01c6f0f0-9ffe-4ccc-ac07-409523755e72 '{"credentials":"my_kafka_credentials"}' $VAULT_TOKEN 
        #vault kv put $USER_VAULT_PATH/87ffdca3-8b5d-4f77-99f9-0cb1fba1f73f/01c6f0f0-9ffe-4ccc-ac07-409523755e72  credentials="my_kafka_credentials"

        # create secret generic user-vault-unseal-keys
        echo -n $VAULT_TOKEN > ./token.txt
        kubectl create secret generic user-vault-unseal-keys --from-file=user-vault-root=./token.txt || true
        rm ./token.txt
}

kustomize_build() {
        local operation=$1
        local pass=$2
        local TEMP=$(mktemp -d)
        cp -r base/* $TEMP
        cd $TEMP

        local images="egr-connector opa-connector vault-connector"
        for image in ${images}; do \
                kustomize edit set image ${image}=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/${image}:${DOCKER_TAGNAME}; \
        done
        kustomize build . | kubectl $operation -f - || $pass

        cd -
}

connectors_delete() {
        kustomize_build delete true
}

connectors_create() {
        kustomize_build apply false
}

kube_cluster_info() {
        printf "\nSleep ...\n"
        sleep 10

        printf "\nGetting current status of active cluster. Please wait...\n"
        kubectl cluster-info
        kubectl get all

        printf "\n(TIP:) You can use the command \'watch kubectl get all\' to continuously monitor the cluster resources!\n"
        printf "\nThe deployment script has completed successfully!\n"
}

istio_sidecar_injection(){
      local svc="egr-connector opa-connector vault-connector"
      for svc in ${svc}; do \
                kubectl get deployment ${svc} -o yaml | istioctl kube-inject -f - | kubectl apply -f -; \
                kubectl wait --for=condition=available -n ${KUBE_NAMESPACE} deployment/${svc} --timeout=120s; \
      done
}

istio_sidecar_uninjection() {
        local svc="egr-connector opa-connector vault-connector"
        for svc in ${svc}; do \
                kubectl get deployment ${svc} -o yaml | istioctl x kube-uninject -f - | kubectl apply -f -; \
        done
}

istio_deploy_policies(){
        local svc="egr-connector opa-connector vault-connector"
        for svc in ${svc}; do \
                kubectl apply -f base/istio/${svc}-authorization.yaml; \
        done
}

istio_undeploy_policies(){
        local svc="egr-connector opa-connector vault-connector"
        for svc in ${svc}; do \
                kubectl delete -f base/istio/${svc}-authorization.yaml; \
        done
}

wait_for_istio() {
        printf "\nWaiting for istio sidecars to be ready. Please wait...\n"
        local svc="egr-connector opa-connector vault-connector"
        for svc in ${svc}; do \
                kubectl wait --for=condition=available -n ${KUBE_NAMESPACE} deployment/${svc} --timeout=120s; \
        done
}

undeploy() {
        $WITHOUT_VAULT || vault_delete
        $WITHOUT_VAULT || istio_sidecar_uninjection
        connectors_delete
        $WITHOUT_ISTIO || istio_undeploy_policies
        kube_cluster_info
}

deploy() {
        $WITHOUT_VAULT || vault_delete
        $WITHOUT_VAULT || vault_create
        connectors_delete
        connectors_create
        $WITHOUT_ISTIO || istio_sidecar_injection
        $WITHOUT_ISTIO || wait_for_istio
        $WITHOUT_ISTIO || istio_deploy_policies
        kube_cluster_info
}

case "$1" in
undeploy)
    undeploy
    ;;
*)
    deploy
    ;;
esac
