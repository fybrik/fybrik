#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -e

: ${KUBE_NAMESPACE:=m4d-system}
: ${ROOT_DIR=../..}

NAME=cloud-registry

registry_delete() {
        local namespace=$1
        kubectl delete secret $NAME \
            -n $namespace || true
}

registry_create() {
        local namespace=$1

        kubectl create namespace $namespace 2>/dev/null || true

        kubectl delete secret $NAME \
            -n $namespace 2>/dev/null || true

        kubectl create secret docker-registry $NAME \
            --docker-server="$DOCKER_HOSTNAME" \
            --docker-username="$DOCKER_USERNAME" \
            --docker-password="$DOCKER_PASSWORD" \
            -n $namespace

        kubectl patch serviceaccount default -p \
            "{\"imagePullSecrets\": [{\"name\": \"$NAME\"}]}" \
            -n $namespace
}

[ -n "$DOCKER_PASSWORD" ] || exit 0

case "$1" in
undeploy)
    registry_delete default
    registry_delete $KUBE_NAMESPACE
    ;;
deploy)
    registry_create default
    registry_create $KUBE_NAMESPACE
    ;;
*)
    echo "usage: $0 [deploy|undeploy]"
    exit 1
esac
