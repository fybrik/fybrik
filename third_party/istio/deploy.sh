#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

op=$1

istio_status() {
    kubectl get services -n istio-system
    kubectl get pods -n istio-system
}

istio_install() {
    istioctl install -y -f istiooperator.yaml
}

istio_uninstall() {
    istioctl x uninstall -y --purge
}

istio_wait() {
    kubectl wait --for=condition=available -n istio-system deployment/istiod --timeout=120s
}

case "$op" in
    cleanup)
        echo "Uninstalling istio"
        istio_uninstall
        ;;
    status)
        istio_status
        ;;
    wait)
        istio_wait
	;;
    *)
        echo "Installing istio"
        istioctl version
        istio_install
        ;;
esac
