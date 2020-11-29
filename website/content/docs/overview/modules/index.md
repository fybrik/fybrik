---
title: "Modules"
date: 2020-04-30T22:08:28+03:00
draft: false
weight: 40
---

The project currently has two extension mechanisms, namely connectors and modules. 
This page describes what modules are and how you can create your own modules.

# What are modules?

As described in the [Architecture]({{< baseurl >}}/docs/overview/architecture/) page, the control plane generates a description of a data plane based on policies and application requirements. That data plane includes components that are deployed by the control plane to fulfill different data-centric requirements.

For example, a component that can mask data can be deployed to enforce a data masking policy, or a component that copies data may be deplyed to create a local data copy to meet performace requirements, etc. 

Modules are the way to package such data plane components and make them available to the control plane. 
Specifically, a module is packaged as a [Helm](https://helm.sh/) chart that the control plane can install to a data plane. 
To make a module available to the control plane you need to register it by applying a [`M4DModule`]({{< baseurl >}}/docs/reference/api/generated/app/#k8s-api-github-com-ibm-the-mesh-for-data-manager-apis-app-v1alpha1-m4dmodule) CRD.


# Available modules

The table below lists the currently available modules:

 Name | Description | M4DModule 
 ---  | ---         | ---      
 [arrow-flight-module](https://{{< github_base >}}/the-mesh-for-data-flight-module) | reading datasets while performing data transformations | https://raw.githubusercontent.com/IBM/the-mesh-for-data-flight-module/master/module.yaml

<!-- implicit-copy-module is not listed because it's still only available as part of the project tests -->

# Registering modules

Register a module by applying its `M4DModule` YAML in the `m4d-system` namespace.
For example, the following registers the `arrow-flight-module`:
```bash
kubectl apply -f https://raw.githubusercontent.com/IBM/the-mesh-for-data-flight-module/master/module.yaml -n m4d-system
```
# How to create my own module?

When creating a new module 
There are three parts to a module:

1. [M4DModule YAML](#m4dmodule): links to the Helm chart and provides details like what functions the module implements
1. [Module Helm chart](#module-helm-chart): the package that the control plane installs as part of a data plane
1. [Module workload](#module-workload): the workload that runs once the Helm chart is installed

## M4DModule YAML

TODO

The specification of the `M4DModule` Kubernetes CRD is available in the [API documentation]({{< baseurl >}}/docs/reference/api/generated/app/#k8s-api-github-com-ibm-the-mesh-for-data-manager-apis-app-v1alpha1-m4dmodule). This section provides more information to help writing a `M4DModule` YAML file.

```yaml
apiVersion: app.m4d.ibm.com/v1alpha1
kind: M4DModule
metadata:
  name: "<module name>"
  namespace: m4d-system  # control plane namespace
spec:
  chart: "<helm chart link>" # e.g.: ghcr.io/username/chartname:chartversion
  statusIndicators:
    ... # omitted for brevity
  dependencies:
    - type: module # currently must be "module"
      name: "<module name>" # the `metadata.name` field of another applied M4DModule
  flows:
    - read  # optional
    - write # optional
    - copy  # optional
  capabilities:
    ... # omitted for brevity
```


The `spec.chart` field is a link to a the Helm chart stored in an OCI image registry. This is similar to how a Kubernetes `Pod` references a container image. See [Module Helm chart](#module-helm-chart) for more details.

The `statusIndicators` field is used for tracking the status of the module in terms of success or failure. In many cases this can be omitted and the status will be detected automatically. 

The `dependencies` field lists any requirements that must be fulfilled before installing the module. A dependency has a `type` and a `name`. Currently dependencies of type `module` are supported, indicating that another module must also be installed for this module to work.

The `flows` field indicates the types of capabilities supported by the module. {{< name >}} currently supports three data flows: `read` for enabling an application to read data, `write` for enabling an application to write data, and `copy` for performing an implicit data copy on behalf of the application. A module is associated with one or more data flow based on its functionality.

The `capabilities` field lists the capabilities of the module.
```yaml
capabilities:
  credentials-managed-by: secret-provider 
  api:
    protocol: "<protocol>"
    dataformat: "<format>" # optional
  supportedInterfaces:
    - flow: "<read/write/copy>"
      source: #optional
        protocol: "<protocol>"
        dataformat: "<format>"
      sink: #optional
        protocol: "<protocol>"
        dataformat: "<format>"
  actions:
    - id: "<enforcement action identifier>" 
      level: "<dataset/column>"
```

* **`api`** specifies what client-facing API the module exposes to a caller that invokes read or write commands 
* **`supportedInterfaces`** lists the supported data services that the module can read data from and write data to
* **`actions`** lists the enforcement actions (e.g., transforms) supported by the module. 
<!-- The actions are taken from a defined [Enforcement Actions Taxonomy](about:blank) -->

A `protocol` field can take a value such as `kafka`, `s3`, `jdbc-db2`, `m4d-arrow-flight`, etc.

A `format` field can take a value such as `avro`, `parquet`, `json`, or `csv`.

Note that a module that targets copy flows will omit the `api` field and contain just `source` and `sink`, a module that only supports reading data assets will omit the `sink` field and only contain `api` and `source`, a module that does not perform any enforcement actions can omit the `actions` field, etc.

## Module workload

TODO

Containerized workload that implements one or more data-centric functions.


## Module Helm chart

TODO

A module is packed as a Helm chart to be installed dynamically by the control plane whenever needed. The Helm chart must be stored in an [OCI registry](https://helm.sh/docs/topics/registries/) capable of holding Helm charts, such as [ghcr.io](https://ghcr.io). The project provides tooling for pushing modules to such Helm repositories.

The control plane performs the equivalent of `helm install` and related commands. The `values` it passes to the install command are described in [ModuleArguments]({{< baseurl >}}//docs/reference/api/generated/app/#k8s-api-github-com-ibm-the-mesh-for-data-manager-apis-app-v1alpha1-modulearguments). Thus, the Chart templates should use these values appropriately.  
