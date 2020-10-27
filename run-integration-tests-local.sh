#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

# This script sets up a local kind, builds images, deploys the-mesh-for-data to kind and runs integration
# tests against it.
# In order to run this locally an entry for `kind-registry` has to be added to the local /etc/hosts file:
# 127.0.0.1 kind-registry
#
# As the script will add a generated CA certificate to the local OS it needs root or a password to run.

export DOCKER_HOSTNAME=kind-registry:5000
export DOCKER_NAMESPACE=m4d-system
export HELM_EXPERIMENTAL_OCI=1

make kind-setup
make cluster-prepare
make docker-minimal-it
make cluster-prepare-wait
make -C secret-provider configure-vault
make -C secret-provider deploy
make -C manager deploy_it
make -C manager wait_for_manager
make helm
make -C manager run-integration-tests