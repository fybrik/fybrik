## Vault Helm deployment

This directory contains helm chart for [Vault](https://www.vaultproject.io/) deployment.

### Vault helm chart values

The dataset credentials in the fybrik project are retrieved from where they are stored (data catalog/Kubernetes secrets) using Vault API. Vault uses a custom secret engine implemented with [Vault plugins](https://www.vaultproject.io/docs/internals/plugins) to retrieve the credentials from where they are stored. An example for such plugin is [vault-plugin-secrets-kubernetes-reader](https://github.com/fybrik/vault-plugin-secrets-kubernetes-reader) plugin which reads dataset credentials that are stored in Kubernetes secrets.

The helm chart values in the env/dev/ directory contain the setup of the plugins as the following describes:

- `plugin-secrets-kubernetes-reader-values.yaml` file contains helm chart values to deploy vault with  [vault-plugin-secrets-kubernetes-reader](https://github.com/fybrik/vault-plugin-secrets-kubernetes-reader) plugin.
- `vault-single-cluster-values.yaml` file contains values to deploy Vault on a single cluster setup. This includes the following:
  - enabling [vault-plugin-secrets-kubernetes-reader](https://github.com/fybrik/vault-plugin-secrets-kubernetes-reader) plugin.
  - enable kueberentes auth method for the cluster.
  - add policy and a role for the modules to access secrets in the plugin path.


In addition, `values.yaml` file contains values for setting RBAC related to plugins.