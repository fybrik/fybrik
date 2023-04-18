# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
# This script is meant for local development with kind

export DOCKER_HOSTNAME=localhost:5000
export DOCKER_NAMESPACE=fybrik-system
export VALUES_FILE=test/charts/integration-tests.values.yaml
export HELM_SETTINGS="--set "coordinator.catalog=katalog""
export DEPLOY_OPENMETADATA_SERVER=0

make kind
kubectl config use-context kind-control
make cluster-prepare
make docker-build docker-push
make -C test/services docker-build docker-push
make cluster-prepare-wait
make deploy-fybrik
make configure-vault
