# Quick Start Guide

This guide lets you quickly evaluate Mesh for Data. 

## Before you begin

Ensure that you have the following:

- `git`
- `make`
- `kubectl` version 1.16 and above
- `helm` version 3.3 and above

## About this guide

By the end of this guide you will have the following installed to your Kubernetes cluster:

- The control plane of Mesh for Data
- Hashicorp Vault and connector as the credentials manager 
- ODPi Egeria lab and connectors as the data catalog
- Open Policy Agent (OPA) and connector as the policy manager.
- Arrow-Flight data access module for reading Parquet and CSV datasets

## Prepare for installing Mesh for Data

1.  Obtain a local copy of Mesh for Data repository
    ```bash
    git clone https://github.com/ibm/the-mesh-for-data.git
    ```
1.  Change to the root directory of the repository
    ```bash
    cd the-mesh-for-data
    ```

## Install Mesh for Data

1. Set the current namespace to `m4d-system`

    ```bash
    kubectl config set-context --current --namespace=m4d-system
    ```

1. Run the quick install script to install the control plane elements.

    ```bash
    ./hack/install.sh
    ```

    to install on OpenShift you need to run ```WITHOUT_OPENSHIFT=false ./hack/install.sh``` instead.

1. Enable the use of the [arrow flight module](https://github.com/ibm/the-mesh-for-data-flight-module)

    ```
    kubectl apply -f https://raw.githubusercontent.com/IBM/the-mesh-for-data-flight-module/master/module.yaml
    ```

## Next steps

You can now start using Mesh for Data. For samples please see:

- [Sample Kubeflow notebook with Mesh for Data](../samples/notebook.md)
