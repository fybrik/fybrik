---
title: Notebook sample
weight: 10
---

{{< warning >}}
Incomplete. Show usage of OPA, usage of Egeria, Starting the notebook, Writing code; Also should be a KubeFlow Notebook sample.
{{</ warning >}}

This guides describes sample usage of {{< name >}} by describing a sample Jupyter notebook application.
It describes how you, as a data scientist, can use a Jupyter notebook for building a fraud detection model using a 
data asset you find in the data catalog. 

# Deploy the application

{{< tip >}}
Kubernetes applications can be deployed using any standard deployment tool, from plane `kubectl apply` to [Razee](https://razee.io/).  {{< name >}} only requires a common label.
{{</ tip >}}

You have a Kubernetes namespace (OpenShift project) that you can work in, called `fraud-detection`.
```
kubectl create namespace fraud-detection
```

Deploy the Jupyter notebook workload by applying a yaml:
```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  name: notebook-deployment
  labels:
    app: my-notebook-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-notebook-app
  template:
    metadata:
      labels:
        app: my-notebook-app
      annotations:
        sidecar.istio.io/inject: "true"
    spec:
      containers:
        - name: notebook
          image: jupyter/datascience-notebook
          ports:
            - containerPort: 8888
              protocol: TCP
---
kind: Service
apiVersion: v1
metadata:
  name: notebook-service
  labels:
    app: my-notebook-app
spec:
  ports:
    - protocol: TCP
      name: http
      port: 80
      targetPort: 8888
  selector:
    app: my-notebook-app
  type: ClusterIP
```

# Register with {{< name >}}

Your notebook don't have data access credentials and would not be able to operate without registering the Application to the {{< name >}} control plane. 

Register your application by creating a `M4DApplication` resource that provides metadata about your application:

```yaml
apiVersion: app.m4d.ibm.com/v1alpha1
kind: M4DApplication
metadata:
  name: my-notebook-app
  labels:
    app: my-notebook-app
spec:
  selector:
    matchLabels:
      app: my-notebook-app
  appInfo:
    purpose: fraud-detection
    processingGeography: US
    role: Security
  data:
    - catalogID: 87ffdca3-8b5d-4f77-99f9-0cb1fba1f73f
      dataSetID: 01c6f0f0-9ffe-4ccc-ac07-409523755e72 
      ifDetails:
        protocol: S3
        dataformat: parquet
```

The definition above contains `appInfo` with declared attributes of your application. It also contains a `data` section with a list of datasets from the catalog. For each dataset you also specify the interface you want to use in order to consume the asset.

