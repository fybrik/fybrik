#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

FYBRIK_NAMESPACE=fybrik-system

kubectl create ns ${FYBRIK_NAMESPACE}
# Apply certificates for tls authentication
kubectl -n ${FYBRIK_NAMESPACE} apply -f ca-certificate.yaml
kubectl -n ${FYBRIK_NAMESPACE} apply -f katalog-connector-certificates.yaml
kubectl -n ${FYBRIK_NAMESPACE} apply -f opa-server-certificates.yaml
kubectl -n ${FYBRIK_NAMESPACE} apply -f opa-connector-certificates.yaml
kubectl -n ${FYBRIK_NAMESPACE} apply -f manager-certificates.yaml
kubectl -n ${FYBRIK_NAMESPACE} apply -f vault-certificates.yaml


