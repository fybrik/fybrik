#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


: ${KUBE_NAMESPACE:=m4d-system}

deploy() {
    kubectl apply -n $KUBE_NAMESPACE -f secret-provider/secret-provider-rbac.yaml
    kubectl apply -n $KUBE_NAMESPACE -f secret-provider/secret-provider.yaml
}

undeploy() {
    kubectl delete -n $KUBE_NAMESPACE -f secret-provider
}

case "$1" in
    deploy)
      deploy
    ;;
    undeploy)
      undeploy
    ;;
    *)
      echo "must be one of [deploy, undeploy]"
      exit 1
    ;;
esac
