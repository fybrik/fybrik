#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


cd "${0%/*}"
source ./common.sh


header_text "Checking for bin/opa ${OPA_VERSION}"
[[ -f bin/opa && `bin/opa version | cut -f2 -d" " | head -n1` == ${OPA_VERSION} ]] && exit 0

header_text "Installing bin/opa ${OPA_VERSION}"
mkdir -p ./bin
curl -L -o ./bin/opa https://openpolicyagent.org/downloads/v${OPA_VERSION}/opa_${os}_amd64 
chmod +x ./bin/opa
