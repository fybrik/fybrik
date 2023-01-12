#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0


kubectl get secret test-tls-ca-certs -n fybrik-system -o json | jq -r '.data."ca.crt"' | base64 -d > ca.crt

# copy the ca.crt to system certificate store in each pod
export OPENMETADATA_POD=$(kubectl get pods -n fybrik-system |grep openmetadata-connector | awk '{print $1}')
# tar tool is needed for the copy operation.
kubectl exec -it ${OPENMETADATA_POD} -n fybrik-system -- microdnf install tar
kubectl cp ca.crt  ${OPENMETADATA_POD}:/etc/ssl/certs/ -n fybrik-system
export OPA_CONNECTOR_POD=$(kubectl get pods -n fybrik-system |grep opa-connector | awk '{print $1}')
kubectl cp ca.crt ${OPA_CONNECTOR_POD}:/etc/ssl/certs/ -n fybrik-system
export MANAGER_POD=$(kubectl get pods -n fybrik-system |grep manager | awk '{print $1}')
kubectl cp ca.crt ${MANAGER_POD}:/etc/ssl/certs/ -c manager -n fybrik-system
