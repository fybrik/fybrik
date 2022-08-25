# Running OpenMetadata on Kind Kubernetes

We are working on offering Fybrik users the option to use OpenMetadata as the Fybrik data catalog.

The files in this directory can be used to deploy OpenMetadata on Kuberenetes, specifically on Kind Kubernetes. They are based on: https://github.com/open-metadata/OpenMetadata/issues/6324

To deploy OpenMetadata in the `open-metadata` namespace, run:
```bash
make
```

If you would like to change the namespace in which OpenMetadata would be deployed, or the default credentials for MYSQL or Airflow, edit the first lines in `Makefile.env` before you run `make`.

Please note that deploying OpenMetadata could take a long time (over 20 minutes on my VM).

if you want to access the OpenMetadata GUI, direct your browser to `localhost:8585` after running:
```bash
export POD_NAME=$(kubectl get pods --namespace open-metadata -l "app.kubernetes.io/name=openmetadata,app.kubernetes.io/instance=openmetadata" -o jsonpath="{.items[0].metadata.name}")
export CONTAINER_PORT=$(kubectl get pod --namespace open-metadata $POD_NAME -o jsonpath="{.spec.containers[0].ports[0].containerPort}")
echo "Visit http://127.0.0.1:8585 to use your application"
kubectl --namespace open-metadata port-forward $POD_NAME 8585:$CONTAINER_PORT
```
