#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


cd "${0%/*}"
source ./common.sh

BUILD_FILE=fzn-or-tools-build
VERSION_FILE=fzn-or-tools-version

case ${os} in
    linux)
        dyn_lib_ext=so
        target_os=ubuntu-20.04
        if [ -f /etc/redhat-release ]; then
            target_os=centos-8
        fi
        ;;
    darwin)
        arch=`uname -m`
        dyn_lib_ext=dylib
        target_os=macOS-13.0.1
        ;;
esac

header_text "Checking for bin/fzn-or-tools ${OR_TOOLS_VERSION}.${OR_TOOLS_BUILD}"
# [[ -f bin/fzn-or-tools ]] && [[ -f "${BUILD_FILE}" ]] && [[ `cat "${BUILD_FILE}"` == "${OR_TOOLS_BUILD}" ]] && [[ -f "${VERSION_FILE}" ]] && [[ `cat "${VERSION_FILE}"` == "${OR_TOOLS_VERSION}" ]] && exit 0

header_text "Installing bin/fzn-or-tools ${OR_TOOLS_VERSION}.${OR_TOOLS_BUILD}"
mkdir -p ./bin
mkdir -p ./lib

download_file=or-tools_${arch}_${target_os}_cpp_v${OR_TOOLS_VERSION}.${OR_TOOLS_BUILD}.tar.gz
curl -L -O https://github.com/google/or-tools/releases/download/v${OR_TOOLS_VERSION}/${download_file}
trap "rm ${download_file}" err exit
tmp=$(mktemp -d /tmp/or-tools.XXXXXX)
tar -zxvf ./${download_file} -C $tmp
mv $tmp/*/bin/fzn-ortools ./bin/fzn-or-tools
mv $tmp/*/lib*/lib*.${dyn_lib_ext}* ./lib
chmod +x ./bin/fzn-or-tools
# echo ${OR_TOOLS_BUILD} > ${BUILD_FILE}
# echo ${OR_TOOLS_VERSION} > ${VERSION_FILE}
