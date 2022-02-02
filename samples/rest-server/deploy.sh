#!/bin/bash
# Copyright 2020 IBM Corp.
# SPDX-License-Identifier: Apache-2.0

echo "kind: Deployment
apiVersion: apps/v1
metadata:
  name: gui
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gui
  template:
    metadata:
      labels:
        app: gui
    spec:
      containers:
      - name: datauserserver
        image: "$DOCKER_HOSTNAME"/"$DOCKER_NAMESPACE"/datauserserver:latest
        imagePullPolicy: Always
        envFrom:
        - configMapRef:
            name: fybrikgui-config
        ports:
        - containerPort: 8080
      restartPolicy: Always" > Deployment.yaml

kubectl apply -f gui_configmap.yaml
kubectl delete service datauserserver || true
kubectl delete deployment gui
kubectl apply -f Deployment.yaml
kubectl apply -f resources.yaml


