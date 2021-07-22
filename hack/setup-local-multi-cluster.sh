# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
# This script is meant for local development with kind

export DOCKER_HOSTNAME=localhost:5000
export DOCKER_NAMESPACE=fybrik-system
export HELM_EXPERIMENTAL_OCI=1
export VALUES_FILE=charts/fybrik/kind-control.values.yaml

make kind-setup-multi
kubectl config use-context kind-control
make -C third_party/razee all

kubectl config use-context kind-control
make -C cluster-prepare
make docker-minimal-it
make cluster-prepare-wait
make deploy
make vault-setup-kind-multi
make -C modules helm-chart-push

export VALUES_FILE=charts/fybrik/kind-kind.values.yaml
kubectl config use-context kind-kind
make -C third_party/cert-manager deploy
make -C third_party/datashim deploy
make cluster-prepare-wait
make deploy

# Switch to control cluster after setup
kubectl config use-context kind-control
