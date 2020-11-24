#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e

: ${KUBE_NAMESPACE=m4d-system}
: ${WITHOUT_ISTIO=false}
: ${ROOT_DIR=../}

enable_sidecar_injection() {
       kubectl label namespace ${KUBE_NAMESPACE} istio-injection=enabled
}

disable_sidecar_injection() {
       kubectl label namespace ${KUBE_NAMESPACE} istio-injection-
}

deploy_policy() {
        cd policies && kustomize edit set namespace ${KUBE_NAMESPACE}
        kustomize build --load_restrictor none . | kubectl apply -f -
        cd -
}

undeploy_policy() {
       kustomize build --load_restrictor none policies | kubectl delete -f -
}

undeploy() {
        $WITHOUT_ISTIO || undeploy_policy
        $WITHOUT_ISTIO || disable_sidecar_injection
}

deploy() {
        $WITHOUT_ISTIO || enable_sidecar_injection
        $WITHOUT_ISTIO || deploy_policy
}


case "$1" in
undeploy)
    undeploy
    ;;
*)
    deploy
    ;;
esac

