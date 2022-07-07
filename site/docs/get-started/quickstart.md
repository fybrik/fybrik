{% if FybrikRelease is ne('__Release__') %}
    {% set currentRelease = FybrikRelease %}
    {% set fybrikVersionFlag = '--version ' + currentRelease|replace("v","") %}
    {% if arrowFlight[currentRelease]  is defined %}
         {% set arrowFlightRelease = arrowFlight[currentRelease] %}
    {% elif arrowFlight[currentRelease|truncate(4, True, '', 0)] is defined %}
        {% set arrowFlightRelease = arrowFlight[currentRelease|truncate(4, True, '', 0)] %}
    {% endif %}
{% endif %}

{% if arrowFlightRelease  is not defined %}
    {% set arrowFlightRelease = 'latest' %}
{% endif %}

# Quick Start Guide

Follow this guide to install Fybrik using default parameters that are suitable for experimentation on a single cluster.

<!-- For a full installation refer to the [full installation guide](./setup/install) instead. -->

## Before you begin

Ensure that you have the following:

- [Helm](https://helm.sh/) 3.3 or greater must be installed and configured on your machine.
- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.20 or newer must be installed on your machine.
- Access to a Kubernetes cluster such as [Kind](http://kind.sigs.k8s.io/) as a cluster administrator.


## Add required Helm repositories

```bash
helm repo add jetstack https://charts.jetstack.io
helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo add fybrik-charts https://fybrik.github.io/charts
helm repo update
```

## Install cert-manager

Fybrik requires [cert-manager](https://cert-manager.io) to be installed to your cluster[^1]. 
Many clusters already include cert-manager. Check if `cert-manager` namespace exists in your cluster and only run the following if it doesn't exist:

```bash
helm install cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --version v1.6.2 \
    --create-namespace \
    --set installCRDs=true \
    --wait --timeout 120s
``` 

## Install Hashicorp Vault and plugins

[Hashicorp Vault](https://www.vaultproject.io/) and a [secrets-kubernetes-reader](https://github.com/fybrik/vault-plugin-secrets-kubernetes-reader) plugin are used by Fybrik for credential management.

??? tip "Install latest development version from GitHub"

    The published Helm charts are only available for released versions.
    To install the `dev` version install the charts from the source code.
    For example:
	=== "Kubernetes"
		```bash
		git clone https://github.com/fybrik/fybrik.git
		cd fybrik
		helm dependency update charts/vault
		helm install vault charts/vault --create-namespace -n fybrik-system \
			--set "vault.injector.enabled=false" \
			--set "vault.server.dev.enabled=true" \
			--values charts/vault/env/dev/vault-single-cluster-values.yaml
		kubectl wait --for=condition=ready --all pod -n fybrik-system --timeout=120s
		```
	=== "OpenShift"

		```bash
		git clone https://github.com/fybrik/fybrik.git
		cd fybrik
		helm dependency update charts/vault
		helm install vault charts/vault --create-namespace -n fybrik-system \
			--set "vault.global.openshift=true" \
			--set "vault.injector.enabled=false" \
			--set "vault.server.dev.enabled=true" \
			--values charts/vault/env/dev/vault-single-cluster-values.yaml
		kubectl wait --for=condition=ready --all pod -n fybrik-system --timeout=120s
	    ```


Run the following to install vault and the plugin in development mode:

=== "Kubernetes" 

    ```bash
    helm install vault fybrik-charts/vault --create-namespace -n fybrik-system \
        --set "vault.injector.enabled=false" \
        --set "vault.server.dev.enabled=true" \
        --values https://raw.githubusercontent.com/fybrik/fybrik/{{ currentRelease|default('master') }}/charts/vault/env/dev/vault-single-cluster-values.yaml
    kubectl wait --for=condition=ready --all pod -n fybrik-system --timeout=120s
    ```

=== "OpenShift"

    ```bash
    helm install vault fybrik-charts/vault --create-namespace -n fybrik-system \
        --set "vault.global.openshift=true" \
        --set "vault.injector.enabled=false" \
        --set "vault.server.dev.enabled=true" \
        --values https://raw.githubusercontent.com/fybrik/fybrik/{{ currentRelease|default('master') }}/charts/vault/env/dev/vault-single-cluster-values.yaml
    kubectl wait --for=condition=ready --all pod -n fybrik-system --timeout=120s
    ```

## Install control plane

??? tip "Install latest development version from GitHub"

    The published Helm charts are only available for released versions. 
    To install the `dev` version install the charts from the source code.
    For example:
    ```bash
    git clone https://github.com/fybrik/fybrik.git
    cd fybrik
    helm install fybrik-crd charts/fybrik-crd -n fybrik-system --wait
    helm install fybrik charts/fybrik --set global.tag=master -n fybrik-system --wait
    ```

The control plane includes a `manager` service that connects to a data catalog and to a policy manager. 
Install the Fybrik release with a built-in data catalog and with [Open Policy Agent](https://www.openpolicyagent.org) as the policy manager:

```bash
helm install fybrik-crd fybrik-charts/fybrik-crd -n fybrik-system {{ fybrikVersionFlag }} --wait
helm install fybrik fybrik-charts/fybrik -n fybrik-system {{ fybrikVersionFlag }}  --wait
```

## Install modules

??? tip "Install latest development version from GitHub"

    To apply the latest development version of arrow-flight-module:
    ```bash
    kubectl apply -f https://raw.githubusercontent.com/fybrik/arrow-flight-module/master/module.yaml -n fybrik-system
    ```

[Modules](../concepts/modules.md) are plugins that the control plane deploys whenever required. The [arrow flight module](https://github.com/fybrik/arrow-flight-module) enables reading data through Apache Arrow Flight API. 

Install the {{ arrowFlightRelease }}[^2] release of arrow-flight-module:

```bash
{% if arrowFlightRelease != 'latest' %}
  kubectl apply -f https://github.com/fybrik/arrow-flight-module/releases/download/{{ arrowFlightRelease }}/module.yaml -n fybrik-system
{% else %}
  kubectl apply -f https://github.com/fybrik/arrow-flight-module/releases/{{ arrowFlightRelease }}/download/module.yaml -n fybrik-system
{% endif %}
```

[^1]:Fybrik version 0.6.0 and lower should use cert-manager 1.2.0
[^2]: Refer to the [documentation](https://github.com/fybrik/arrow-flight-module/blob/master/README.md#register-as-a-fybrik-module) of arrow-flight-module for other versions
