#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e

: ${KUBE_NAMESPACE:=m4d-system}
: ${WITHOUT_VAULT=true}

source vault-util.sh

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
  deploy)
    deploy
    ;;
  undeploy)
    undeploy
    ;;
  configure_path)
      export AUTH_METHOD=K8S
      configure_path
    ;;
  *)
    echo "usage: %0 [deploy|undeploy|configure_path]"
    exit 1
esac
