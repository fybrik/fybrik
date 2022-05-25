# Prepare Fybrik environment

Typically, this would be done by an IT administrator.

- Install Fybrik using the [Quick Start](../get-started/quickstart.md) guide.
  This sample assumes the use of the built-in catalog, Open Policy Agent (OPA) and flight module.

# Create a namespace for the sample

Create a new Kubernetes namespace and set it as the active namespace:

```bash
kubectl create namespace fybrik-notebook-sample
kubectl config set-context --current --namespace=fybrik-notebook-sample
```

This enables easy [cleanup](../samples/cleanup.md) once you're done experimenting with the sample.