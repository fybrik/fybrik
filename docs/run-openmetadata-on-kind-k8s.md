# Running OpenMetadata on Kind Kubernetes

We are working on offering Fybrik users the option to OpenMetadata as the Fybrik data catalog.

The instructions here explain how to deploy OpenMetadata on Kuberenetes, specifically on Kind Kubernetes. They are based on: https://github.com/open-metadata/OpenMetadata/issues/6324

To deploy OpenMetadata in the `fybrik-system` namespace:
```bash
cd third_party/openmetadata
make deploy-openmetadata
```
Please note that deploying OpenMetadata could take a long time (over 20 minutes on my VM).
