#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

cd "${0%/*}"
source ./common.sh


header_text "Checking for bin/kubeval ${KUBEVAL_VERSION}"
[[ -f bin/kubeval && `bin/kubeval version | cut -f2 -d" " | head -n1` == ${KUBEVAL_VERSION} ]] && exit 0

mkdir -p ./bin

curl -sSLo envtest-bins.tar.gz https://github.com/instrumenta/kubeval/releases/download/v${KUBEVAL_VERSION}/kubeval-linux-amd64.tar.gz

tar -zvxf envtest-bins.tar.gz
mv kubeval bin
rm envtest-bins.tar.gz
