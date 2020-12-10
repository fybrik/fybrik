#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

# These scripts are there to install/remove razee from a local kind cluster
# They have not been verified against any other K8s system.
# In order to set up a production system please refer to the razee documentation or use
# a managed service such as IBM Cloud Satellite.

op=$1

source ../../hack/tools/common.sh

RAZEE_USER=${RAZEE_USER:="razee-dev@example.com"}
RAZEE_PASSWORD=${RAZEE_PASSWORD:="password123"}

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
    echo "Or use 'make setup_user' in order to create/login a local user"
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

setup_user() {
    # The passwords below are only used for development on local machines and email addresses are fake
    DATA=$(curl --request POST \
      --url http://localhost:3333/graphql \
      --header 'Content-Type: application/json' \
      --data "{\"query\":\"mutation {\n  signUp(\n    username: \\\"razee-dev\\\"\n    email: \\\"$RAZEE_USER\\\"\n    password: \\\"$RAZEE_PASSWORD\\\"\n    orgName: \\\"dev-org\\\"\n    role: \\\"ADMIN\\\"\n  ) {\n    token\n  }\n}\"}")
    if [[ $DATA =~ "E11000" ]]; then
      echo User already exists!
      DATA=$(curl --request POST \
        --url http://localhost:3333/graphql \
        --header 'Content-Type: application/json' \
        --data "{\"query\":\"mutation {\n  signIn(\n    login: \\\"$RAZEE_USER\\\"\n    password: \\\"$RAZEE_PASSWORD\\\"\n  ) {\n    token\n  }\n}\"}")
      TOKEN=$(echo $DATA | jq -r -c ".data.signIn.token")
    else
      TOKEN=$(echo $DATA | jq -r -c ".data.signUp.token")
    fi

    echo $TOKEN > token
}

setup_remotes() {
    RAZEE_USER=$RAZEE_USER RAZEE_PASSWORD=$RAZEE_PASSWORD ./setupCluster.sh kind-control "http://razeedash-api-lb.razee.svc.cluster.local:8081"
    RAZEE_USER=$RAZEE_USER RAZEE_PASSWORD=$RAZEE_PASSWORD ./setupCluster.sh kind-kind "http://control-control-plane:30333"
}

create_razee_secret() {
  # fetch the token
  DATA=$(curl --request POST \
        --url http://localhost:3333/graphql \
        --header 'Content-Type: application/json' \
        --data "{\"query\":\"mutation {\n  signIn(\n    login: \\\"$RAZEE_USER\\\"\n    password: \\\"$RAZEE_PASSWORD\\\"\n  ) {\n    token\n  }\n}\"}")
  TOKEN=$(echo $DATA | jq -r -c ".data.signIn.token")

  ordId=$(curl --request POST \
  --url http://localhost:3333/graphql \
  --header 'content-type: application/json' \
  --header "Authorization: Bearer $TOKEN" \
  --data '{"query":"query {\n  me {\n    orgId \n  }\n}"}')

  echo  "ordId = $ordId"
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
    setup_user)
        header_text "Installing razee on clusters"
        setup_user
        ;;
    create_razee_secret)
        create_razee_secret
        ;;
    *)
        header_text "Installing razee"
        setup_control_cluster
        setup_user
        setup_remotes
        ;;
esac