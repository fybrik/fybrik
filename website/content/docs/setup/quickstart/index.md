---
title: Quick Start
weight: 10
---

This guide lets you quickly evaluate {{< name >}}. For a custom setup follow the [Installation instructions](/docs/setup/install).

## Before you begin
Ensure that you have the following:
- `kubectl` with access to a Kubernetes cluster (this guide was tested with kind v0.10.0 and OpenShift 4.3)
- `git`
- `make`

## About this sample
By the end of this guide you will have the following installed to your Kubernetes cluster:
- The control plane of {{< name >}}
- Hashicorp Vault and connector as the credentials manager 
- ODPi Egeria lab and connectors as the data catalog
- Open Policy Agent (OPA) and connector as the policy manager.
- Arrow-Flight data access module for reading Parquet and CSV datasets

## Prepare for installing {{< name >}}

1.  Obtain a local copy of {{< name >}} repository
    ```bash
    git clone https://github.com/IBM/the-mesh-for-data.git
    ```
1.  Change to the root directory of the repository
    ```bash
    cd the-mesh-for-data
    ```
1. Set the current namespace to `m4d-system`
    ```bash
    kubectl config set-context --current --namespace=m4d-system
    ```

## Install {{< name >}}

1. Run the quick install script to install the control plane elements.

    ```bash
    ./hack/install.sh
    ```

    {{< warning >}}
    to install on OpenShift you need to run ```WITHOUT_OPENSHIFT=false ./hack/install.sh``` instead.
    {{< /warning >}}

1. Enable the use of the [arrow flight module](https://{{< github_base >}}/the-mesh-for-data-flight-module)
    ```
    kubectl apply -f https://raw.githubusercontent.com/IBM/the-mesh-for-data-flight-module/master/module.yaml
    ```

## Next steps
You can now start using {{< name >}}. For samples please see:
- [Sample Kubeflow notebook with {{< name >}}](/docs/usage/notebook-sample)
