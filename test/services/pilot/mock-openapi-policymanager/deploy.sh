#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e

: ${KUBE_NAMESPACE:=fybrik-system}
: ${WITHOUT_VAULT=true}
: ${ROOT_DIR=../../../..}

manager_client_delete() {
        printf "\nRemoving kubectl resources on active cluster"
        $ROOT_DIR/hack/tools/bin/kustomize build . | kubectl delete -f - || true
}

manager_client_create() {
        $ROOT_DIR/hack/tools/bin/kustomize build . | kubectl apply -f -
}

config() {
        kubectl create secret docker-registry cloud-registry  \
                --docker-server="$DOCKER_HOSTNAME" \
                --docker-username="$DOCKER_USERNAME" \
                --docker-password="$DOCKER_PASSWORD" \
                -n $KUBE_NAMESPACE || true
        kubectl patch serviceaccount default -p \
                "{\"imagePullSecrets\": [{\"name\": \"cloud-registry\"}]}" \
                -n $KUBE_NAMESPACE || true
}

undeploy() {
        manager_client_delete
}

deploy() {
        undeploy
	config
        manager_client_create
}

case "$1" in
undeploy)
    undeploy
    ;;
*)
    deploy
    ;;
esac
