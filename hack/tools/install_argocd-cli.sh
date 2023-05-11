#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

source ./common.sh

target_os="$os"
if [[ "$target_os" == "darwin" ]]; then
    target_os="mac"
fi

VERSION=${ARGOCD_CLI}
header_text "Checking for bin/argocd-cli ${VERSION}"
[[ -f bin/argocd &&  `bin/argocd version | grep argocd: |awk '{print $2}'` =~ ^v${VERSION} ]] && exit 0

header_text "Installing bin/argocd-cli ${VERSION}"

mkdir -p ./bin

if [[ "$target_os" != "mac" ]]; then
    curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/download/v$VERSION/argocd-linux-amd64
    install -m 555 argocd-linux-amd64 ./bin/argocd
    rm argocd-linux-amd64
else
    curl -sSL -o argocd-darwin-amd64 https://github.com/argoproj/argo-cd/releases/download/v$VERSION/argocd-darwin-amd64
    install -m 555 argocd-darwin-amd64 ./bin/argocd
    rm argocd-darwin-amd64
fi
chmod +x bin/argocd
