#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

source ./common.sh

header_text "Checking for bin/yq"
[[ -f bin/yq ]] && exit 0

header_text "Installing bin/yq"
mkdir -p ./bin
curl -L https://github.com/mikefarah/yq/releases/download/v4.6.0/yq_${os}_${arch} -o bin/yq
chmod +x bin/yq
