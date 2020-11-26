#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

# Usage: ./removeCluster.sh <cluster name>

if [ -z "$APIKEY" ]; then
  echo "You need to supplied an APIKEY for razee!"
  exit
fi

CLUSTERNAME=$1

echo Removing cluster $CLUSTERNAME

cat razeedeploy-delta-remove-template.yaml | envsubst | kubectl apply --context $CLUSTERNAME -f -

# Graceful deletion by starting razee-deploy remove job
kubectl wait job/razeedeploy-job-remove -n razeedeploy --for=condition=complete --context $CLUSTERNAME --timeout 60s

ORGANIZATIONS=$(curl --request POST \
  --url http://localhost:3333/graphql \
  --header 'content-type: application/json' \
  --header "x-api-key: $APIKEY" \
  --data '{"query":"query {organizations {\n  id\n  name\n}}"}')

ORGID=$(echo $ORGANIZATIONS | jq -r -c ".data.organizations[0].id")

CLUSTERIDRES=$(curl --request POST \
  --url http://localhost:3333/graphql \
  --header 'content-type: application/json' \
  --header "x-api-key: $APIKEY" \
  --data "{\"query\":\"query {clusterByName(orgId: \\\"$ORGID\\\", clusterName: \\\"$CLUSTERNAME\\\") {\n  clusterId\n}}\"}")

CLUSTERID=$(echo $CLUSTERIDRES | jq -r .data.clusterByName.clusterId)

curl --request POST \
  --url http://localhost:3333/graphql \
  --header 'content-type: application/json' \
  --header "x-api-key: $APIKEY" \
  --data "{\"query\":\"mutation {deleteClusterByClusterId(orgId: \\\"$ORGID\\\", clusterId: \\\"$CLUSTERID\\\"){\n  deletedClusterCount\n  deletedResourceCount\n}}\"}"

kubectl delete ns razeedeploy --context $CLUSTERNAME