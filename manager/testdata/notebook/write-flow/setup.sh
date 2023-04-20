#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

kubectl delete namespace fybrik-notebook-sample || true
kubectl create namespace fybrik-notebook-sample
kubectl config set-context --current --namespace=fybrik-notebook-sample

FYBRIK_NAMESPACE=fybrik-system
if [[ -z "${ADMIN_CRS_NAMESPACE}" ]]; then
  ADMIN_CRS_NAMESPACE=fybrik-system
fi

# Create the storage-accounts
kubectl apply -f bucket-creds.yaml -n ${ADMIN_CRS_NAMESPACE}
kubectl apply -f theshire-storage-account.yaml -n ${ADMIN_CRS_NAMESPACE}
kubectl apply -f neverland-storage-account.yaml -n ${ADMIN_CRS_NAMESPACE}

# Avoid using webhooks in tests
kubectl delete validatingwebhookconfiguration fybrik-system-validating-webhook

# Apply policies
kubectl -n ${FYBRIK_NAMESPACE} apply -f policy-cm.yaml
kubectl -n ${FYBRIK_NAMESPACE} label configmap policy openpolicyagent.org/policy=rego
while [[ $(kubectl get cm policy -n ${FYBRIK_NAMESPACE} -o 'jsonpath={.metadata.annotations.openpolicyagent\.org/policy-status}') != '{"status":"ok"}' ]]; do echo "waiting for policy to be applied" && sleep 5; done

if [[ -z "${LATEST_BACKWARD_SUPPORTED_AFM_VERSION}" ]]; then
  # Use master version of arrow-flight-module according to https://github.com/fybrik/arrow-flight-module#version-compatbility-matrix
  kubectl apply -f https://raw.githubusercontent.com/fybrik/arrow-flight-module/master/module.yaml -n ${ADMIN_CRS_NAMESPACE}
else
  kubectl apply -f https://github.com/fybrik/arrow-flight-module/releases/download/${LATEST_BACKWARD_SUPPORTED_AFM_VERSION}/module.yaml -n ${ADMIN_CRS_NAMESPACE}
fi

# Forward port of test S3 instance
kubectl port-forward -n ${FYBRIK_NAMESPACE} svc/s3 9090:9090 &
