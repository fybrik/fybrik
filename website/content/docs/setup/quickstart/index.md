---
title: Quick Start
weight: 10
---

This guide lets you quickly evaluate {{< name >}} using the builtin sample connectors and data access module. Instaling {{< name >}} control plane includes: 
- Installing Hashicorp Vault and connector as the credentials manager 
- Installing ODPi Egeria lab and connectors as the data catalog
- Installing Open Policy Agent (OPA) and connector as the policy manager.
- Arrow-Flight data access module for reading Parquet and CSV datasets

For a custom setup follow the [Installation instructions](/docs/setup/install)

The following sample was verified with the following software versions:
- Kubernetes: kind v0.10.0, OpenShift 4.3
You will also need to have
- git

The following steps assume you have the following available:
- Kubernetes cluster
- Kubectl

1.  Obtain a local copy of {{< name >}} repository
    ```bash
    git clone https://github.com/IBM/the-mesh-for-data.git
    ```
1.  Change to the root directory of the repository
    ```bash
    cd the-mesh-for-data
    ```
1. Set Mesh for Data namespace
    ```bash
    kubectl config set-context --current --namespace=m4d-system
    ```
1. Install {{< name >}} control plane including the built-in sample connectors and software:
    - Credentials manager: Hashicorp Vault (aka Vault)
    - Data catalog: ODPi Egeria (aka Egeria)
    - Policy manager: Open Policy Agent (aka OPA)

    Install on OpenShift:
    ```bash
    WITHOUT_OPENSHIFT=false ./hack/install.sh
    ```
    **Or**
    
    Install on Kind:
    ```bash
    ./hack/install.sh
    ```
1. Install the Arrow-Flight module
    ```
    kubectl apply -f https://raw.githubusercontent.com/IBM/the-mesh-for-data-flight-module/master/module.yaml
    ```

You can now start using {{< name >}}. For samples please see:
- [Sample Kubeflow notebook with {{< name >}}](/docs/usage/notebook-sample)

