# Running OpenMetadata on Kind Kubernetes

These instructions are based on: https://github.com/open-metadata/OpenMetadata/issues/6324

The instructions explain how to install OpenMetadata in the `default` namespace. Make sure that the current namespace is `default`:

```bash
kubectl config set-context --current --namespace=default
```

As the instructions indicate, one needs to run:

```bash
kubectl apply -f pv1.yaml
kubectl apply -f pv2.yaml
kubectl create secret generic airflow-mysql-secrets --from-literal=airflow-mysql-password=airflow_pass
helm install openmetadata-dependencies open-metadata/openmetadata-dependencies --values values-deps.yaml
kubectl create secret generic mysql-secrets --from-literal=openmetadata-mysql-password=openmetadata_password
kubectl create secret generic airflow-secrets --from-literal=openmetadata-airflow-password=admin
helm install openmetadata open-metadata/openmetadata
```

Note: **BEFORE running these commands**, change the tag of the airflow image to the latest tag. In other words, in `values-deps.yaml`, replace `0.11.1` with the latest version (`0.11.4` as of the writing of this document).

Also note that deploying OpenMetadata could take a long time (over 20 minutes on my VM).
