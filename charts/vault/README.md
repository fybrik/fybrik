
## Vault Helm Installation

The dataset credentials are retreived from where they are stored (catalog/kubernetes secrets) using [Vault plugins](https://www.vaultproject.io/docs/internals/plugins).

An example for such plugin is [vault-plugin-secrets-kubernetes-reader](https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader) which reads dataset credentials that are stored in kubernetes secrets.

The Vault Helm Installation can support plugins for different configurations depending on the values provided to the chart.

The values/ directory contains yaml files for the different plugins available.

