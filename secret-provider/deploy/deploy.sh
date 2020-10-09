#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e

: ${KUBE_NAMESPACE:=m4d-system}
: ${WITHOUT_VAULT=true}
: ${ROOT_DIR=../..}

kustomize_build() {
        local operation=$1
        local pass=$2
        local TEMP=$(mktemp -d)
        cp -r base/* $TEMP
        cd $TEMP

        local image=secret-provider
        kustomize edit set image ${image}=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/${image}:${DOCKER_TAGNAME}
        kustomize edit set namespace $KUBE_NAMESPACE
        kustomize build . | kubectl $operation -f - || $pass

        cd -
}

undeploy() {
        kustomize_build delete true
}

deploy() {
        kustomize_build apply false
}

case "$1" in
undeploy)
    undeploy
    ;;
*)
    deploy
    ;;
esac
