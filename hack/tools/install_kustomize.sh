#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0


source ./common.sh

version=3.2.0

header_text "Checking for bin/kustomize ${KUSTOMIZE_VERSION}"
[[ -f bin/kustomize && `bin/kustomize version --short` == ${KUSTOMIZE_VERSION} ]] && exit 0

header_text "Installing for bin/kustomize ${KUSTOMIZE_VERSION}"
mkdir -p ./bin
curl -L https://github.com/kubernetes-sigs/kustomize/releases/download/v${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_${os}_${arch} -o ./bin/kustomize
chmod +x ./bin/kustomize
