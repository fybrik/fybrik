#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

cd "${0%/*}"
source ./common.sh

VERSION_FILE=lib/fzn-or-tools-version

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

DEPLOY="${OR_TOOLS_VERSION}.${OR_TOOLS_BUILD}_for_${arch}_${target_os}"

header_text "Checking for bin/fzn-or-tools ${DEPLOY}"
[[ -f bin/fzn-or-tools ]] && [[ -f "${VERSION_FILE}" ]] && [[ `cat "${VERSION_FILE}"` == "${DEPLOY}" ]] && exit 0
header_text "Installing bin/fzn-or-tools ${DEPLOY}"

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
cp ./bin/fzn-or-tools ./bin/solver
echo ${DEPLOY} > ${VERSION_FILE}
