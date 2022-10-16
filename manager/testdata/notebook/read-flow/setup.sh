#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


kubectl create namespace fybrik-notebook-sample
kubectl config set-context --current --namespace=fybrik-notebook-sample

# Create asset and secret
kubectl -n fybrik-notebook-sample apply -f asset.yaml
kubectl -n fybrik-notebook-sample apply -f s3credentials.yaml

# Avoid using webhooks in tests
kubectl delete validatingwebhookconfiguration fybrik-system-validating-webhook

if [[ -z "${LATEST_BACKWARD_SUPPORTED_AFM_VERSION}" ]]; then
  # Use master version of arrow-flight-module according to https://github.com/fybrik/arrow-flight-module#version-compatbility-matrix
  kubectl apply -f https://raw.githubusercontent.com/fybrik/arrow-flight-module/master/module.yaml -n fybrik-system
else
  kubectl apply -f https://github.com/fybrik/arrow-flight-module/releases/download/${LATEST_BACKWARD_SUPPORTED_AFM_VERSION}/module.yaml -n fybrik-system
fi

# If vault server uses tls then copy its CA certificate to
# a secret in fybrik-blueprints namespace and configure the arrow flight module
# to use it when getting connecting to Vault.
if ! [[ -z "$VAULT_CA_CERT_SECRET" ]]; then
  # Create a secret in fybrik-blueprints namespace with vault server CA certificate.
  CACERT=$(kubectl get secret "$VAULT_CA_CERT_SECRET" -n fybrik-system -o jsonpath="{.data.ca\.crt}" | base64 -d)
  kubectl create secret generic ca-cert-secret --from-literal=ca.crt="$CACERT"  -n fybrik-blueprints
  # Patch FybrikModule to use the secret
  kubectl patch fybrikmodules.app.fybrik.io arrow-flight-module -n fybrik-system -p "{\"spec\": {\"chart\":{\"values\":{\"tls.certs.cacertSecretName\":\"ca-cert-secret\"}}}}" --type="merge"
fi

# Forward port of test S3 instance
kubectl port-forward -n fybrik-system svc/s3 9090:9090 &
