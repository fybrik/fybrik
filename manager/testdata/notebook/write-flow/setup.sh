#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

kubectl delete namespace fybrik-notebook-sample || true
kubectl create namespace fybrik-notebook-sample
kubectl config set-context --current --namespace=fybrik-notebook-sample

FYBRIK_NAMESPACE=fybrik-system

# Create the storage-accounts
kubectl -n ${FYBRIK_NAMESPACE} apply -f bucket-creds.yaml
kubectl -n ${FYBRIK_NAMESPACE} apply -f theshire-storage-account.yaml
kubectl -n ${FYBRIK_NAMESPACE} apply -f neverland-storage-account.yaml

kubectl apply -f https://github.com/fybrik/arrow-flight-module/releases/download/v0.9.0/module.yaml -n ${FYBRIK_NAMESPACE}


# Forward port of test S3 instance
kubectl port-forward -n ${FYBRIK_NAMESPACE} svc/s3 9090:9090 &
