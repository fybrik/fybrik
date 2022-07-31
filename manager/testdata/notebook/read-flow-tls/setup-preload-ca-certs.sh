#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

export ROOT_DIR=../../../../

# Apply certificates for tls authentication
kubectl -n fybrik-system apply -f ca-certificate.yaml
kubectl -n fybrik-system apply -f katalog-connector-certificates.yaml
kubectl -n fybrik-system apply -f opa-connector-certificates.yaml
kubectl -n fybrik-system apply -f manager-certificates.yaml

while ! kubectl get secret test-tls-ca-certs -n fybrik-system; do echo "Waiting for my secret. "; sleep 1; done
kubectl get secret test-tls-ca-certs -o json | jq -r '.data."ca.crt"' | base64 -d > ca.crt

cp ca.crt ${ROOT_DIR}/connectors/katalog/
cp ca.crt ${ROOT_DIR}/connectors/opa/
cp ca.crt ${ROOT_DIR}/manager/
cp ${ROOT_DIR}/connectors/katalog/Dockerfile katalog-connector-Dockerfile.orig
cp ${ROOT_DIR}/connectors/opa/Dockerfile opa-connector-Dockerfile.orig
cp ${ROOT_DIR}/manager/Dockerfile manager-Dockerfile.orig
cp katalog-connector-Dockerfile ${ROOT_DIR}/connectors/katalog/Dockerfile
cp opa-connector-Dockerfile ${ROOT_DIR}/connectors/opa/Dockerfile
cp manager-Dockerfile ${ROOT_DIR}/manager/Dockerfile





