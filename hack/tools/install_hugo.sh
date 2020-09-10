#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

source ./common.sh

header_text "Checking for bin/hugo"
[[ -f bin/hugo ]] && exit 0

header_text "Installing bin/hugo"

HUGO_VERSION="0.70.0"

case ${arch} in
    amd64)
        arch=64bit
        ;;
esac

case ${os} in
    darwin)
        os=macOS
        ;;
    linux)
        os=Linux
        ;;
esac

pkg=hugo_extended_${HUGO_VERSION}_${os}-${arch}.tar.gz

wget -q https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/$pkg
tar xf $pkg hugo
mv hugo bin/hugo
rm -rf $pkg
