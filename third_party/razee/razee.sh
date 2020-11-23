#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

# These scripts are there to install/remove razee from a local kind cluster
# They have not been verified against any other K8s system.
# In order to set up a production system please refer to the razee documentation or use
# a managed service such as IBM Cloud Satellite.

op=$1

source ../../hack/tools/common.sh

setup_control_cluster() {
    kubectl config use-context kind-control
    # Install razee dash API and UI
    kubectl apply -f razeedash-all-in-one.yaml
    kubectl apply -f razeedash.yaml
    kubectl apply -f razeedash-api.yaml
    kubectl apply -f razee-nodeports.yaml
    ./kc_create_razeedash_config_map.sh || true
    kubectl wait --for=condition=available -n razee deployment/mongo --timeout=180s
    kubectl wait --for=condition=available -n razee deployment/razeedash-api --timeout=180s
    kubectl wait --for=condition=available -n razee deployment/razeedash --timeout=180s
    echo "Please follow Step 7 of the Razee Documentation to set up authentication https://github.com/razee-io/Razee/blob/master/README.md#installing-razeedash"
    echo "Once done please export the api key with 'export APIKEY=mykey'"
}

delete_razee() {
    kubectl delete ns razee --context kind-control &
    kubectl -n razee delete pv mongo-pv-volume --context kind-control &
    delete_razee_remotes &
    wait
}

delete_razee_remotes() {
    ./removeCluster.sh kind-control &
    ./removeCluster.sh kind-kind &
    wait
}

setup_remotes() {
    ./setupCluster.sh kind-control "http://razeedash-api-lb.razee.svc.cluster.local:8081/api/v2"
    ./setupCluster.sh kind-kind "http://control-control-plane:30333/api/v2"
}

case "$op" in
    cleanup)
        header_text "Uninstalling razee from clusters"
        delete_razee
        ;;
    cleanup_remotes)
        header_text "Uninstalling razee from remote clusters"
        delete_razee_remotes
        ;;
    install_control)
        header_text "Installing razee control cluster"
        setup_control_cluster
        ;;
    setup_remotes)
        header_text "Installing razee on clusters"
        setup_remotes
        ;;
    *)
        header_text "Installing razee"
        setup_control_cluster
        setup_remotes
        ;;
esac