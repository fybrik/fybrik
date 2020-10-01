#!/usr/bin/env bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


: ${KUBE_NAMESPACE:=m4d-system}
: ${PORT_TO_FORWARD:=8200}
: ${WITHOUT_PORT_FORWARD:=false}
: ${DATA_PROVIDER_USERNAME:=data_provider}
: ${DATA_PROVIDER_PASSWORD:=password}
: ${KUBERNETES_AUTH_ROLE:=demo}
: ${SECRET_PATH:=secret}

source vault-util.sh

case "$1" in
    configure_path)
      configure_path
    ;;
    populate_demo_secrets)
      populate_demo_secrets
    ;;
    *)
      echo "usage: %0 [configure|populate_demo_secrets]"
      exit 1
    ;;
esac