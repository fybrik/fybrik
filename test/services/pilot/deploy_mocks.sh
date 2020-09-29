#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e

: ${KUBE_NAMESPACE:=irltest4}
: ${WITHOUT_VAULT=false}
: ${ROOT_DIR=../../../}

POLICY_DIR=$ROOT_DIR/pkg/policy-compiler
source $POLICY_DIR/policy-compiler.env


registry_delete() {
        [ -n "$DOCKER_USERNAME" ] || return
        kubectl delete secret docker-registry cloud-registry || true
}

registry_create() {
        [ -n "$DOCKER_USERNAME" ] || return

        kubectl create secret docker-registry cloud-registry \
            --docker-server="$DOCKER_HOSTNAME" \
            --docker-username="$DOCKER_USERNAME" \
            --docker-password="$DOCKER_PASSWORD" \
            -n $KUBE_NAMESPACE

        kubectl patch serviceaccount default -p \
            '{"imagePullSecrets": [{"name": "cloud-registry"}]}' \
            -n $KUBE_NAMESPACE
}

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

kube_cluster_info() {
        printf "\nSleep ...\n"
        sleep 10

        printf "\nGetting current status of active cluster. Please wait...\n"
        kubectl cluster-info
        kubectl get all

        printf "\n(TIP:) You can use the command \'watch kubectl get all\' to continuously monitor the cluster resources!\n"
        printf "\nThe deployment script has completed successfully!\n"
}

case "$REGISTRY" in
docker)
    echo "Registry docker not supported" || exit 1
    ;;
localhost)
    echo "Registry localhost not supported" || exit 1
    ;;
*)
    REGISTRY=ibmcloud 
    ;;
esac

undeploy() {
        registry_delete
        mocks_delete
        kube_cluster_info
}

deploy() {
        registry_delete
        registry_create
        mocks_delete
        mocks_create
        kube_cluster_info
}

case "$1" in
undeploy)
    undeploy
    ;;
*)
    deploy
    ;;
esac
