#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

source ./common.sh


header_text "Checking for bin/yq ${YQ_VERSION}"
[[ -f bin/yq && `bin/yq --version | awk '{print $3}'` == ${YQ_VERSION} ]] && exit 0

header_text "Installing bin/yq ${YQ_VERSION}"
mkdir -p ./bin
curl -L https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_${os}_${arch} -o bin/yq
chmod +x bin/yq
