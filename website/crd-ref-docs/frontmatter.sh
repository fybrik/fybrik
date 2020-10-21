#!/usr/bin/env bash

set -u
set -e

file=$1
title=$2
weight=$3

header="---
title: ${title}
weight: ${weight}
---
"

prepend_header() {
    temporary_file=$(mktemp -t tmp.XXXXXXXXXX)
	echo "$header" >> $temporary_file
	cat $file >> $temporary_file
	mv $temporary_file $file
}

prepend_header
