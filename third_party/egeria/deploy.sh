#!/usr/bin/env bash

set -e
set -x

: ${WITHOUT_OPENSHIFT=false}

CHART=odpi-egeria-lab
RELEASE=lab
NAMESPACE=irltest1
TIMEOUT=8m
VERSION="V2.1"

undeploy() {
        # scc permissions for egeria pods to execute
        # this is done as per guidelines mentioned in https://github.com/odpi/egeria/issues/2790#issuecomment-674879248    
        $WITHOUT_OPENSHIFT || oc delete securitycontextconstraints --ignore-not-found egeria-restricted

        helm delete $RELEASE \
            --namespace=$NAMESPACE \
            2>/dev/null \
            || true
}

deploy() {
        helm repo add bitnami https://charts.bitnami.com/bitnami
        helm repo update

        rm -rf egeria
        git clone \
            --depth=1 \
            --branch=$VERSION \
            --filter=blob:none \
            https://github.com/odpi/egeria.git

        # create the ns if it doesn't exist
        kubectl create ns $NAMESPACE || true

        # scc permissions for egeria pods to execute
        # this is done as per guidelines mentioned in https://github.com/odpi/egeria/issues/2790#issuecomment-674879248 
        $WITHOUT_OPENSHIFT || oc create -f egeria-scc.yaml
        $WITHOUT_OPENSHIFT || oc adm policy add-scc-to-user egeria-restricted -z default -n $NAMESPACE

        local chartpath=$(find . -name $CHART | grep charts)
        helm dep update $chartpath
        helm install $RELEASE $chartpath \
            -f lab.yaml \
            --namespace=$NAMESPACE \
            --wait --timeout=${TIMEOUT}

        rm -rf egeria
}

config() {
        echo "Executing Egaria Lab configuration"
        local pod=$(kubectl -n $NAMESPACE get pods \
            -l app.kubernetes.io/component=jupyter,app.kubernetes.io/name=$CHART,app.kubernetes.io/instance=$RELEASE \
            -o=jsonpath='{.items[0].metadata.name}')

        kubectl exec -n $NAMESPACE -it $pod -- \
            jupyter nbconvert --execute egeria-server-config.ipynb

        kubectl exec -n $NAMESPACE -it $pod -- \
            jupyter nbconvert --execute egeria-server-start.ipynb \
            --ExecutePreprocessor.timeout=-1
    }

case "$1" in
    deploy)
        undeploy
        deploy
        config
        ;;
    undeploy)
        undeploy
        ;;
    *)
        echo "usage: $0 [deploy|undeploy]"
        exit 1
        ;;
esac
