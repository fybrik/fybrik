#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

: ${RELEASE:=latest}
: ${TOOLBIN:=./hack/tools/bin}

${TOOLBIN}/yq eval --inplace ".version = \"$RELEASE\"" ./charts/m4d/Chart.yaml
${TOOLBIN}/yq eval --inplace ".appVersion = \"$RELEASE\"" ./charts/m4d/Chart.yaml
