# HashiCorp Vault plugins

[HashiCorp Vault plugins](https://www.vaultproject.io/docs/internals/plugins) are standalone applications that Vault server executes to enable third-party secret engines and auth methods. 
After their enablement during Vault server initialization, the plugins can be used as a regular auth or secrets backends. 
This project uses secrets plugins to retrieve dataset credentials by the running [modules](./modules.md). The plugins retrieve the credentials from where they are stored, for example, data catalog or in kubernetes secret.
[Vault-plugin-secrets-kubernetes-reader](https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader) plugin is an example of Vault custom secret plugin which retrieves dataset credentials stored in a kubernetes secret.

Additional secret plugins can be developed to retrieve credentials additional location. This [tutorial](https://learn.hashicorp.com/tutorials/vault/plugin-backends?in=vault/app-integration) can serve as a good starting point to learn about Vault plugin development.

The following steps are for configuring a secret plug-in for Mesh for Data:

1. Enable the plugin during Vault server initialization in a specific path. An example of that can be found in helm chart [values.yaml](https://github.com/mesh-for-data/mesh-for-data/blob/master/third_party/vault/vault-single-cluster/values.yaml) file in the project where [Vault-plugin-secrets-kubernetes-reader](https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader) plugin is enabled in `kubernetes-secrets` path:


```bash
      vault secrets enable -path=kubernetes-secrets vault-plugin-secrets-kubernetes-reader
```

2. Add [Vault policy ](https://www.vaultproject.io/docs/concepts/policies) to allow the [modules](./modules.md) to access secrets using the plugin.
Following is an example of a policy which gives permission to read secrets in Vault path `kubernetes-secrets`:

```bash
vault policy write "allow-all-dataset-creds" - <<EOF
      path "kubernetes-secrets/*" {
      capabilities = ["read"]
      }
      EOF
```
3. Have the `CatalogDatasetInfo` structure from the [data catalog response](https://mesh-for-data.github.io/mesh-for-data/dev/reference/connectors/#data_catalog_responseproto) contain the Vault secret path which should be used to retrieve the credentials for a given asset. When Vault plugin is used to retrieve the credentials the parameters to the plugin should follow the plugin usage instructions. This path will later be passed on to the [modules](./modules.md).
For example, when the credentials are stored in kubernetes secret as is done in the [Katalog](../reference/katalog.md) built-in data catalog; the [Vault-plugin-secrets-kubernetes-reader](https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader) plugin can be used to retrieve the credentials. In this case two parameters should be passed: `paysim-csv`  which is the kubernetes secret name that holds the credentials and `m4d-notebook-sample` is the secret namespace, both are known to the katalog when constructing the path.

The following snippet shows `CatalogDatasetInfo` structure with Vault secret path in `CredentialsInfo` field.

```bash
	connectors.CatalogDatasetInfo{
		DatasetId: m4d-notebook-sample/paysim-csv,
		Details: &connectors.DatasetDetails{
			Name:       m4d-notebook-sample/paysim-csv,
			Geo:        theshire,
			DataStore:  m4d-notebook-sample/paysim-csv,
			CredentialsInfo: &connectors.CredentialsInfo{
				VaultSecretPath: "/v1/kubernetes-secrets/paysim-csv?namespace=m4d-notebook-sample"
			},
		},
    }
```
4. Update the [modules](./modules.md) to use the [Vault related values](https://mesh-for-data.github.io/mesh-for-data/dev/reference/crds/#blueprintspecflowstepsindexargumentscopydestinationvault) to retrieve dataset credentias during their runtime execution. The values contain `secretPath` field with the plugin path as described in the previous step.
The following snippet, taken from [hello-world-module](https://github.com/mesh-for-data/hello-world-module) [values.yaml](https://github.com/mesh-for-data/hello-world-module/blob/main/hello-world-module/values.yaml) file, contains an example of such values. 

```bash
    vault:
      # Address is Vault address
      address: http://vault.m4d-system:8200
      # AuthPath is the path to auth method used to login to Vault
      authPath: /v1/auth/kubernetes/login
      # Role is the Vault role used for retrieving the credentials
      role: module
      # SecretPath is the path of the secret holding the Credentials in Vault
      secretPath: /v1/kubernetes-secrets/paysim-csv?namespace=m4d-notebook-sample
```

