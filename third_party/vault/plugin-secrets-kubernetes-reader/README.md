## Vault Helm Installation with vault-plugin-secrets-kubernetes-reader plugin

The dataset credentials are retrieved from where they are stored (catalog/kubernetes secrets) using [Vault plugins](https://www.vaultproject.io/docs/internals/plugins).

This directory contains values.yaml file with values to install [vault-plugin-secrets-kubernetes-reader](https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader) plugin which reads dataset credentials that are stored in kubernetes secrets.
