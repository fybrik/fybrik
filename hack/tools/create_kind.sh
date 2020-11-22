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
            --network kind \
            -v ${PWD}/../registry:/registry \
            -e REGISTRY_HTTP_ADDR=0.0.0.0:5000 \
            -e REGISTRY_HTTP_TLS_CERTIFICATE=/registry/registry.crt \
            -e REGISTRY_HTTP_TLS_KEY=/registry/registry.key \
            registry:2
        fi
}

certs_create() {
    mkdir ../registry -p || true
    openssl genrsa -out ../registry/ca.key 2048
    openssl req -new -x509 -key ../registry/ca.key -out ../registry/ca.crt -subj '/C=US/ST=NY/O=IBM/CN=ibm' -extensions EXT -config <(printf "[dn]\nCN=ibm\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:ibm\nbasicConstraints=CA:TRUE,pathlen:0")
    openssl genrsa -out ../registry/registry.key 2048
    openssl req -new -key ../registry/registry.key -out ../registry/registry.csr -subj '/C=US/ST=NY/O=IBM/CN=kind-registry' -extensions EXT -config <(printf "[dn]\nCN=kind-registry\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:kind-registry,DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
    openssl x509 -req -in ../registry/registry.csr -CA ../registry/ca.crt -CAkey ../registry/ca.key -CAcreateserial -out ../registry/registry.crt
}

install_certs() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        sudo cp ../registry/ca.crt /usr/local/share/ca-certificates
	      sudo update-ca-certificates
	  elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo OSX will ask to provide your password in order to install the CA certificate to keychain!
        # Installs the CA certificate to the user local keychain
        security add-trusted-cert -r trustRoot -k ~/Library/Keychains/login.keychain-db -e hostnameMismatch ../registry/ca.crt
    else
        echo Please install the certificates in $PWD/../registry/ca.crt !
    fi
}

certs_delete() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
      security remove-trusted-cert  ../registry/ca.crt
    fi
    rm -rf ../registry
}

kind_delete() {
        bin/kind delete cluster --name $1
}

kind_create() {
        bin/kind create cluster --name $1 \
             -v 4 --retain --wait=0s \
             --config ./$2 \
             --image=kindest/node:$K8S_VERSION
        bin/kubectl config use-context kind-$1
        for node in $(kind get nodes --name $1); do
          bin/kubectl annotate node "${node}" "tilt.dev/registry=kind-registry:5000";
          docker cp ../registry/ca.crt "$node":/usr/local/share/ca-certificates
          docker exec "$node" update-ca-certificates
        done
}

case "$op" in
    cleanup)
        header_text "Uninstalling kind cluster"
        registry_delete || true
        certs_delete || true
        kind_delete kind || true
        kind_delete control || true
        ;;
    multi)
        header_text "Installing kind multi-cluster"
        certs_create
        install_certs
        kind_create kind kind-config.yaml
        registry_create
        kind_create control kind-control-config.yaml
        ;;
    *)
        header_text "Installing kind cluster"
        certs_create
        install_certs
        kind_create kind kind-config.yaml
        registry_create
        ;;
esac
