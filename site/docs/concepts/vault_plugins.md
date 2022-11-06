# HashiCorp Vault plugins

[HashiCorp Vault plugins](https://www.vaultproject.io/docs/internals/plugins) are standalone applications that Vault server executes to enable third-party secret engines and auth methods. 
After their enablement during Vault server initialization, the plugins can be used as a regular auth or secrets backends. 
This project uses secrets plugins to retrieve dataset credentials by the running [modules](./modules.md). The plugins retrieve the credentials from where they are stored, for example, data catalog or in kubernetes secret.
[Vault-plugin-secrets-kubernetes-reader](https://github.com/fybrik/vault-plugin-secrets-kubernetes-reader) plugin is an example of Vault custom secret plugin which retrieves dataset credentials stored in a kubernetes secret.

Additional secret plugins can be developed to retrieve credentials additional location. This [tutorial](https://learn.hashicorp.com/tutorials/vault/plugin-backends?in=vault/app-integration) can serve as a good starting point to learn about Vault plugin development.

Details on adding a new Vault plugin for Fybrik can be found in this [task](../tasks/add-vault-plugin.md).
