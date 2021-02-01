#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x

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
      --network kind \
      -e REGISTRY_HTTP_ADDR=0.0.0.0:5000 \
      registry:2
  fi
}

kind_delete() {
  bin/kind delete cluster --name $1
}

kind_create() {
  bin/kind create cluster --name $1 \
    -v 4 --retain --wait=0s \
    --config ./$2 \
    --image=kindest/node:$K8S_VERSION
  for node in $(kind get nodes --name $1); do
    bin/kubectl annotate node "${node}" "tilt.dev/registry=localhost:5000" --context kind-${1}
    docker cp ../registry/themeshfordata-ca.crt "$node":/usr/local/share/ca-certificates
    docker exec "$node" update-ca-certificates
  done
}

case "$op" in
cleanup)
  header_text "Uninstalling kind cluster"
  registry_delete || true
  kind_delete kind || true
  kind_delete control || true
  ;;
multi)
  header_text "Installing kind multi-cluster"
  kind_create kind kind-config.yaml &
  kind_create control kind-control-config.yaml &
  wait
  registry_create
  ;;
*)
  header_text "Installing kind cluster"
  kind_create kind kind-config.yaml
  registry_create
  ;;
esac
