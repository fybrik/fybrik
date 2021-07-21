#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


cd "${0%/*}"
source ./common.sh

header_text "Checking for bin/openapi-generator-cli"
[[ -f bin/openapi-generator-cli ]] && exit 0

header_text "Installing bin/openapi-generator-cli"
mkdir -p ./bin
curl -L -o ./bin/openapi-generator-cli https://raw.githubusercontent.com/OpenAPITools/openapi-generator/master/bin/utils/openapi-generator-cli.sh
chmod +x ./bin/openapi-generator-cli
