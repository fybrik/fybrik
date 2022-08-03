#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


source ./common.sh

header_text "Checking for bin/istioctl ${ISTIO_VERSION}"
[[ -f bin/istioctl && `bin/istioctl version --remote=false` == ${ISTIO_VERSION} ]] && exit 0

header_text "Installing bin/istioctl ${ISTIO_VERSION}"
mkdir -p ./bin

curl -sL https://istio.io/downloadIstioctl | ISTIO_VERSION=${ISTIO_VERSION} sh -
install ${HOME}/.istioctl/bin/istioctl bin/.
