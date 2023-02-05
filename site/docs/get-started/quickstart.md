<!--
{% set arrowFlightRelease = arrow_flight_module_version(FybrikRelease,arrowFlight) %}
{% set currentRelease = fybrik_version(FybrikRelease) %}
{% set fybrikVersionFlag = fybrik_version_flag(FybrikRelease) %}
{% set CertMangerVersion = CertMangerVersion %}

{% set devCommentStart = fybrik_version_comment_start(FybrikRelease, "false") %}
{% set devCommentEnd = fybrik_version_comment_end(FybrikRelease, "false" ) %}
{% set prodCommentStart = fybrik_version_comment_start(FybrikRelease, "true") %}
{% set prodCommentEnd = fybrik_version_comment_end(FybrikRelease, "true" ) %}
-->
# Quick Start Guide

Follow this guide to install Fybrik using default parameters that are suitable for experimentation on a single cluster.

<!-- For a full installation refer to the [full installation guide](./setup/install) instead. -->

For a One Click Demo of Fybrik and a read data scenario, refer to [OneClickDemo](./OneClickDemo.md).

## Before you begin

Ensure that you have the following:

- [Helm](https://helm.sh/) 3.7.0 or greater must be installed and configured on your machine.
- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.23 or newer must be installed on your machine.
- Access to a Kubernetes cluster such as [Kind](http://kind.sigs.k8s.io/) as a cluster administrator. Kubernetes version 
support range is 1.23 - 1.25 although older versions may work well.


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
    --version {{CertMangerVersion}} \
    --create-namespace \
    --set installCRDs=true \
    --wait --timeout 120s
``` 

## Install Hashicorp Vault and plugins

[Hashicorp Vault](https://www.vaultproject.io/) and a [secrets-kubernetes-reader](https://github.com/fybrik/vault-plugin-secrets-kubernetes-reader) plugin are used by Fybrik for credential management.

??? tip "Install the latest development version from GitHub"

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

## Install data catalog

Fybrik assumes the existence of a data catalog that contains the metadata and connection information for data assets. Fybrik currently supports:

 1. [OpenMetadata](https://open-metadata.org/): An open-source end-to-end metadata management solution that includes data discovery, governance, data quality, observability, collaboration, and lineage.
 2. [Katalog](https://fybrik.io/dev/reference/katalog/): a data catalog stub used for testing and evaluation purposes.

If you plan to use [Katalog](https://fybrik.io/dev/reference/katalog/), you can skip to the [next section](#install-control-plane), but keep in mind that Katalog is mostly suitable for development and testing.

To use OpenMetadata, you can either use an existing deployment, or run the following commands to deploy OpenMetadata in kubernetes.

*Note:* OpenMetadata deployment requires a cluster storage provisioner that has PersistentVolume capability of ReadWriteMany Access Mode.
Below we provide examples of OpenMetadata installations on a single node kind cluster (for development and testing) and an 
OpenShift cluster on [IBM Cloud](https://www.ibm.com/cloud). For other deployments please check [OpenMetadata Kubernetes deployment](https://docs.open-metadata.org/deployment/kubernetes) 

=== "A single node Kind cluster"
    ```bash
    export FYBRIK_BRANCH={{ currentRelease|default('master') }}
    curl https://raw.githubusercontent.com/fybrik/fybrik/{{ currentRelease|default('master') }}/third_party/openmetadata/install_OM.sh | bash -
    ```

    The installation of OpenMetadata could take a long time (around 20 minutes on a VM running [kind](https://kind.sigs.k8s.io/) Kubernetes).

    Alternatively, if you want to change the OpenMetadata configuration parameters, run:
    ```bash
    export FYBRIK_BRANCH={{ currentRelease|default('master') }}
    curl https://raw.githubusercontent.com/fybrik/fybrik/{{ currentRelease|default('master') }}/third_party/openmetadata/install_OM.sh | bash -s -- --operation getFiles
    ```
    This command downloads the installation files to a temporary directory. Follow the instructions that appear on screen to change the configuration parameters and then run `make`. 
    Once the installation is over, be sure to remove the temporary directory.

=== "IBM OpenShift"
    ```bash
    export FYBRIK_BRANCH={{ currentRelease|default('master') }}
    curl https://raw.githubusercontent.com/fybrik/fybrik/{{ currentRelease|default('master') }}/third_party/openmetadata/install_OM.sh | bash -s -- --k8s-type ibm-openshift
    ```

    The installation of OpenMetadata could take a long time (around 20 minutes on a VM running [kind](https://kind.sigs.k8s.io/) Kubernetes).

    Alternatively, if you want to change the OpenMetadata configuration parameters, run:
    ```bash
    export FYBRIK_BRANCH={{ currentRelease|default('master') }}
    curl https://raw.githubusercontent.com/fybrik/fybrik/{{ currentRelease|default('master') }}/third_party/openmetadata/install_OM.sh | bash -s -- --k8s-type ibm-openshift --operation getFiles
    ```
    This command downloads the installation files to a temporary directory. Follow the instructions that appear on screen to change the configuration parameters and then run `make`. 
    Once the installation is over, be sure to remove the temporary directory.

=== "Existing deployment"
    If you want to use an existing OpenMetadata deployment, you have to configure it according to Fybrik requirements:
    Run the following commands to download the configuration files:
    ```bash
    export FYBRIK_BRANCH={{ currentRelease|default('master') }}
    curl https://raw.githubusercontent.com/fybrik/fybrik/{{ currentRelease|default('master') }}/third_party/openmetadata/install_OM.sh | bash -s -- --operation getFiles
    ```
    Follow the instructions that appear on screen to change the OpenMetadata location and credentials (`OPENMETADATA_ENDPOINT`, `OPENMETADATA_USER` and `OPENMETADATA_PASSWORD`) and then run `make prepare-openmetadata-for-fybrik`.

Running `make` installs OpenMetadata in the `open-metadata` namespace. To install OpenMetadata in another namespace, or to change the credentials of the different services used by OpenMetadata, edit the variables in the `Makefile.env` file.

## Install control plane

The control plane includes a `manager` service that connects to a data catalog and to a policy manager.

{{ prodCommentStart }}
=== "With OpenMetadata"
    Install the Fybrik release with [OpenMetadata](https://open-metadata.org/) as the data catalog and with
    [Open Policy Agent](https://www.openpolicyagent.org) as the policy manager.

    > **NOTE**: When installing fybrik with OpenMetadata as its data catalog, you need to specify the API endpoint for 
    OpenMetadata. The default value for that endpoint is `http://openmetadata.open-metadata:8585/api`.
    If you are using a different OpenMetadata deployment, replace the `openmetadataConnector.openmetadata_endpoint` 
    value in the helm installation command.

    The published Helm charts are only available for released versions. To install the `dev` version install the charts from the source code.
    ```bash
    git clone https://github.com/fybrik/fybrik.git
    cd fybrik
    helm install fybrik-crd charts/fybrik-crd -n fybrik-system --wait
    helm install fybrik charts/fybrik --set global.tag=master --set coordinator.catalog=openmetadata --set openmetadataConnector.openmetadata_endpoint=http://openmetadata.open-metadata:8585/api -n fybrik-system --wait
    ```
=== "With Katalog"
    Install the Fybrik release with [Katalog](https://fybrik.io/dev/reference/katalog/) as the data catalog and with
    [Open Policy Agent](https://www.openpolicyagent.org) as the policy manager.

    The published Helm charts are only available for released versions. To install the `dev` version install the charts from the source code.
    ```bash
    git clone https://github.com/fybrik/fybrik.git
    cd fybrik
    helm install fybrik-crd charts/fybrik-crd -n fybrik-system --wait
    helm install fybrik charts/fybrik --set global.tag=master --set coordinator.catalog=katalog -n fybrik-system --wait
    ```
{{ prodCommentEnd }}

{{ devCommentStart }}
=== "With OpenMetadata"
    Install the Fybrik release with [OpenMetadata](https://open-metadata.org/) as the data catalog and with
    [Open Policy Agent](https://www.openpolicyagent.org) as the policy manager.

    > **NOTE**: When installing fybrik with OpenMetadata as its data catalog, you need to specify the API endpoint for 
    OpenMetadata. The default value for that endpoint is `http://openmetadata.open-metadata:8585/api`.
    If you are using a different OpenMetadata deployment, replace the `openmetadataConnector.openmetadata_endpoint` 
    value in the helm installation command.
    
    ??? tip "Install the latest development version from GitHub"
            To apply the latest development version of fybrik:

            ```bash
            git clone https://github.com/fybrik/fybrik.git
            cd fybrik
            helm install fybrik-crd charts/fybrik-crd -n fybrik-system --wait
            helm install fybrik charts/fybrik --set global.tag=master --set coordinator.catalog=openmetadata --set openmetadataConnector.openmetadata_endpoint=http://openmetadata.open-metadata:8585/api -n fybrik-system --wait
            ```
    ```bash 
    helm install fybrik-crd fybrik-charts/fybrik-crd -n fybrik-system {{ fybrikVersionFlag }} --wait
    helm install fybrik fybrik-charts/fybrik --set coordinator.catalog=openmetadata --set openmetadataConnector.openmetadata_endpoint=http://openmetadata.open-metadata:8585/api -n fybrik-system {{ fybrikVersionFlag }} --wait
    ```
=== "With Katalog"
    Install the Fybrik release with [Katalog](https://fybrik.io/dev/reference/katalog/) as the data catalog and with
    [Open Policy Agent](https://www.openpolicyagent.org) as the policy manager.

    ??? tip "Install latest development version from GitHub" 
            To apply the latest development version of fybrik:

            ```bash
            git clone https://github.com/fybrik/fybrik.git
            cd fybrik
            helm install fybrik-crd charts/fybrik-crd -n fybrik-system --wait
            helm install fybrik charts/fybrik --set global.tag=master --set coordinator.catalog=katalog -n fybrik-system --wait
            ```
        
    ```bash
    helm install fybrik-crd fybrik-charts/fybrik-crd -n fybrik-system {{ fybrikVersionFlag }} --wait
    helm install fybrik fybrik-charts/fybrik --set coordinator.catalog=katalog -n fybrik-system {{ fybrikVersionFlag }} --wait
    ```
{{ devCommentEnd }}

## Install modules

??? tip "Install the latest development version from GitHub"

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
