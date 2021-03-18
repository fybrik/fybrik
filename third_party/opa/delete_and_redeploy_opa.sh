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
kubectl create configmap meshfordata-external-data --from-file=taxonomies.json --from-file=medical_taxonomies.json
kubectl label configmap meshfordata-external-data   openpolicyagent.org/data=opa

cd ../meshfordata-policy-lib
kubectl create configmap meshfordata-policy-lib --from-file=./
kubectl label configmap meshfordata-policy-lib  openpolicyagent.org/policy=rego

cd ../user1-policies
kubectl create configmap user1-policies --from-file=./
kubectl label configmap user1-policies openpolicyagent.org/policy=rego

cd ../user2-policies
kubectl create configmap user2-policies --from-file=./
kubectl label configmap user2-policies openpolicyagent.org/policy=rego