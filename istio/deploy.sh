#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e

: ${KUBE_NAMESPACE=m4d-system}
: ${WITHOUT_ISTIO=false}
: ${ROOT_DIR=../}

inject_manager_sidecar() {
       kubectl get deployment m4d-controller-manager -o yaml | istioctl kube-inject -f - | kubectl apply -f -; \
       kubectl wait --for=condition=available -n ${KUBE_NAMESPACE} deployment/m4d-controller-manager --timeout=120s; \
}

uninject_manager_sidecar() {
       kubectl get deployment m4d-controller-manager -o yaml | istioctl x kube-uninject -f - | kubectl apply -f -; \
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
        $WITHOUT_ISTIO || uninject_manager_sidecar
}

deploy() {
        $WITHOUT_ISTIO || inject_manager_sidecar
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

