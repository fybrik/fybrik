#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


cd "${0%/*}"
source ./common.sh


header_text "Checking for bin/oc ${OC_VERSION}"
[[ -f bin/oc && `bin/oc version --client | awk -F "-" '{print $3}'` == `echo ${OC_VERSION} | cut -d"." -f1,2` ]] && exit 0
header_text "Installing bin/oc ${OC_VERSION}"
mkdir -p ./bin
target_os="$os"
if [[ "$target_os" == "darwin" ]]; then
    target_os="mac"
fi
curl -L -O https://mirror.openshift.com/pub/openshift-v4/clients/ocp/${OC_VERSION}/openshift-client-${target_os}-${OC_VERSION}.tar.gz
trap "rm openshift-client-${target_os}-${OC_VERSION}.tar.gz" err exit
tmp=$(mktemp -d /tmp/openshift-client.XXXXXX)
tar -zxvf ./openshift-client-${target_os}-${OC_VERSION}.tar.gz -C $tmp
mv $tmp/oc ./bin/oc
chmod +x ./bin/oc