## Configure Vault Kubernetes auth method in each cluster

The steps below were tested on Kind cluster version 8.28 and OpenShift Container Platform 4.5.

The Mesh for Data uses [HashiCorp Vault](https://www.vaultproject.io/) to provide running M4D modules in the clusters with the dataset credentials when accessing data.

This section describe the steps to enable the modules to authenticate to Vault in order for them to retrieve the dataset credentials.
Some of the steps described below are not specific to the Mesh for Data project but rather are Vault specific and can be found in Vault related online tutorials.

Module authentication is done by configuring Vault to use [Kubernetes auth method](https://www.vaultproject.io/docs/auth/kubernetes) in each cluster. Using this method the modules can authenticate to Vault by providing their service account token. Behind the scenes Vault authenticates the token by submitting TokenReview request to the API server of the kubernetes cluster where the module is running.

### Prerequisites unless M4D modules are running on the same cluster as the Vault instance:

1. The running Vault instance should have connectivity to the cluster API server for each cluster running M4D modules.
2. The running Vault instance should have an Ingress resource to enable M4D modules getting credentials.

### Enabling Kubernetes auth for each cluster with running M4D modules:

   1. Create a token reviewer service account called vault-auth in the m4d-system namespace and give it permissions to create tokenreviews.authentication.k8s.io at the cluster scope:

```bash
apiVersion: v1
kind: ServiceAccount
metadata:
  name: vault-auth
  namespace: m4d-system
---
apiVersion: v1
kind: Secret
metadata:
  name: vault-auth
  namespace: m4d-system
  annotations:
    kubernetes.io/service-account.name: vault-auth
type: kubernetes.io/service-account-token
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: role-tokenreview-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
  - kind: ServiceAccount
    name: vault-auth
    namespace: m4d-system
```

  2. Login to Vault.

  3. Enable the Kubernetes auth method in a new path:

```bash   
vault auth enable -path=<auth path> kubernetes
```

  4. Use the /config endpoint to configure Vault to talk to Kubernetes:
```bash
TOKEN_REVIEW_JWT=$(kubectl get secret vault-auth -n m4d-system -o jsonpath="{.data.token}" | base64 --decode)
vault write auth/<auth path>/config \
    token_reviewer_jwt="$TOKEN_REVIEW_JWT" \
    kubernetes_host=<Kubernetes api server address> \
    kubernetes_ca_cert=@ca.crt
```
More details on the parameters in the command above can be found [here](https://www.vaultproject.io/api/auth/kubernetes).

5. Add a role called `module` to allow the modules in in `m4d-blueprints` namespace to access secret that contains the dataset credentials:
```bash
vault write auth/<auth path>/role/module \
    bound_service_account_names="*" \
    bound_service_account_namespaces=m4d-blueprints \
    policies="read-dataset-creds" \
    ttl=24h
```
6. Add the cluster auth path to `cluster-metadata` ConfigMap in `m4d-system` namespace in each cluster as follows:
```bash
"VaultAuthPath":<auth path>
```
7. Restart of the manager in order to use the new configuration.

