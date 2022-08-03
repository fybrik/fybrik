#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

source ./common.sh


header_text "Checking for bin/etcd, bin/kube-apiserver and bin/kubectl ${KUBE_VERSION}"
[[ -f bin/etcd && -f bin/kubectl && -f bin/kube-apiserver && `bin/kubectl version -o=yaml 2> /dev/null | bin/yq e '.clientVersion.gitVersion' -` == "v${KUBE_VERSION}" ]] && exit 0
header_text "Installing bin/etcd, bin/kube-apiserver and bin/kubectl ${KUBE_VERSION}"

mkdir -p ./bin
# path until https://github.com/kubernetes-sigs/kubebuilder/issues/1932 will be resolved
#curl -sSLo envtest-bins.tar.gz "https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-${KUBE_VERSION}-$(go env GOOS)-$(go env GOARCH).tar.gz"
curl -sSLo envtest-bins.tar.gz "https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-${KUBE_VERSION}-$(go env GOOS)-${arch}.tar.gz"

tar -zvxf envtest-bins.tar.gz
mv kubebuilder/bin/* bin
rm envtest-bins.tar.gz
rm -r kubebuilder