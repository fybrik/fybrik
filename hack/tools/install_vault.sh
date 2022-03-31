#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


cd "${0%/*}"
source ./common.sh

version=1.8.9

header_text "Checking for bin/vault"
[[ -f bin/vault ]] && exit 0

header_text "Installing bin/vault"
mkdir -p ./bin
curl -L -o ./bin/vault.zip https://releases.hashicorp.com/vault/${version}/vault_${version}_${os}_amd64.zip 
unzip ./bin/vault.zip -d ./bin/
rm ./bin/vault.zip

chmod +x ./bin/vault
