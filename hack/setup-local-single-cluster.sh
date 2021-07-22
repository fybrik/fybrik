# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
# This script is meant for local development with kind

export DOCKER_HOSTNAME=localhost:5000
export DOCKER_NAMESPACE=fybrik-system
export HELM_EXPERIMENTAL_OCI=1
export VALUES_FILE=charts/fybrik/integration-tests.values.yaml

make kind
kubectl config use-context kind-control
make cluster-prepare
make docker-minimal-it
make cluster-prepare-wait
make deploy
make configure-vault
make -C modules helm-chart-push
