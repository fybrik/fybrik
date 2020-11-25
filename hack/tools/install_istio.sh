#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


source ./common.sh

export ISTIO_VERSION=1.6.2

header_text "Checking for bin/istioctl"
[[ -f bin/istioctl ]] && exit 0

header_text "Installing bin/istioctl"
mkdir -p ./bin

curl -sL https://istio.io/downloadIstioctl | sh -
install ${HOME}/.istioctl/bin/istioctl bin/.
