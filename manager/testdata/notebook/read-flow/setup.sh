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
# Use master version of arrow-flight-module according to https://github.com/fybrik/arrow-flight-module#version-compatbility-matrix
kubectl apply -f https://raw.githubusercontent.com/fybrik/arrow-flight-module/master/module.yaml -n fybrik-system
# Forward port of test S3 instance
kubectl port-forward -n fybrik-system svc/s3 9090:9090 &
