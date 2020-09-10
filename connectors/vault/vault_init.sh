#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

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
echo "DIR_ROOT is set to $DIR_ROOT"
echo "DIR_ABSTOOLBIN is set to $DIR_ABSTOOLBIN"

echo "policy-compiler.env is at location $ROOT_DIR/pkg/policy-compiler/policy-compiler.env"
source $DIR_ROOT/pkg/policy-compiler/policy-compiler.env

export VAULT_PATH=$DIR_ABSTOOLBIN/vault
echo "VAULT_PATH env variable is set to $VAULT_PATH"
echo "VAULT_SECRET_ENGINE_PATH_FOR_LOCAL_TESTING env variable is set to $VAULT_SECRET_ENGINE_PATH_FOR_LOCAL_TESTING"
echo "VAULT_PATH_FOR_LOCAL_TESTING env variable is set to $VAULT_PATH_FOR_LOCAL_TESTING"
echo "PASSWORD_FOR_VAULT_LOCAL_TESTING env variable  is set to $PASSWORD_FOR_VAULT_LOCAL_TESTING"
echo "USERNAME_FOR_VAULT_LOCAL_TESTING env variable is set to $USERNAME_FOR_VAULT_LOCAL_TESTING"
echo "OWNERID_FOR_VAULT_LOCAL_TESTING env variable is set to $OWNERID_FOR_VAULT_LOCAL_TESTING"

$VAULT_PATH secrets enable -path=$VAULT_SECRET_ENGINE_PATH_FOR_LOCAL_TESTING kv
$VAULT_PATH secrets list -detailed
$VAULT_PATH kv put $VAULT_SECRET_ENGINE_PATH_FOR_LOCAL_TESTING/$VAULT_PATH_FOR_LOCAL_TESTING  username=$USERNAME_FOR_VAULT_LOCAL_TESTING password=$PASSWORD_FOR_VAULT_LOCAL_TESTING ownerId=$OWNERID_FOR_VAULT_LOCAL_TESTING
$VAULT_PATH kv get $VAULT_SECRET_ENGINE_PATH_FOR_LOCAL_TESTING/$VAULT_PATH_FOR_LOCAL_TESTING
