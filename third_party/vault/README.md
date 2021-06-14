## Vault Helm deployment


This directory contains subfolders each with helm chart values.yaml file and RBAC related resources for the different Vault deployments.

The dataset credentials are retrieved from where they are stored (data catalog/Kubernetes secrets) using Vault API. Vault uses a custom secret engine implemented with [Vault plugins](https://www.vaultproject.io/docs/internals/plugins) to retrieve the credentials from where they are stored. An example for such plugin is [vault-plugin-secrets-kubernetes-reader](https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader) plugin which reads dataset credentials that are stored in Kubernetes secrets.

The helm chart values in the subfolders contain the setup of the plugins as the following describes:

- plugin-secrets-kubernetes-reader/ folder contains helm chart values to deploy vault with  [vault-plugin-secrets-kubernetes-reader](https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader) plugin.
- vault-single-cluster/ folder contains values to deploy Vault on a single cluster setup. This includes the following:
  - enabling [vault-plugin-secrets-kubernetes-reader](https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader) plugin.
  - enable kueberentes auth method for the cluster.
  - add policy and a role for the modules to access secrets in the plugin path.
