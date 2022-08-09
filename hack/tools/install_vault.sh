#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


cd "${0%/*}"
source ./common.sh


header_text "Checking for bin/vault ${VAULT_VERSION}"
[[ -f bin/vault &&  `bin/vault -v | awk '{print $2}'` == "v${VAULT_VERSION}" ]] && exit 0

header_text "Installing bin/vault ${VAULT_VERSION}"
mkdir -p ./bin
curl -L -o ./bin/vault.zip https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_${os}_amd64.zip
unzip -o ./bin/vault.zip -d ./bin/
rm ./bin/vault.zip

chmod +x ./bin/vault
