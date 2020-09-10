#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

source ./common.sh

version=1.12.3

header_text "Checking for bin/minikube"
[[ -f bin/minikube ]] && exit 0

header_text "Installing for bin/minikube"
mkdir -p ./bin
curl -Lo ./bin/minikube https://github.com/kubernetes/minikube/releases/download/v${version}/minikube-${os}-${arch}
chmod +x ./bin/minikube
