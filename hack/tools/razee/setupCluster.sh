#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

# Usage: ./registerCluster.sh <cluster name>

if [ -z "$APIKEY" ]; then
  echo "You need to supply an APIKEY for razee as environment variable!"
  exit
fi

CLUSTERNAME=$1

if [ -z "$2" ]; then
  export RAZEEDASH_URL="http://control-control-plane:30333/api/v2"
else
  export RAZEEDASH_URL=$2
fi

bin/kubectl config use-context $CLUSTERNAME

echo Creating cluster entry for $CLUSTERNAME

ORGANIZATIONS=$(curl --request POST \
  --url http://localhost:3333/graphql \
  --header 'content-type: application/json' \
  --header "x-api-key: $APIKEY" \
  --data '{"query":"query {organizations {\n  id\n  name\n}}"}')

ORGID=$(echo $ORGANIZATIONS | jq -r -c ".data.organizations[0].id")

CURLRESULT=$(curl --request POST \
  --url http://localhost:3333/graphql \
  --header 'content-type: application/json' \
  --header "x-api-key: $APIKEY" \
  --data "{\"query\":\"mutation (\$registration: JSON!){\n  registerCluster(orgId: \\\"$ORGID\\\", registration: \$registration) {\n    url\n    orgId\n    orgKey\n    clusterId\n    regState\n    registration\n  }\n}\n\",\"variables\":{\"registration\":{\"name\":\"$CLUSTERNAME\"}}}")

export ORGAPIKEY=$(echo $CURLRESULT | jq .data.registerCluster.orgKey -r)
export CLUSTERID=$(echo $CURLRESULT | jq .data.registerCluster.clusterId -r)

if [ $ORGAPIKEY == "null" ]; then
  echo "Could not register cluster!"
  exit
fi

export CLUSTERNAMEB64=$(echo "{\"name\": \"$CLUSTERNAME\"}" | base64)

if [ -f razeedeploy-delta-install-template.yaml ]; then
  cat razeedeploy-delta-install-template.yaml | envsubst | kubectl apply --context $CLUSTERNAME -f -
else
  cat razee/razeedeploy-delta-install-template.yaml | envsubst | kubectl apply --context $CLUSTERNAME -f -
fi



