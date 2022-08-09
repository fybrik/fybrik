#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0


source ./common.sh

header_text "Checking for bin/helm ${HELM_VERSION}"
[[ -f bin/helm &&  `bin/helm version --template='{{.Version}}'` == ${HELM_VERSION} ]] && exit 0

header_text "Installing bin/helm ${HELM_VERSION}"

mkdir -p ./bin
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
chmod 700 get_helm.sh
HELM_INSTALL_DIR=bin ./get_helm.sh -v ${HELM_VERSION} --no-sudo
rm -rf get_helm.sh

