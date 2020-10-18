#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

: ${ROOT_DIR=:../../}
: ${KUBE_NAMESPACE:=m4d-system}
: ${PORT_TO_FORWARD:=8200}
: ${WITHOUT_PORT_FORWARD:=false}

source $ROOT_DIR/third_party/vault/vault-util.sh

populate_demo_secrets() {
  # Push some secrets, assuming the user has exported the APIKEY
  export VAULT_TOKEN=$(kubectl get secrets vault-unseal-keys -n $KUBE_NAMESPACE -o jsonpath={.data.vault-root} | base64 --decode)
  echo "pushing some secrets to vault"
  push_secret "secret/cos" '{"api_key":"'"$APIKEY"'"}' $VAULT_TOKEN
  push_secret "secret/fake-key" '{"api_key":"abcdefgh12345678"}' $VAULT_TOKEN
  push_secret "secret/db2y" '{"password":"s3cr3t", "username":"user1"}' $VAULT_TOKEN
  push_secret "secret/some-secret" '{"password":"pass-pass", "username":"data-provider-1"}' $VAULT_TOKEN
  #push_secret "external/some-secret" '{"password":"pass-pass", "username":"data-provider-2"}' $VAULT_TOKEN

  #Equivalent using the CLI:
  #vault kv put secret/cos api_key=$APIKEY
  #vault kv put secret/fake-key api_key=abcdefgh12345678
  #vault kv put secret/db2 username=user1 password=s3cr3t
  #vault kv put secret/some-secret username=data-provider password=pass-pass
  #vault kv put external/some-secret username=data-provider password=pass-pass
}

case "$1" in
    populate_demo_secrets)
      populate_demo_secrets
    ;;
    *)
      echo "usage: %0 [populate_demo_secrets]"
      exit 1
    ;;
esac