# Tools used by the actors

- The data owner would typically register the dataset in a proprietary or open source catalog. We use [OpenMetadata](https://open-metadata.org/).
- The data owner needs to provide credentials for accessing a dataset. This is usually done via the data catalog, but credentials could be stored in kubernetes secrets as an alternative. 
- Proprietary and open source data governance systems are available either as part of a data catalog or as stand-alone systems.  This sample uses the open source [OpenPolicyAgent](https://www.openpolicyagent.org/).  The data governance officer writes the policies in OPA's [rego](https://www.openpolicyagent.org/docs/latest/policy-language/#what-is-rego) language.
- Any editor can be used to write the FybrikApplication.yaml via which the data user expresses the data usage requirements.
- A jupyter notebook is the workload from which the data is consumed by the data user.
- A Web Browser

# Prepare Fybrik environment

Typically, this would be done by an IT administrator.

- Install Fybrik using the [Quick Start](../get-started/quickstart.md) guide.
  This sample assumes the use of [OpenMetadata](https://open-metadata.org/), [OpenPolicyAgent](https://www.openpolicyagent.org/) and the [flight module](https://github.com/fybrik/arrow-flight-module/blob/master/README.md#register-as-a-fybrik-module).

# Create a namespace for the sample

Create a new Kubernetes namespace and set it as the active namespace:

```bash
kubectl create namespace fybrik-notebook-sample
kubectl config set-context --current --namespace=fybrik-notebook-sample
```

This enables easy [cleanup](../samples/cleanup.md) once you're done experimenting with the sample.
