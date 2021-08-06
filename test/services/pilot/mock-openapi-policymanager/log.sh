#!/usr/bin/env bash

set -e
set -x

POD=$(kubectl get pod -l app=manager-client -n fybrik-system -o jsonpath="{.items[0].metadata.name}")
kubectl logs -n fybrik-system $POD >out-managerclient.log


POD=$(kubectl get pod -l app.kubernetes.io/name=wkc-connector -n fybrik-system -o jsonpath="{.items[0].metadata.name}")
kubectl logs -n fybrik-system $POD >out-wkc-connector.log


POD=$(kubectl get pod -l app.kubernetes.io/component=opa-connector -n fybrik-system -o jsonpath="{.items[0].metadata.name}")
kubectl logs -n fybrik-system $POD >out-opa-connector.log