---
title: ODPi Egeria Install
linktitle: ODPi Egeria Install
weight: 30
---
{{< tip >}}
This guide was tested against ODPi Egeria [version 2.1](https://github.com/odpi/egeria/releases/tag/V2.1)
{{</ tip >}}

By default, [ODPi Egeria](https://www.odpi.org/projects/egeria) is used as the data catalog. The steps below provide instructions on how to install ODPi Egeria in a Kubernetes environment.


## Installation

- Clone or download the project and then cd to the egeria folder
    ```bash
    cd third_party/egeria
    ```
- Create a dedicated Kubernetes namespace for Egeria:
    ```bash
    kubectl create namespace egeria-catalog
    ```
- To install egeria, execute:
    ```bash
    make deploy
    ```
- To uninstall egeria, execute:
    ```bash
    make undeploy
    ```
- To install egeria in Openshift, execute:
    ```bash
    WITHOUT_OPENSHIFT=false make deploy
    ```
- To uninstall egeria in Openshift, execute:
    ```bash
    WITHOUT_OPENSHIFT=false make undeploy
    ```
- Alternately, you can install egeria by following the instructions provided in the [ODPI Egeria website](https://egeria.odpi.org/open-metadata-resources/open-metadata-deployment/charts/odpi-egeria-lab/).
