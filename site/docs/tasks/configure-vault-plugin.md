
# Configure a new HashiCorp Vault Plugin

The following steps show how to configure a new [Vault secret plugin](../concepts/vault_plugins.md) for Fybrik.

## Before you begin

Ensure that you have the [Vault v1.9.x](https://www.vaultproject.io/downloads) to execute [Vault CLI](https://www.vaultproject.io/docs/commands) commands.

## Steps to use the plugin

1. [Login into Vault](https://www.vaultproject.io/docs/commands/login)

2. Register and enable the plugin during Vault server initialization in a specific path. 
<br/>An example of that can be found in helm chart [values.yaml](https://github.com/fybrik/fybrik/blob/master/charts/vault/env/dev/plugin-secrets-values.yaml) file in the project where [Vault-plugin-secrets-kubernetes-reader](https://github.com/fybrik/vault-plugin-secrets-kubernetes-reader) plugin is enabled in `kubernetes-secrets` path:
```bash
    SHA256=$(sha256sum /usr/local/libexec/vault/vault-plugin-secrets-kubernetes-reader | cut -d ' ' -f1) &&
    vault plugin register -sha256=$SHA256 secret vault-plugin-secrets-kubernetes-reader
    vault secrets enable -path=kubernetes-secrets vault-plugin-secrets-kubernetes-reader
```
3. Add [Vault policy ](https://www.vaultproject.io/docs/concepts/policies) to allow the [modules](../concepts/modules.md) to access secrets using the plugin.
<br/>Following is an example of a policy which gives permission to read secrets in Vault path `kubernetes-secrets`:
```bash
vault policy write "allow-all-dataset-creds" - <<EOF
      path "kubernetes-secrets/*" {
      capabilities = ["read"]
      }
      EOF
```
4. Have the data catalog [getAsset response](../../reference/connectors-datacatalog/Models/GetAssetResponse) contain the Vault secret path which should be used to retrieve the credentials for a given asset. When the Vault plugin is used to retrieve the credentials; the parameters to the plugin should follow the plugin usage instructions. This path will later be passed on to the [modules](./modules.md).
For example, when the credentials are stored in kubernetes secret as is done in the [Katalog](../reference/katalog.md) built-in data catalog; the [Vault-plugin-secrets-kubernetes-reader](https://github.com/fybrik/vault-plugin-secrets-kubernetes-reader) plugin can be used to retrieve the credentials. In this case two parameters should be passed: `paysim-csv`  which is the kubernetes secret name that holds the credentials and `fybrik-notebook-sample` is the secret namespace, both are known to the katalog when constructing the path. The `credentails` field in getAsset response should contain `"/v1/kubernetes-secrets/paysim-csv?namespace=fybrik-notebook-sample"` in this case.
5. Update the [modules](../concepts/modules.md) to use the [Vault related values](../../reference/crds#blueprintspecmoduleskeyargumentsassetsindexargsindexvaultkey) to retrieve dataset credentias during their runtime execution. The values contain `secretPath` field with the plugin path as described in the previous step.
The following snippet contains an example of such values. 

```bash
    vault:
      # Address is Vault address
      address: http://vault.fybrik-system:8200
      # AuthPath is the path to auth method used to login to Vault
      authPath: /v1/auth/kubernetes/login
      # Role is the Vault role used for retrieving the credentials
      role: module
      # SecretPath is the path of the secret holding the Credentials in Vault
      secretPath: /v1/kubernetes-secrets/paysim-csv?namespace=fybrik-notebook-sample
```

