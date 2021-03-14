## Configure Vault Kubernetes auth method in each cluster

The Mesh for Data uses [HashiCorp Vault](https://www.vaultproject.io/) to provide M4D Modules running in the remote clusters with the dataset credentials to execute their tasks.

This section contains the required steps to enable the modules to authenticate to Vault in order for them to retrieve the dataset credentials.

Module authentication is done by configuring Vault to use [Kubernetes auth method](https://www.vaultproject.io/docs/auth/kubernetes) in each cluster. Using this method the modules can authenticate to Vault by providing their service account token. Behind the scenes Vault authenticates the token by submitting TokenReview request to the API server of the kubernetes cluster where the module is running. 

**Prerequisite for multi-cluster setup:**

1. Vault should have connectivity to the remote clusters to allow it to communicate to the clusters API server.
2. An [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) resource should be created in the coordinator cluster to enable communication to Vault service from the remote clusters.


Some of the steps described below are not specific to the Mesh for Data project but rather are Vault specific and can be found in Vault related online tutorials. The steps were tested on Kind cluster version 8.28 and OpenShift Container Platform 4.5.

It is assumed that Vault is deployed in the coordinator cluster in `m4d-system` namespace prior to executing the steps.

1. Setup [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) to Vault service in the coordinator cluster for multi-cluster setup:
In multi-cluster setup a kubernetes Ingress should be configured in the coordinator cluster to enable communication to Vault from remote clusters.
The ingress should be deployed in the `m4d-system` namespace where Vault is deployed.

2. Add Vault address to `m4d-config`:
The address of Vault should be added to the kubernetes ConfigMap called `m4d-config` in `m4d-system` namespace in the coordinator cluster as follows:
```bash
VAULT_ADDRESS: <Vault address>
```
In a multi-cluster setup the Vault address should be the Ingress address described above while in single cluster it should set to: `http://vault.m4d-system:8200/`.

**Pre-steps for configuring Vault:**
The next steps describe how to configure Vault to use Kubernetes auth method for each cluster. To do that, Vault root token should be used in order to execute the configuration commands against Vault. The root token can be extracted by the following commands executed in the coordinator cluster:
```bash
# Port forward, so we could access vault
kubectl port-forward service/vault -n m4d-system 8200:8200&
export VAULT_TOKEN=$(kubectl get secrets vault-unseal-keys -n m4d-system -o jsonpath={.data.vault-root} | base64 --decode)
# Kill the port-forward if nessecarry
kill -9 %%
```

After that, we can login to Vault from every cluster in the mesh with the following commands:
```bash
export VAULT_ADDR=<Vault address>
vault login <VAULT_TOKEN>
```

In a multi-cluster setup the Vault address should be the Ingress address described above while in single cluster it should set to: `http://127.0.0.1:8200` given port forwarding for Vault was done prior to executing the commands.


3. Setup Kubernetes Vault auth backend for each cluster:

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

  2. Enable the Kubernetes auth method in a new path:

```bash   
vault auth enable -path=<auth path> kubernetes
```

  3. Use the /config endpoint to configure Vault to talk to Kubernetes:
```bash
TOKEN_REVIEW_JWT=$(kubectl get secret vault-auth -n m4d-system -o jsonpath="{.data.token}" | base64 --decode)
vault write auth/<auth path>/config \
    token_reviewer_jwt="$TOKEN_REVIEW_JWT" \
    kubernetes_host=<Kubernetes api server address> \
    kubernetes_ca_cert=@ca.crt
```
More details on the parameters in the command above can be found [here](https://www.vaultproject.io/api/auth/kubernetes).

4. Add a role called `module` to allow the modules in in `m4d-blueprints` namespace to access secret that contains the dataset credentials:
```bash
vault write auth/<auth path>/role/module \
    bound_service_account_names="*" \
    bound_service_account_namespaces=m4d-blueprints \
    policies="read-dataset-creds" \
    ttl=24h
```
5. Add the cluster auth path to `cluster-metadata` ConfigMap in `m4d-system` namespace in each cluster as follows:
```bash
"VaultAuthPath":<auth path>
```
6. Restart of the manager in order to use the new configuration.

