#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


op=$1

source ./common.sh

K8S_VERSION=${K8S_VERSION:-v1.16.9}

registry_delete() {
        docker network disconnect kind kind-registry
	docker kill kind-registry
        docker rm -f kind-registry
}

registry_create() {
        running="$(docker inspect -f '{{.State.Running}}' "kind-registry" 2>/dev/null || true)"
        if [ "${running}" != 'true' ]; then
          docker run \
            -d --restart=always -p "5000:5000" --name "kind-registry" \
            registry:2
        fi
        docker network connect kind kind-registry
}

kind_delete() {
        bin/kind version
        bin/kind delete cluster --name kind
}

kind_create() {
        bin/kind version
        bin/kind create cluster --name kind \
             -v 4 --retain --wait=1m \
             --config ./kind-config.yaml \
             --image=kindest/node:$K8S_VERSION
        bin/kubectl config use-context kind-kind
        for node in $(kind get nodes); do
          bin/kubectl annotate node "${node}" "tilt.dev/registry=localhost:5000";
        done
}

case "$op" in
    cleanup)
        header_text "Uninstalling kind cluster"
        registry_delete || true
        kind_delete || true
        ;;
    *)
        header_text "Installing kind cluster"
        kind_create
        registry_create
        ;;
esac
