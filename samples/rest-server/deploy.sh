#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

echo "kind: Deployment
apiVersion: apps/v1
metadata:
  name: rest-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rest-server
  template:
    metadata:
      labels:
        app: rest-server
    spec:
      containers:
      - name: datauserserver
        image: "$DOCKER_HOSTNAME"/"$DOCKER_NAMESPACE"/datauserserver:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
      restartPolicy: Always" > Deployment.yaml

kubectl delete service datauserserver || true
kubectl delete deployment rest-server
kubectl apply -f Deployment.yaml
kubectl apply -f resources.yaml


