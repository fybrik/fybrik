#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

FYBRIK_NAMESPACE=fybrik-system

kubectl create ns ${FYBRIK_NAMESPACE}
# Apply certificates for tls authentication
kubectl -n ${FYBRIK_NAMESPACE} apply -f ca-certificate.yaml --wait
kubectl -n ${FYBRIK_NAMESPACE} apply -f openmetadata-connector-certificates.yaml --wait
kubectl -n ${FYBRIK_NAMESPACE} apply -f opa-server-certificates.yaml --wait
kubectl -n ${FYBRIK_NAMESPACE} apply -f opa-connector-certificates.yaml --wait
kubectl -n ${FYBRIK_NAMESPACE} apply -f manager-certificates.yaml --wait
kubectl -n ${FYBRIK_NAMESPACE} apply -f vault-certificates.yaml --wait
kubectl -n ${FYBRIK_NAMESPACE} apply -f arrow-flight-module-certificates.yaml --wait
kubectl -n ${FYBRIK_NAMESPACE} apply -f localhost-certificates.yaml --wait

