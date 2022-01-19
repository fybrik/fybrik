#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


: ${NS_TEST:=test-crd-charts}
: ${PATH_TO_ROOT:=..}

test_crd_charts() {
  errors=0
  "$PATH_TO_ROOT"/"$TOOLBIN"/kubectl create ns "$NS_TEST"
  "$PATH_TO_ROOT"/"$TOOLBIN"/kubectl get ns

  # First install the Asset crd, which is installed by default
  "$PATH_TO_ROOT"/"$TOOLBIN"/helm install fybrik-crd fybrik-crd  --namespace "$NS_TEST" --wait --timeout 120s
  name=$("$PATH_TO_ROOT"/"$TOOLBIN"/kubectl get crds assets.katalog.fybrik.io -o json | jq -r '.metadata.name')

  if [ "$name" != "assets.katalog.fybrik.io" ]; then
    echo "assets.katalog.fybrik.io wasn't installed when it should have!"
    errors=1
  fi

  "$PATH_TO_ROOT"/"$TOOLBIN"/helm uninstall fybrik-crd --namespace "$NS_TEST"

  # Now install the CRDs without the Asset CRD
  "$PATH_TO_ROOT"/"$TOOLBIN"/helm install fybrik-crd fybrik-crd  --namespace "$NS_TEST" --wait --timeout 120s --set asset-crd.enabled=false
  name=$("$PATH_TO_ROOT"/"$TOOLBIN"/kubectl get crds assets.katalog.fybrik.io -o json --ignore-not-found | jq '.metadata.name')
  if [[ -n "$name" ]]; then
      echo "assets.katalog.fybrik.io was installed when it shouldn't have!"
      errors=1
    fi

  "$PATH_TO_ROOT"/"$TOOLBIN"/helm uninstall fybrik-crd --namespace "$NS_TEST"
  "$PATH_TO_ROOT"/"$TOOLBIN"/kubectl delete ns "$NS_TEST" --timeout=120s
  
  exit "$errors"
}

case "$1" in
  test_crd_charts)
    test_crd_charts
    ;;
esac
