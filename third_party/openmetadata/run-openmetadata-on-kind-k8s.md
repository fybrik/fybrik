# Running OpenMetadata on Kind Kubernetes

We are working on offering Fybrik users the option to OpenMetadata as the Fybrik data catalog.

The files in this directory can be used to deploy OpenMetadata on Kuberenetes, specifically on Kind Kubernetes. They are based on: https://github.com/open-metadata/OpenMetadata/issues/6324

To deploy OpenMetadata in the `fybrik-system` namespace, run:
```bash
make
```

If you would like to change the namespace in which OpenMetadata would be deployed, or the default credentials for MYSQL or Airflow, edit the first lines in `Makefile.env` before you run `make`.

Please note that deploying OpenMetadata could take a long time (over 20 minutes on my VM).
