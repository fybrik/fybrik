#!/bin/bash

kubectl delete -f policy-editor.yaml
kubectl delete -f policy-editor-rolebinding.yaml
kubectl delete -f opa.yaml
kubectl delete configmap meshfordata-external-data
kubectl delete configmap meshfordata-policy-lib
kubectl delete configmap user1-policies
kubectl delete configmap user2-policies
kubectl scale deployment opa-connector  --replicas=0
kubectl scale deployment opa-connector  --replicas=1

cd meshfordata-external-data
rm meshfordata-external-data.yaml
cd ../meshfordata-policy-lib
rm meshfordata-policy-lib.yaml
cd ../user1-policies
rm user1-policies.yaml
cd ../user2-policies
rm user2-policies.yaml

kubectl create -f policy-editor.yaml
kubectl create -f policy-editor-rolebinding.yaml
kubectl create -f opa.yaml

cd data-and-policies/meshfordata-external-data
kubectl kustomize ./ |kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/data=opa > meshfordata-external-data.yaml
kubectl create -f meshfordata-external-data.yaml

cd ../meshfordata-policy-lib
kubectl kustomize ./ |kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego > meshfordata-policy-lib.yaml
kubectl create -f meshfordata-policy-lib.yaml

cd ../user1-policies
kubectl kustomize ./ |kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego > user1-policies.yaml
kubectl create -f user1-policies.yaml

cd ../user2-policies
kubectl kustomize ./ |kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego > user2-policies.yaml
kubectl create -f user2-policies.yaml