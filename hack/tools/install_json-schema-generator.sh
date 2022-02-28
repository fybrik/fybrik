#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

cd "${0%/*}"
source ./common.sh

case ${arch} in
    amd64)
        arch=x86_64
        ;;
esac

version=2.3.1
TARGET_VERSION=0.2.1

header_text "Checking for bin/json-schema-generator"
[[ -f bin/json-schema-generator && `bin/json-schema-generator -v | awk 'NF>1{print $NF}'` == ${TARGET_VERSION} ]] && exit 0

header_text "Installing bin/json-schema-generator"
mkdir -p ./bin



curl -sSLo json-schema-generator.tar.gz "https://github.com/fybrik/json-schema-generator/releases/download/v${TARGET_VERSION}/json-schema-generator_${TARGET_VERSION}_${os}_${arch}.tar.gz" 

mkdir -p json-schema-generator
tar -C json-schema-generator -zvxf json-schema-generator.tar.gz

mv json-schema-generator/json-schema-generator bin/
rm json-schema-generator.tar.gz
rm -r json-schema-generator