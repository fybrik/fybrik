# Quick Start Guide

Follow this guide to install Mesh for Data using default parameters that are suitable for experimentation on a single cluster.

<!-- For a full installation refer to the [full installation guide](./setup/install) instead. -->

## Before you begin

Ensure that you have the following:

- [Helm](https://helm.sh/) 3.3 or newer must be installed and configured on your machine.
- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.16 or newer must be installed on your machine.
- Access to a Kubernetes cluster such as [Kind](http://kind.sigs.k8s.io/) as a cluster administrator.

## Install cert-manager

Mesh for Data requires [cert-manager](https://cert-manager.io) to be installed to your cluster. 
Many clusters already include cert-manager. Check if `cert-manager` namespace exists in your cluster and only run the following if it doesn't exist:

```bash
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --version v1.2.0 \
    --create-namespace \
    --set installCRDs=true \
    --wait --timeout 120s
``` 

## Install Hashicorp Vault and plugins

[Hashicorp Vault](https://www.vaultproject.io/) and a [secrets-kubernetes-reader](https://github.com/mesh-for-data/vault-plugin-secrets-kubernetes-reader) plugin are used by Mesh for Data for credential management.

Run the following to install vault and the plugin in development mode:

=== "Kubernetes" 

    ```bash
    helm repo add hashicorp https://helm.releases.hashicorp.com
    helm repo update
    helm install vault hashicorp/vault --version 0.9.1 --create-namespace -n m4d-system \
        --set "server.dev.enabled=true" \
        --set "injector.enabled=false" \
        --values https://raw.githubusercontent.com/mesh-for-data/mesh-for-data/v0.1.0/third_party/vault/vault-single-cluster/values.yaml \
        --wait --timeout 120s
    kubectl wait --for=condition=ready --all pod -n m4d-system --timeout=120s
    kubectl apply -f https://raw.githubusercontent.com/mesh-for-data/mesh-for-data/v0.1.0/third_party/vault/vault-single-cluster/vault-rbac.yaml -n m4d-system
    ```

=== "OpenShift"

    ```bash
    helm repo add hashicorp https://helm.releases.hashicorp.com
    helm repo update
    helm install vault hashicorp/vault --version 0.9.1 --create-namespace -n m4d-system \
        --set "global.openshift=true" \
        --set "injector.enabled=false" \
        --set "server.dev.enabled=true" \
        --values https://raw.githubusercontent.com/mesh-for-data/mesh-for-data/v0.1.0/third_party/vault/vault-single-cluster/values.yaml \
        --wait --timeout 120s
    kubectl wait --for=condition=ready --all pod -n m4d-system --timeout=120s
    kubectl apply -f https://raw.githubusercontent.com/mesh-for-data/mesh-for-data/v0.1.0/third_party/vault/vault-single-cluster/vault-rbac.yaml -n m4d-system
    ```

## Install control plane

??? tip "Install latest development version from GitHub"

    The published Helm charts are only available for released versions. 
    To install the `dev` version install the charts from the source code.
    For example:
    ```bash
    git clone https://github.com/mesh-for-data/mesh-for-data.git
    cd mesh-for-data
    helm install m4d-crd charts/m4d-crd -n m4d-system --wait
    helm install m4d charts/m4d --set global.tag=latest -n m4d-system --wait
    ```

The control plane includes a `manager` service that connects to a data catalog and to a policy manager. 
Install the latest release of Mesh for Data with a built-in data catalog and with [Open Policy Agent](https://www.openpolicyagent.org) as the policy manager:

```bash
helm repo add m4d-charts https://mesh-for-data.github.io/charts
helm repo update
helm install m4d-crd m4d-charts/m4d-crd -n m4d-system --wait
helm install m4d m4d-charts/m4d -n m4d-system --wait
```


## Install modules

[Modules](../concepts/modules.md) are plugins that the control plane deploys whenever required. The [arrow flight module](https://github.com/mesh-for-data/arrow-flight-module) enables reading data through Apache Arrow Flight API. 

Install the latest[^1] release of arrow-flight-module:

```bash
kubectl apply -f https://github.com/mesh-for-data/arrow-flight-module/releases/latest/download/module.yaml -n m4d-system
```

[^1]: Refer to the [documentation](https://github.com/mesh-for-data/arrow-flight-module/blob/master/README.md#register-as-a-mesh-for-data-module) of arrow-flight-module for other versions