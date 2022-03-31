#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


cd "${0%/*}"
source ./common.sh

DESIRED_VERSION=1.8.9

header_text "Checking for bin/vault $DESIRED_VERSION"
[[ -f bin/vault &&  `bin/vault -v | awk '{print $2}'` == "v$DESIRED_VERSION" ]] && exit 0

header_text "Installing bin/vault $DESIRED_VERSION"
mkdir -p ./bin
curl -L -o ./bin/vault.zip https://releases.hashicorp.com/vault/${DESIRED_VERSION}/vault_${DESIRED_VERSION}_${os}_amd64.zip
unzip -o ./bin/vault.zip -d ./bin/
rm ./bin/vault.zip

chmod +x ./bin/vault
