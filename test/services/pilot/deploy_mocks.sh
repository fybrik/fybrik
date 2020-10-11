#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e

: ${KUBE_NAMESPACE:=default}
: ${WITHOUT_VAULT=false}
: ${ROOT_DIR=../../../}

POLICY_DIR=$ROOT_DIR/pkg/policy-compiler
source $POLICY_DIR/policy-compiler.env

kustomize_build() {
        local operation=$1
        local pass=$2
        local TEMP=$(mktemp -d)
        cp -aR deploy/base/* $TEMP
        cd $TEMP

        local images="ctlg-cred-mock plcy-mgr-mock"
        for image in ${images}; do \
                kustomize edit set image ${image}=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/${image}:${DOCKER_TAGNAME}; \
        done
        kustomize build . | kubectl $operation -f - || $pass

        cd -
}

mocks_delete() {
        printf "\nRemoving kubectl mock resources on active cluster"
        kubectl delete --ignore-not-found -f $ROOT_DIR/manager/config/prod/deployment_configmap.yaml
        #cd $ROOT_DIR/test/services/pilot/deploy
        #kustomize build patch/$REGISTRY | kubectl delete -f - || true
        kustomize_build delete true
        #cd -
}

mocks_create() {
        kubectl apply -f $ROOT_DIR/manager/config/prod/deployment_configmap.yaml
        pwd
        #cd $ROOT_DIR/test/services/pilot/deploy
        #kustomize build patch/$REGISTRY | kubectl apply -f -
        kustomize_build apply false
}

undeploy() {
        mocks_delete
}

deploy() {
        mocks_delete
        mocks_create
}

case "$1" in
undeploy)
    undeploy
    ;;
*)
    deploy
    ;;
esac
