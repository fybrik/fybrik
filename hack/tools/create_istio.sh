#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


op=$1

source ./common.sh

istio_status() {
    kubectl get services -n istio-system
    kubectl get pods -n istio-system
}

manifest_args="
        --set values.global.istioNamespace=istio-system
        --set values.gateways.istio-ingressgateway.enabled=true
        --set values.gateways.istio-egressgateway.enabled=true
        --set values.meshConfig.accessLogFile="/dev/stdout"
        --set values.kiali.enabled=true
        --set values.istiocoredns.enabled=true
"

istio_install() {
    bin/istioctl manifest apply $manifest_args
}

istio_uninstall() {
    bin/istioctl manifest generate $manifest_args | kubectl delete -f -
}

istio_config_gateway() {
    enable=$1
    if [ "$enable" == "true" ]; then
        mode=REGISTER_ONLY
    else
        mode=ALLOW_ANY
    fi
    kubectl get configmap istio -n istio-system -o yaml | \
        sed "s/mode: .*/mode: $mode/g" | \
        kubectl replace -n istio-system -f -
}

istio_config() {
    istio_config_gateway true
    kubectl label namespace default istio-injection=enabled --overwrite
}

istio_kiali() {
    cleartext=admin
    admin=$(echo -n $cleartext | base64)
    cat <<EOF | kubectl apply -n istio-system -f -
apiVersion: v1
kind: Secret
metadata:
  name: kiali
  labels:
    app: kiali
type: Opaque
data:
  username: $admin
  passphrase: $admin
EOF
    echo "Invoke Kiali by running: 'istioctl dashboard kiali'"
    echo "Login to Kiali using $cleartext/$cleartext"
}

case "$op" in
    cleanup)
        header_text "Uninstalling istio"
        istio_uninstall
        ;;
    status)
        istio_status
        ;;
    *)
        header_text "Installing istio"
        bin/istioctl version
        istio_install
        istio_config
        istio_kiali
        ;;

esac
