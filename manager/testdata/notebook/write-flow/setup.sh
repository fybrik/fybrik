#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

kubectl create namespace fybrik-notebook-sample
kubectl config set-context --current --namespace=fybrik-notebook-sample

FYBRIK_NAMESPACE=fybrik-system

# Create asset and secret
kubectl -n ${FYBRIK_NAMESPACE} apply -f bucket-creds.yaml
kubectl -n ${FYBRIK_NAMESPACE} apply -f theshire-storage-account.yaml
kubectl -n ${FYBRIK_NAMESPACE} apply -f neverland-storage-account.yaml

# Avoid using webhooks in tests
kubectl delete validatingwebhookconfiguration fybrik-system-validating-webhook
# Use master version of arrow-flight-module according to https://github.com/fybrik/arrow-flight-module#version-compatbility-matrix
kubectl apply -f https://raw.githubusercontent.com/fybrik/arrow-flight-module/master/module.yaml -n ${FYBRIK_NAMESPACE}
# Forward port of test S3 instance
kubectl port-forward -n ${FYBRIK_NAMESPACE} svc/s3 9090:9090 &
