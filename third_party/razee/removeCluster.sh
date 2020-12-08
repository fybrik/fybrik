#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

# Usage: ./removeCluster.sh <cluster name>

AUTHENTICATION_HEADER=""

if [ ! -f token ]; then
  if [ -z "$APIKEY" ]; then
    echo "You need to supply an APIKEY for razee as environment variable!"
    exit
  else
    echo Using APIKEY for authentication
    AUTHENTICATION_HEADER="x-api-key: $APIKEY"
  fi
else
  echo Using token for authentication
  TOKEN=$(cat token)
  AUTHENTICATION_HEADER="Authorization: Bearer $TOKEN"
fi

if [ -z "$AUTHENTICATION_HEADER" ]; then
  echo "You need to supply an APIKEY as enivonrment variable or a local token file!"
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
  --header "$AUTHENTICATION_HEADER" \
  --data '{"query":"query {organizations {\n  id\n  name\n}}"}')

ORGID=$(echo $ORGANIZATIONS | jq -r -c ".data.organizations[0].id")

CLUSTERIDRES=$(curl --request POST \
  --url http://localhost:3333/graphql \
  --header 'content-type: application/json' \
  --header "$AUTHENTICATION_HEADER" \
  --data "{\"query\":\"query {clusterByName(orgId: \\\"$ORGID\\\", clusterName: \\\"$CLUSTERNAME\\\") {\n  clusterId\n}}\"}")

CLUSTERID=$(echo $CLUSTERIDRES | jq -r .data.clusterByName.clusterId)

curl --request POST \
  --url http://localhost:3333/graphql \
  --header 'content-type: application/json' \
  --header "$AUTHENTICATION_HEADER" \
  --data "{\"query\":\"mutation {deleteClusterByClusterId(orgId: \\\"$ORGID\\\", clusterId: \\\"$CLUSTERID\\\"){\n  deletedClusterCount\n  deletedResourceCount\n}}\"}"

kubectl delete ns razeedeploy --context $CLUSTERNAME