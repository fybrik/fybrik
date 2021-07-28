#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

source ./common.sh

version=1.1.3

header_text "Checking for bin/vangen"
[[ -f bin/vangen ]] && exit 0

header_text "Installing bin/vangen"
mkdir -p ./bin

VANGEN_TAR_GZ=vangen_${version}_${os}_${arch}.tar.gz 
curl -OL https://github.com/leighmcculloch/vangen/releases/download/v${version}/$VANGEN_TAR_GZ
tar -C bin -xzf $VANGEN_TAR_GZ vangen
rm -f $VANGEN_TAR_GZ