#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

oc adm policy add-scc-to-user privileged -z default 
helm install lab odpi-egeria-lab -f lab.yaml