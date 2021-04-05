#!/usr/bin/env bash

: ${KUBE_NAMESPACE:=m4d-system}
: ${WITHOUT_OPENSHIFT:=true}

REPO=banzaicloud-stable
URL=https://kubernetes-charts.banzaicloud.com 
RELEASE=vault-operator
CHART=banzaicloud-stable/vault-operator 
VERSION=1.11.2

deploy() {
    kubectl create namespace $KUBE_NAMESPACE 2>/dev/null || true

    $WITHOUT_OPENSHIFT || kubectl apply -f vault-openshift.yaml -n $KUBE_NAMESPACE

    helm repo add $REPO $URL
    helm upgrade --install $RELEASE $CHART --version $VERSION -n $KUBE_NAMESPACE

    kubectl apply -f vault-rbac.yaml -n $KUBE_NAMESPACE
    kubectl apply -f vault.yaml -n $KUBE_NAMESPACE
}

deploy-wait() {
    # We're using old-school while b/c we can't wait on object that haven't been created, and we can't know for sure that the statefulset had been created so far
	# See https://github.com/kubernetes/kubernetes/issues/75227
	while [[ $(kubectl get -n ${KUBE_NAMESPACE} pods -l statefulset.kubernetes.io/pod-name=vault-0 -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do \
	    echo "waiting for vault pod to become ready"; \
	    sleep 5; \
	done
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
    deploy-wait)
        deploy-wait
        ;;
    undeploy)
        undeploy
        ;;
    *)
        echo "usage: $0 [deploy|deploy-wait|undeploy]"
        exit 1
        ;;
esac
