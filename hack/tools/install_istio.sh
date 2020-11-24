#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


source ./common.sh

export ISTIO_VERSION=1.8.0

header_text "Checking for bin/istioctl"
[[ -f bin/istioctl ]] && exit 0

header_text "Installing bin/istioctl"
mkdir -p ./bin
mkdir -p ./istio
curl -L https://istio.io/downloadIstio | sh -
install ./istio-$ISTIO_VERSION/bin/istioctl bin/.
cp -rf ./istio-$ISTIO_VERSION/samples istio
rm -rf ./istio-$ISTIO_VERSION
