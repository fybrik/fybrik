# HashiCorp Vault plugins

[HashiCorp Vault plugins](https://www.vaultproject.io/docs/internals/plugins) are standalone applications that Vault server executes to enable third-party secret engines and auth methods. 
After their enablement during Vault server initialization, the plugins can be used as a regular auth or secrets backends. 
This project uses secrets plugins to retrieve dataset credentials by the running [modules](./modules.md). The plugins retrieve the credentials from where they are stored, for example, data catalog or in kubernetes secret.
[Vault-plugin-secrets-kubernetes-reader](https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader) plugin is an example of Vault custom secret plugin which retrieves dataset credentials stored in a kubernetes secret.

The steps to use a new Vault plugin in the project are as follows:

1. Develop a plugin to retrieve the credentials from where they are stored. This [tutorial](https://learn.hashicorp.com/tutorials/vault/plugin-backends?in=vault/app-integration) can serve as a good starting point to learn about Vault plugin development.
2. Enable the plugin during Vault server initialization in a specific path. An example of that can be found in helm chart [values.yaml](https://github.com/IBM/the-mesh-for-data/blob/master/third_party/vault/vault-single-cluster/values.yaml) file in the project where [Vault-plugin-secrets-kubernetes-reader](https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader) plugin is enabled in `kubernetes-secrets` path. In addition, a policy to allow the [modules](./modules.md) to access secrets using the plugin should be added.
3. Have the [data catalog response](https://ibm.github.io/the-mesh-for-data/dev/reference/connectors/#connectors.CredentialsInfo) contain the Vault secret path which should be used to retrieve the credentials. This path will later be passed on to the [modules](./modules.md). The secret path contains the plugin path, for example, a secret path for the [Vault-plugin-secrets-kubernetes-reader](https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader) plugin could be `/v1/kubernetes-secrets/my-secret?namespace=default`, where `my-secret` is the kubernetes secret name which holds the credentials and `default` is the secret namespace.
4. Have the [modules](./modules.md) use the [Vault related values](https://ibm.github.io/the-mesh-for-data/dev/reference/crds/#blueprintspecflowstepsindexargumentscopydestinationvault) to retrieve dataset credentias during their runtime execution. The values contain the Vault secret path which contains the plugin path as described in the previous step,.

