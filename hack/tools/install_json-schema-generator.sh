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


header_text "Checking for bin/json-schema-generator ${JSON_SCHEMA_GENERATOR_VERSION}"
[[ -f bin/json-schema-generator && `bin/json-schema-generator -v | awk 'NF>1{print $NF}'` == ${JSON_SCHEMA_GENERATOR_VERSION} ]] && exit 0

header_text "Installing bin/json-schema-generator ${JSON_SCHEMA_GENERATOR_VERSION}"
mkdir -p ./bin



curl -sSLo json-schema-generator.tar.gz "https://github.com/fybrik/json-schema-generator/releases/download/v${JSON_SCHEMA_GENERATOR_VERSION}/json-schema-generator_${JSON_SCHEMA_GENERATOR_VERSION}_${os}_${arch}.tar.gz" 

mkdir -p json-schema-generator
tar -C json-schema-generator -zvxf json-schema-generator.tar.gz

mv json-schema-generator/json-schema-generator bin/
rm json-schema-generator.tar.gz
rm -r json-schema-generator