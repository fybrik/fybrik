#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

set -x

kubectl delete namespace fybrik-notebook-sample || true
kubectl create namespace fybrik-notebook-sample
kubectl config set-context --current --namespace=fybrik-notebook-sample

# Create asset and secret
kubectl -n fybrik-notebook-sample apply -f s3credentials.yaml

if [[ "${USE_OPENMETADATA_CATALOG}" -eq 0 ]]; then
  # Deploy katalog asset
  kubectl -n fybrik-notebook-sample apply -f katalog-asset.yaml
else
  port=8080
  local_port=8081
  prefix=http

  if [[ "${DEPLOY_TLS_TEST_CERTS}" -eq 1 ]]; then
    port=8443
    prefix=https
    local_port=8443
    kubectl get secret test-tls-localhost-certs -n fybrik-system -o json | jq -r '.data."tls.key"' | base64 -d > tls.key
    kubectl get secret test-tls-localhost-certs -n fybrik-system -o json | jq -r '.data."tls.crt"' | base64 -d > tls.crt
    certs=" --cert tls.crt --key tls.key "
  fi
  # Deploy openmetadata asset
  kubectl port-forward svc/openmetadata-connector -n fybrik-system $local_port:$port &
  # Wait until curl command succeed
  c=0
  # -k flag is used to skip server verification to avoid errors regarding target host name 'localhost'
  while [[ $(curl $certs -k -X POST $prefix://localhost:$local_port/createAsset -d @om-asset.json) != *'assetID'* ]]
  do
    echo "waiting for curl command to createAsset to succeed"
    ((c++)) && ((c==25)) && break
    sleep 1
  done
fi

# Avoid using webhooks in tests
kubectl delete validatingwebhookconfiguration fybrik-system-validating-webhook

if [[ -z "${LATEST_BACKWARD_SUPPORTED_AFM_VERSION}" ]]; then
  # Use master version of arrow-flight-module according to https://github.com/fybrik/arrow-flight-module#version-compatbility-matrix
  kubectl apply -f https://raw.githubusercontent.com/fybrik/arrow-flight-module/master/module.yaml -n fybrik-system
else
  kubectl apply -f https://github.com/fybrik/arrow-flight-module/releases/download/${LATEST_BACKWARD_SUPPORTED_AFM_VERSION}/module.yaml -n fybrik-system
fi

# When Vault uses mutual TLS the certificates and private key for the arrow-flight-module
# are stored in a secret in fybrik-system namespace and copied to fybrik-blueprint namespace
# using a mechanism for syncing secrets across namespaces.
# ref: https://cert-manager.io/docs/tutorials/syncing-secrets-across-namespaces
if ! [[ -z "$PATCH_FYBRIK_MODULE" ]]; then
  # Patch FybrikModule to use the secret
  kubectl patch fybrikmodules.app.fybrik.io arrow-flight-module -n fybrik-system -p "{\"spec\": {\"chart\":{\"values\":{\"tls.certs.cacertSecretName\":\"test-tls-arrow-flight-certs\"}}}}" --type="merge"
  kubectl patch fybrikmodules.app.fybrik.io arrow-flight-module -n fybrik-system -p "{\"spec\": {\"chart\":{\"values\":{\"tls.certs.certSecretName\":\"test-tls-arrow-flight-certs\"}}}}" --type="merge"
fi

# Forward port of test S3 instance
kubectl port-forward -n fybrik-system svc/s3 9090:9090 &
