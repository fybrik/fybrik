# Notebook tests with TLS

[`run-notebook-readflow-tls-tests`](https://github.com/fybrik/fybrik/blob/master/Makefile#L116) test runs the read flow where the following components
are configured to use TLS:

- OPA connector
- Openmetadata connector
- Openmetadata server
- OPA server
- Fybrik manager
- Vault

All servers above except Openmetadata server and OPA server use mutual TLS.

The flow of the tests is as follows:

1) [cert-manager](https://cert-manager.io) is deployed in `fybrik-system` namespace.
2) [reflector](https://github.com/emberstack/kubernetes-reflector/blob/main/README.md) is deployed in `default` namespace. This is a mechanism for syncing secrets across namespaces.
3) TLS certificates for the control-plane components are deployed in `fybrik-system` namespace.
The certificates are generated using [cert-manager](https://cert-manager.io) resource called [`Certificate`](https://cert-manager.io/docs/concepts/certificate/) and signed by a single self-signed CA certificate, deplyed in `fybrik-system` namespace, as shown in [`testdata/notebook/read-flow-tls/setup-certs.sh`](https://github.com/fybrik/fybrik/blob/master/manager/testdata/notebook/read-flow-tls/setup-certs.sh) script.
The generated TLS certificates are automatically stored by cert-manager in secrets in `fybrik-system` namespace and used to configure the control-plane components to use mutual TLS upon their deployments in step 5.
In addition, to support Vault mutual TLS, the TLS certificate and private key for the arrow-flight module need to be generated in `fybrik-blueprints` namespace which is the namespace where arrow-flight module is deployed. These certificates are used when the arrow-flight module communicates with Vault for retrieving the dataset credentials.
To simplify the process, the same CA certificate which is used to sign the components in `fybrik-system` is used to sign the arrow-flight-module certificates.
To do so, a cert-manager [`Certificate`](https://cert-manager.io/docs/concepts/certificate/) resource is created in `fybrik-system` namespace for the arrow-flight-module as before, where the `dnsNames` field in the certificate contains the module service name. This will create a secret in `fybrik-system` namespace with the module certificates.
To create this secret in `fybrik-blueprints` namespace the [reflector](https://github.com/emberstack/kubernetes-reflector/blob/main/README.md) mechanism for syncing secrets across namespaces is used as shown in this [tutorial](https://cert-manager.io/docs/tutorials/syncing-secrets-across-namespaces).
4) Vault is deployed with values from `charts/vault/env/ha/vault-single-cluster-values-tls.yaml` which configure it to use mutual TLS.
5) Openmetadata server is deployed without TLS support.
6) Fybrik is deployed in `fybrik-system` namespace with values from `charts/fybrik/notebook-test-readflow.tls.values.yaml` file which configures the control-plane components to use mutual tls.
7) The deployed FybrikModule resource is patched with the details of the secret generated in step 3 which includes the arrow-flight-module TLS certificates. This will allow the arrow-flight-module to communicate with Vault using mutual TLS once it is up and running.

[`run-notebook-readflow-tls-system-cacerts-tests`](https://github.com/fybrik/fybrik/blob/master/Makefile#L123) test is similar to the above test with the following exceptions:

- Vault server, Openmetadata server and OPA server do not use TLS.
- The CA certificates of the components in `fybrik-system` namespace are copied directly to `/etc/ssl/certs/` directory in the manager/connector pods and thus not specified in the helm chart values upon deployment as shown in [`manager/testdata/notebook/read-flow-tls/copy-cacert-to-pods.sh`](https://github.com/fybrik/fybrik/blob/master/manager/testdata/notebook/read-flow-tls/copy-cacert-to-pods.sh).

