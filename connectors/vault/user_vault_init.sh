# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

#This file configures the some local vault to act as "user vault" - user deployed vault that stores there credentials to data source for data assets
#We also put some demo credentials for the assets we check
if [ $# -eq 0 ]
  then
    echo "root dir not specified as an argument"
    echo "abstoolbin dir not specified as an argument"
    exit 1
fi
if [ $# -eq 1 ]
  then
    echo "abstoolbin dir not specified as an argument"
    exit 1
fi
DIR_ROOT=$1
DIR_ABSTOOLBIN=$2

echo "policy-compiler.env is at location $ROOT_DIR/pkg/policy-compiler/policy-compiler.env"
source $DIR_ROOT/pkg/policy-compiler/policy-compiler.env

export VAULT_BIN=$DIR_ABSTOOLBIN/vault
#export USER_VAULT_KAFKA_ASSET_KEY=87ffdca3-8b5d-4f77-99f9-0cb1fba1f73f/01c6f0f0-9ffe-4ccc-ac07-409523755e72
export USER_VAULT_KAFKA_ASSET_KEY="{\"catalog_id\":\"87ffdca3-8b5d-4f77-99f9-0cb1fba1f73f\",\"asset_id\":\"01c6f0f0-9ffe-4ccc-ac07-409523755e72\"}"
export USER_VAULT_KAFKA_ASSET_CRED="my_kafka_credentials"

# export USER_VAULT_EGERIA_ASSET_KEY="{\"ServerName\":\"cocoMDS3\",\"AssetGuid\":\"24cd3ed9-4084-43b9-9e91-5fe1f4fbd6b7\"}"
export USER_VAULT_EGERIA_ASSET_KEY="{\"ServerName\":\"cocoMDS3\",\"AssetGuid\":\"1e2a0403-1946-4e89-a10b-fd96eda5a5dc\"}"
export USER_VAULT_EGERIA_ASSET_CRED=@creds.json

echo "VAULT_BIN env variable is set to $VAULT_BIN"
echo "USER_VAULT_PATH env variable is set to $USER_VAULT_PATH"
echo "USER_VAULT_KAFKA_ASSET_KEY env variable is set to $USER_VAULT_KAFKA_ASSET_KEY"
echo "USER_VAULT_KAFKA_ASSET_CRED env variable  is set to $USER_VAULT_KAFKA_ASSET_CRED"

echo "Executing: $VAULT_BIN secrets enable -address=$USER_VAULT_ADDRESS  -path=$USER_VAULT_PATH kv"
$VAULT_BIN secrets enable -address=$USER_VAULT_ADDRESS   -path=$USER_VAULT_PATH kv
$VAULT_BIN secrets list -address=$USER_VAULT_ADDRESS   -detailed
$VAULT_BIN kv put -address=$USER_VAULT_ADDRESS   $USER_VAULT_PATH/$USER_VAULT_KAFKA_ASSET_KEY  credentials=$USER_VAULT_KAFKA_ASSET_CRED
$VAULT_BIN kv get -address=$USER_VAULT_ADDRESS   $USER_VAULT_PATH/$USER_VAULT_KAFKA_ASSET_KEY

#$VAULT_BIN kv put -address=$USER_VAULT_ADDRESS   $USER_VAULT_PATH/$USER_VAULT_EGERIA_ASSET_KEY  credentials=$USER_VAULT_EGERIA_ASSET_CRED
$VAULT_BIN kv put -address=$USER_VAULT_ADDRESS   $USER_VAULT_PATH/$USER_VAULT_EGERIA_ASSET_KEY  @creds.json
$VAULT_BIN kv get -address=$USER_VAULT_ADDRESS   $USER_VAULT_PATH/$USER_VAULT_EGERIA_ASSET_KEY