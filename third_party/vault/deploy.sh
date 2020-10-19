#!/usr/bin/env bash

: ${KUBE_NAMESPACE:=m4d-system}
: ${WITHOUT_OPENSHIFT:=true}

REPO=banzaicloud-stable
URL=https://kubernetes-charts.banzaicloud.com 
RELEASE=vault-operator
CHART=banzaicloud-stable/vault-operator 

source vault-util.sh

deploy() {
    kubectl create namespace $KUBE_NAMESPACE 2>/dev/null || true

    $WITHOUT_OPENSHIFT || kubectl apply -f vault-openshift.yaml -n $KUBE_NAMESPACE

    helm repo add $REPO $URL
    helm upgrade --install $RELEASE $CHART -n $KUBE_NAMESPACE

    kubectl apply -f vault-rbac.yaml -n $KUBE_NAMESPACE
    kubectl apply -f vault.yaml -n $KUBE_NAMESPACE
}

undeploy() {
    kubectl delete -f vault.yaml -n $KUBE_NAMESPACE
    kubectl delete -f vault-rbac.yaml -n $KUBE_NAMESPACE

    helm uninstall $RELEASE -n $KUBE_NAMESPACE

    $WITHOUT_OPENSHIFT || kubectl delete -f vault-openshift.yaml -n $KUBE_NAMESPACE
}
case "$1" in
    deploy)
        deploy
        ;;
    undeploy)
        undeploy
        ;;
    wait_for_vault)
      wait_for_vault
      ;;
    *)
        echo "usage: $0 [deploy|undeploy|wait_for_vault]"
        exit 1
        ;;
esac
