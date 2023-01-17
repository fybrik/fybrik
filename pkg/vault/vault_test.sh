#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


: ${KUBE_NAMESPACE:=fybrik-system}
: ${PORT_TO_FORWARD:=8204}

port_forward() {
  # Port forward, so we could access vault
  echo "Creating a port-forward from $PORT_TO_FORWARD to 8200 for Vault"
  kubectl port-forward -n $KUBE_NAMESPACE service/vault "$PORT_TO_FORWARD":8200 &
  while ! nc -z localhost "$PORT_TO_FORWARD"; do echo "Waiting for the port-forward from $PORT_TO_FORWARD to 8200 to take effect"; sleep 1; done
}

test() {
  port_forward
  # Get Vault's root token
  export VAULT_TOKEN=$(kubectl get secrets vault-credentials -n $KUBE_NAMESPACE -o jsonpath={.data.VAULT_TOKEN} | base64 --decode)
  export VAULT_ADDRESS="http://127.0.0.1:$PORT_TO_FORWARD/"
  go test $(TEST_OPTIONS) vault_interface.go vault_impl.go vault_dummy.go vault_interface_test.go
  # Kill the port-forward if nessecarry 
  kill -9 %%
}

case "$1" in
    test)
      test
    ;;
    *)
      echo "usage: %0 [test]"
      exit 1
    ;;
esac



