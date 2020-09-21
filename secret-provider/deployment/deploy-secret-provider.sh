#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


: ${KUBE_NAMESPACE:=m4d-system}

# DOCKER_PASSWORD is from the CI or from env variable
deploy() {
    kubectl apply -n $KUBE_NAMESPACE -f secret-provider/secret-provider-rbac.yaml
    if [[ -z $(kubectl get secrets cloud-registry  -n $KUBE_NAMESPACE --ignore-not-found) ]]; then
        kubectl create secret docker-registry cloud-registry \
                --docker-server="$DOCKER_HOSTNAME" \
                --docker-username="$DOCKER_USERNAME" \
                --docker-password="$DOCKER_PASSWORD" \
                -n $KUBE_NAMESPACE
    fi

    kubectl patch serviceaccount secret-provider -p \
    '{"imagePullSecrets": [{"name": "cloud-registry"}]}' \
    -n $KUBE_NAMESPACE

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
