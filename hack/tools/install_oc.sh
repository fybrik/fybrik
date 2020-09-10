#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


cd "${0%/*}"
source ./common.sh

version=4.2.19

header_text "Checking for bin/oc"
[[ -f bin/oc ]] && exit 0

header_text "Installing bin/oc"
mkdir -p ./bin
target_os="$os"
if [[ "$target_os" == "darwin" ]]; then
    target_os="mac"
fi
curl -L -O https://mirror.openshift.com/pub/openshift-v4/clients/ocp/${version}/openshift-client-${target_os}-${version}.tar.gz
trap "rm openshift-client-${target_os}-${version}.tar.gz" err exit
tmp=$(mktemp -d /tmp/openshift-client.XXXXXX)
tar -zxvf ./openshift-client-${target_os}-${version}.tar.gz -C $tmp
mv $tmp/oc ./bin/oc
chmod +x ./bin/oc