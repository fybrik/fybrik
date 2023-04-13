# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
# This script is meant for local development with kind

set -x

export DOCKER_HOSTNAME=localhost:5000
export DOCKER_NAMESPACE=fybrik-system
export VALUES_FILE=charts/fybrik/kind-control.values.yaml
export VAULT_VALUES_FILE=charts/vault/env/dev/local-multi-cluster.yaml
export HELM_SETTINGS="--set "coordinator.catalog=katalog""
export DEPLOY_OPENMETADATA_SERVER=0

make kind-setup-multi
kubectl config use-context kind-control
make -C third_party/razee all

kubectl config use-context kind-control
make cluster-prepare
make docker-build
make docker-push
make cluster-prepare-wait
make deploy-fybrik

export VALUES_FILE=charts/fybrik/kind-kind.values.yaml
kubectl config use-context kind-kind
make -C third_party/cert-manager deploy
make deploy-fybrik

# Switch to control cluster after setup
kubectl config use-context kind-control

make vault-setup-kind-multi
