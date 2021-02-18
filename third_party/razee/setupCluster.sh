#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

# Usage: ./registerCluster.sh <cluster name>

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

if [ -z "$2" ]; then
  export RAZEEDASH_API="http://control-control-plane:30333"
else
  export RAZEEDASH_API=$2
fi

kubectl config use-context $CLUSTERNAME

echo Creating cluster entry for $CLUSTERNAME

ORGANIZATIONS=$(curl --request POST \
  --url http://localhost:3333/graphql \
  --header 'content-type: application/json' \
  --header "$AUTHENTICATION_HEADER" \
  --data '{"query":"query {organizations {\n  id\n  name\n}}"}')

ORGID=$(echo $ORGANIZATIONS | jq -r -c ".data.organizations[0].id")

CURLRESULT=$(curl --request POST \
  --url http://localhost:3333/graphql \
  --header 'content-type: application/json' \
  --header "$AUTHENTICATION_HEADER" \
  --data "{\"query\":\"mutation (\$registration: JSON!){\n  registerCluster(orgId: \\\"$ORGID\\\", registration: \$registration) {\n    url\n    orgId\n    orgKey\n    clusterId\n    regState\n    registration\n  }\n}\n\",\"variables\":{\"registration\":{\"name\":\"$CLUSTERNAME\"}}}")

export ORGAPIKEY=$(echo $CURLRESULT | jq .data.registerCluster.orgKey -r)
export CLUSTERID=$(echo $CURLRESULT | jq .data.registerCluster.clusterId -r)

kubectl create ns m4d-system || true
kubectl -n m4d-system create secret generic razee-credentials --from-literal=RAZEE_URL="$RAZEEDASH_API/graphql" --from-literal=RAZEE_USER="$RAZEE_USER" --from-literal=RAZEE_PASSWORD=$RAZEE_PASSWORD

if [ $ORGAPIKEY == "null" ]; then
  echo "Could not register cluster!"
  exit
fi

export CLUSTERNAMEB64=$(echo "{\"name\": \"$CLUSTERNAME\"}" | base64)

cat razeedeploy-delta-install-template.yaml | envsubst | kubectl apply --context $CLUSTERNAME -f -



