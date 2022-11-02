# Notebook tests with TLS

[`run-notebook-readflow-tls-tests`](https://github.com/fybrik/fybrik/blob/master/Makefile#L116) test runs the read flow where the following components
are configured to use TLS:

- OPA connector
- katalog connector
- OPA server
- Fybrik manager
- Vault

All servers above except Vault server use mutual TLS. Vault server uses TLS.

The flow of the test is as follows:

1) [cert-manager](https://cert-manager.io) is deployed in `fybrik-system` namespace.
2) Certificates for the control-plane components mentioned above are deployed in `fybrik-system` namespace.
The certificates are generated using [cert-manager](https://cert-manager.io) resource called [`Certificate`](https://cert-manager.io/docs/concepts/certificate/) and signed by a single self-signed CA as shown in: [`testdata/notebook/read-flow-tls/setup-certs.sh`](https://github.com/fybrik/fybrik/blob/master/manager/testdata/notebook/read-flow-tls/setup-certs.sh). The generated certificates are stored in secrets in `fybrik-system` namespace and used to configure the control-plane components to use TLS upon their deployments.
3) Vault is deployed with values from `charts/vault/env/ha/vault-single-cluster-values-tls.yaml` which configure it to use TLS.
4) Fybrik is deployed in `fybrik-system` namespace with values from `charts/fybrik/notebook-test-readflow.tls.values.yaml` file which configures the components to use mutual tls.
5) The CA certificate secret which is used to sign the certificates in `fybrik-system` namespace is copied to `fybrik-blueprints` namespace as is shown in [`manager/testdata/notebook/read-flow/setup.sh`](https://github.com/fybrik/fybrik/blob/master/manager/testdata/notebook/read-flow/setup.sh).
An alternative is to use a mechanism for syncing secrets across namespaces as shown in https://cert-manager.io/docs/tutorials/syncing-secrets-across-namespaces.
This step is needed as once the Fybrik module is deployed it will need Vault's CA certificate to communicate with it.
6) The deployed FybrikModule resource is patched with above secret. This will allow the arrow-flight-module to communicate with Vault with TLS once it is up and running.

[`run-notebook-readflow-tls-system-cacerts-tests`](https://github.com/fybrik/fybrik/blob/master/Makefile#L123) test is similar to the above test with the following exceptions:

- Vault server and OPA server do not use TLS.
- The CA certificates of the componenets in `fybrik-system` namespace are copied directly to `/etc/ssl/certs/` direcory in the manager/connector pods and thus not specified in the helm chart values upon deployment as shown in [`manager/testdata/notebook/read-flow-tls/copy-cacert-to-pods.sh`](https://github.com/fybrik/fybrik/blob/master/manager/testdata/notebook/read-flow-tls/copy-cacert-to-pods.sh).

## Testing Vault with mutual TLS

To support Vault with mutual TLS [`tls_require_and_verify_client_cert`](https://developer.hashicorp.com/vault/docs/configuration/listener/tcp) should be set to true in Vault's values file `charts/vault/env/ha/vault-single-cluster-values-tls.yaml`.
In addition, the tls certificate and private key for the arrow-flight-module needs to be generated as well.
The certificates should be stored in a secret in `fybrik-blueprints` namespace where the module is deployed and thus their generation should be done after the `fybrik-blueprints` namespace is created.
Given that for testing purposes the certificates are signed with private CA then this CA certificate should exist before generating the certificates and private key for the arrow-flight-module.
To simplify the process, the same CA certificate which was used to sign the componenets in `fybrik-system` can also be used to sign certificate of arrow-flight-module.
To do so, a cert-manager [`Certificate`](https://cert-manager.io/docs/concepts/certificate/) resource can be created in `fybrik-system` namespace for the arrow-flight-module similar to the certificates deployed in step 2 above where the `dnsNames` field in the certificate should contain the module service name. This will create a secret in `fybrik-system` namespace with the module certificates.
To create a secret with the generated certificates in `fybrik-blueprints` namespace a mechanism for syncing secrets across namespaces can be used as shown in https://cert-manager.io/docs/tutorials/syncing-secrets-across-namespaces.

Then, the FybrikModule resource can be patched with the secret that contains the generated certificates and private key.
