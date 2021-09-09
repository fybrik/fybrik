#!/usr/bin/env bash
# Copyright 2021 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

kubectl create namespace fybrik-notebook-sample
kubectl config set-context --current --namespace=fybrik-notebook-sample

# Create jupyter notebook for debugging. TODO Remove when tests are working
kubectl create deployment my-notebook --image=jupyter/base-notebook --port=8888 -- start.sh jupyter lab --LabApp.token=''
kubectl set env deployment my-notebook JUPYTER_ENABLE_LAB=yes
kubectl label deployment my-notebook app.kubernetes.io/name=my-notebook
kubectl wait --for=condition=available --timeout=120s deployment/my-notebook
kubectl expose deployment my-notebook --port=80 --target-port=8888

# Create asset and secret
kubectl -n fybrik-notebook-sample apply -f example-asset.yaml
kubectl -n fybrik-notebook-sample apply -f s3credentials.yaml

# TODO Use remote version when compatible
# kubectl apply -f https://github.com/fybrik/arrow-flight-module/releases/latest/download/module.yaml -n fybrik-system
kubectl apply -n fybrik-system -f arrow-flight-module.yaml

kubectl -n fybrik-system create configmap sample-policy --from-file=sample-policy.rego
kubectl -n fybrik-system label configmap sample-policy openpolicyagent.org/policy=rego
while [[ $(kubectl get cm sample-policy -n fybrik-system -o 'jsonpath={.metadata.annotations.openpolicyagent\.org/policy-status}') != '{"status":"ok"}' ]]; do echo "waiting for policy to be applied" && sleep 5; done

# Forward port of test S3 instance
kubectl port-forward -n fybrik-system svc/s3 9090:9090 &