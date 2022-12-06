#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


cd "${0%/*}"
source ./common.sh


case ${os} in
    linux)
        target_os=ubuntu-18.04
        dyn_lib_ext=so
        ;;
    darwin)
        arch=`uname -m`
        target_os=macOS-13.0.1
        dyn_lib_ext=dylib
        ;;
esac

header_text "Checking for bin/fzn-or-tools ${OR_TOOLS_VERSION}.${OR_TOOLS_BUILD}"
[[ -f bin/fzn-or-tools ]] && exit 0

header_text "Installing bin/fzn-or-tools ${OR_TOOLS_VERSION}.${OR_TOOLS_BUILD}"
mkdir -p ./bin
mkdir -p ./lib

download_file=or-tools_${arch}_${target_os}_cpp_v${OR_TOOLS_VERSION}.${OR_TOOLS_BUILD}.tar.gz
curl -L -O https://github.com/google/or-tools/releases/download/v${OR_TOOLS_VERSION}/${download_file}
trap "rm ${download_file}" err exit
tmp=$(mktemp -d /tmp/or-tools.XXXXXX)
tar -zxvf ./${download_file} -C $tmp
mv $tmp/*/bin/fzn-ortools ./bin/fzn-or-tools
mv $tmp/*/lib/lib*.${dyn_lib_ext}* ./lib
chmod +x ./bin/fzn-or-tools
