#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


cd "${0%/*}"
source ./common.sh

version=4.2.19

header_text "Checking for bin/opa"
[[ -f bin/opa ]] && exit 0

header_text "Installing bin/opa"
mkdir -p ./bin
curl -L -o ./bin/opa https://openpolicyagent.org/downloads/latest/opa_${os}_amd64 
chmod +x ./bin/opa
