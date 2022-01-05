#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0


source ./common.sh

DESIRED_VERSION=v3.7.0

header_text "Checking for bin/helm $DESIRED_VERSION"
[[ -f bin/helm &&  `bin/helm version --template='{{.Version}}'` == $DESIRED_VERSION ]] && exit 0

header_text "Installing bin/helm $DESIRED_VERSION"
# Helm is currently fixed to 3.6.3 until the deployment process is migrated to 3.7.0 (https://github.com/helm/helm/releases/tag/v3.7.0)
# Some used experimental OCI features have been removed.
mkdir -p ./bin
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
chmod 700 get_helm.sh
HELM_INSTALL_DIR=bin ./get_helm.sh -v $DESIRED_VERSION --no-sudo
rm -rf get_helm.sh

