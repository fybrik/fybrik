#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


cd "${0%/*}"
source ./common.sh


case ${os} in
    linux)
        target_os=ubuntu-18.04
        ;;
    darwin)
        arch=""
        target_os=MacOsX-12.2.1
        ;;
esac

header_text "Checking for bin/fzn-or-tools ${OR_TOOLS_VERSION}.${OR_TOOLS_BUILD}"
[[ -f bin/fzn-or-tools ]] && exit 0

header_text "Installing bin/fzn-or-tools ${OR_TOOLS_VERSION}.${OR_TOOLS_BUILD}"
mkdir -p ./bin
mkdir -p ./lib

download_file=or-tools_${arch}_flatzinc_${target_os}_v${OR_TOOLS_VERSION}.${OR_TOOLS_BUILD}.tar.gz
curl -L -O https://github.com/google/or-tools/releases/download/v${OR_TOOLS_VERSION}/${download_file}
trap "rm ${download_file}" err exit
tmp=$(mktemp -d /tmp/or-tools.XXXXXX)
tar -zxvf ./${download_file} -C $tmp
mv $tmp/*/bin/fzn-or-tools ./bin
mv $tmp/*/lib/lib*.so ./lib
chmod +x ./bin/fzn-or-tools
