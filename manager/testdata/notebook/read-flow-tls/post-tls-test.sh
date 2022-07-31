#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

export ROOT_DIR=../../../../

# restore the original Dockerfiles
cp katalog-connector-Dockerfile.orig ${ROOT_DIR}/connectors/katalog/Dockerfile
cp opa-connector-Dockerfile.orig ${ROOT_DIR}/connectors/opa/Dockerfile
cp manager-Dockerfile.orig ${ROOT_DIR}/manager/Dockerfile

