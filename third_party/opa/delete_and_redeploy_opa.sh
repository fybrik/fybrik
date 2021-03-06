#!/bin/bash

kubectl delete -f policy-editor.yaml
kubectl delete -f policy-editor-rolebinding.yaml
kubectl delete -f opa.yaml
kubectl delete configmap opa-json
kubectl delete configmap opa-policy
kubectl scale deployment opa-connector  --replicas=0
kubectl scale deployment opa-connector  --replicas=1

kubectl create -f policy-editor.yaml
kubectl create -f policy-editor-rolebinding.yaml

kubectl create -f opa.yaml
cd jsonfiles

kubectl kustomize ./ |kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/data=opa > newjson.yaml
kubectl create -f newjson.yaml

cd ../data-and-policies

kubectl kustomize ./ |kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego > newpolicy.yaml
kubectl create -f newpolicy.yaml
