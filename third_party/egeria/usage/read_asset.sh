#!/usr/bin/env bash

# usage: ./read_asset.sh <Egeria asset GUID>

GUID=$1

echo "Read asset info"

echo "ASSET_ID: $GUID"

USER=garygeeke
: ${EGERIA_URL:=localhost:9443}


curl -k -X GET "https://${EGERIA_URL}/servers/mds1/open-metadata/common-services/asset-owner/connected-asset/users/$USER/assets/$GUID" -H  "accept: */*" -H  "Content-Type: application/json"

SCHEMA_GUID=$(curl -k -X GET "https://${EGERIA_URL}/servers/mds1/open-metadata/common-services/asset-owner/connected-asset/users/$USER/assets/$GUID" -H  "accept: */*" -H  "Content-Type: application/json" | jq -r ".schemaType.guid" )

echo "SCHEMA_GUID:"  $SCHEMA_GUID

COLUMNS=$(curl -k -X GET "https://${EGERIA_URL}/servers/mds1/open-metadata/common-services/asset-owner/connected-asset/users/$USER/assets/schemas/$SCHEMA_GUID/schema-attributes?elementStart=0&maxElements=0" -H  "accept: */*" -H  "Content-Type: application/json" | jq -r ".list[].attributeName" )

echo "curl -k -X GET \"https://${EGERIA_URL}/servers/mds1/open-metadata/common-services/asset-owner/connected-asset/users/$USER/assets/schemas/$SCHEMA_GUID/schema-attributes?elementStart=0&maxElements=0\""
echo "COLUMNS:"  $COLUMNS


TAGS_READ=$(curl -k -X GET "https://${EGERIA_URL}/servers/mds1/open-metadata/common-services/asset-consumer/connected-asset/users/$USER/assets/$GUID/informal-tags?elementStart=0&maxElements=20" -H  "accept: */*" -H  "Content-Type: application/json" | jq -r ".list[].name" )

echo "TAGS:"  $TAGS_READ

