#!/usr/bin/env bash

# usage: ./create_new_asset.sh  <asset_json>  <tags to be added>
# asset_json is a json file that should describe newly added asset metadata.
# Structure of asset_json:
# {
#     "class": "NewCSVFileAssetRequestBody",                // this field cannot be changed to anything else
#     "displayName": "<name of the asset>",
#     "description": "<free text description>",
#     "fullPath": "<pointer to external storage>",			// this is json structure that points to the external storage when the asset itself stored, will explain this structure separatly
#     "columnHeaders": [ "column1", "column2" ... ]			// list of all columns of the asset
# }
# We have an assumption that asset is stored in an external storage system. Today we support 3 storage types: S3, DB2, KAFKA
# These are the fields that should be present in the Json describing external storage.
# This json will be retrieve by Mesh for Data and used to access the data, so it is important to follow the exact structure described here.
# DB2:
# 			data_store : "DB2"		//required
#          	url
# 			database
#     		table
#     		port
#     		ssl = 5;          		// "true" or "false"

# S3:
# 			data_store : "S3"		//required
# 			endpoint
# 			bucket
# 			object_key				//can be object name or the prefix for dataset
# 			region					//optional

# Kafka:
# 			data_store : "Kafka"	//required
# 			topic_name
#     		bootstrap_servers
#     		schema_registry
#     		key_deserializer
#     		value_deserializer
#     		security_protocol
#     		sasl_mechanism
#     		ssl_truststore
#     		ssl_truststore_password

# Second parameter given to create_new_asset.sh script are Tags that are added to the asset in format of a list: 'tag1 tag2 tag3'
# For each tag_i, a new inforamtion-tag will be created with name "tag_i" as passed here and this tag will be added to the new asset
# This tag can be used in the governance policies
# Today Egeria supports only csv files creation using REST API, it should change soon, but today only csv files assets can be created (if you point to another file it will work but data-mesh will categorize it as csv file)


ASSET_JSON=$1
TAGS=$2

echo "Create new asset from json: $ASSET_JSON"
echo "Create Tags and add them to the asset: $TAGS"

USER=calliequartile
: ${EGERIA_URL:=localhost:9443}

echo "EGERIA_URL: $EGERIA_URL"


GUID=$(curl -k -X POST "https://${EGERIA_URL}/servers/cocoMDS3/open-metadata/access-services/asset-owner/users/${USER}/assets/data-files/csv" -H  "accept: */*" -H  "Content-Type: application/json"  --data "@$ASSET_JSON" | jq -r ".guids[-1]" )

echo "GUID:"  $GUID

echo https://${EGERIA_URL}/servers/cocoMDS3/open-metadata/access-services/asset-consumer/users/${USER}/tags

for TAG in $TAGS
do
	echo "TAG: $TAG"
	
	TAG_JSON="{\"public\": true, \"class\": \"TagRequestBody\",\"tagName\": \"$TAG\",\"tagDescription\": \"$TAG related data\"}"
	echo "JSON: $TAG_JSON"
	echo ""
	
	TAG_GUID=$(curl -k -X POST "https://${EGERIA_URL}/servers/cocoMDS3/open-metadata/access-services/asset-consumer/users/${USER}/tags" -H  "accept: */*" -H  "Content-Type: application/json"  --data "$TAG_JSON" | jq -r ".guid" )

	echo "TAG GUID: $TAG_GUID"
	echo "Adding this tag to asset using URL:"
	echo "https://${EGERIA_URL}/servers/cocoMDS3/open-metadata/access-services/asset-consumer/users/${USER}/assets/${GUID}/tags/${TAG_GUID}"

	curl -k -X POST "https://${EGERIA_URL}/servers/cocoMDS3/open-metadata/access-services/asset-consumer/users/${USER}/assets/${GUID}/tags/${TAG_GUID}" -H  "accept: */*" -H  "Content-Type: application/json"  --data "{\"public\":true, \"class\":\"FeedbackRequestBody\"}"
	echo " "
	echo "Added tag $TAG to the asset"
done

echo " "
echo "------------ READ ---------------"

./read_asset.sh $GUID
