#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

FYBRIK_NAMESPACE=fybrik-system

# Delete certificates for tls authentication
kubectl -n ${FYBRIK_NAMESPACE} delete -f ca-certificate.yaml
kubectl -n ${FYBRIK_NAMESPACE} delete -f katalog-connector-certificates.yaml
kubectl -n ${FYBRIK_NAMESPACE} delete -f opa-server-certificates.yaml
kubectl -n ${FYBRIK_NAMESPACE} delete -f opa-connector-certificates.yaml
kubectl -n ${FYBRIK_NAMESPACE} delete -f manager-certificates.yaml
kubectl -n ${FYBRIK_NAMESPACE} delete -f vault-certificates.yaml
kubectl -n ${FYBRIK_NAMESPACE} delete -f arrow-flight-module-certificates.yaml
kubectl get secret -n ${FYBRIK_NAMESPACE} --no-headers=true | awk '/test-tls/{print $1}'| xargs  kubectl delete -n ${FYBRIK_NAMESPACE} secret
