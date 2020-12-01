#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e

: ${KUBE_NAMESPACE=m4d-system}
: ${WITHOUT_ISTIO=false}
: ${ROOT_DIR=../../}

enable_sidecar_injection() {
       kubectl label namespace ${KUBE_NAMESPACE} istio-injection=enabled 
}

disable_sidecar_injection() {
       kubectl label namespace ${KUBE_NAMESPACE} istio-injection-
}

kustomize_build() {
        local operation=$1
        local pass=$2
        local TEMP=$(mktemp -d)
        cp -r istio/policies/* $TEMP
        cd $TEMP

        kustomize edit set namespace ${KUBE_NAMESPACE}
        kustomize build . | kubectl $operation -f - || $pass

        cd -
}

undeploy() {
        $WITHOUT_ISTIO || undeploy_policy
        $WITHOUT_ISTIO || kustomize_build delete true
}

deploy() {
        $WITHOUT_ISTIO || enable_sidecar_injection
        $WITHOUT_ISTIO || kustomize_build apply false
}


case "$1" in
undeploy)
    undeploy
    ;;
*)
    deploy
    ;;
esac

