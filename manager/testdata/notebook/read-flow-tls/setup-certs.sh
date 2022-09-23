#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

kubectl create ns fybrik-system
# Apply certificates for tls authentication
kubectl -n fybrik-system apply -f ca-certificate.yaml
kubectl -n fybrik-system apply -f katalog-connector-certificates.yaml
kubectl -n fybrik-system apply -f opa-server-certificates.yaml
kubectl -n fybrik-system apply -f opa-connector-certificates.yaml
kubectl -n fybrik-system apply -f manager-certificates.yaml
kubectl -n fybrik-system apply -f vault-certificates.yaml


