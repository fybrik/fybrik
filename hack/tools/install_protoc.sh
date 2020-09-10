#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


source ./common.sh

case ${arch} in
    amd64)
        arch=x86_64
        ;;
esac

case ${os} in
    darwin)
        os=osx
        ;;
esac

version=3.7.1

header_text "Checking for bin/protoc"
[[ -f bin/protoc ]] && exit 0

header_text "Installing bin/protoc"
mkdir -p ./bin

PROTOC_ZIP=protoc-${version}-${os}-${arch}.zip
curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v${version}/$PROTOC_ZIP
unzip -o $PROTOC_ZIP -d . bin/protoc
rm -f $PROTOC_ZIP
