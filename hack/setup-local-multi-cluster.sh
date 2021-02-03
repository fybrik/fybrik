# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0
# This script is meant for local development with kind

export DOCKER_HOSTNAME=localhost:5000
export DOCKER_NAMESPACE=m4d-system
export HELM_EXPERIMENTAL_OCI=1

make kind-setup-multi
kubectl config use-context kind-control
make -C third_party/razee all

kubectl config use-context kind-control
make cluster-prepare
make docker-minimal-it
make cluster-prepare-wait
make -C secret-provider configure-vault
make -C secret-provider deploy
make -C manager deploy-crd
make -C manager deploy_multi_it
make -C manager wait_for_manager
make -C modules helm-chart-push

kubectl config use-context kind-kind
make cluster-prepare
make cluster-prepare-wait
make -C secret-provider configure-vault
make -C secret-provider deploy
make -C manager deploy-crd
make -C manager deploy_multi_it
make -C manager wait_for_manager

# Switch to control cluster after setup
kubectl config use-context kind-control