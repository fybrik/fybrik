---
title: "Modules"
date: 2020-04-30T22:08:28+03:00
draft: false
weight: 40
---

The project currently has two extension mechanisms, namely connectors and modules. 
Here we describe what modules are and what default modules are installed using the default {{< name >}} instsallation.

# What are modules?

At runtime, the control plane builds a blueprint for each application. The blueprint includes services, jobs, and configurations that are injected into the data path by the control plane to fulfill governance requirements and application requirements.

Modules are the way to pack such injectable services, jobs, and configurations and make them available to the control plane.

# Concept

For the control plane to build a blueprint that includes modules, the control plane needs a catalog of all the modules available for its use.  Therefore, the control plane is attached with a **module repository** listing all registered modules. 


For each module, the control plane needs **module details** describing the capabilities of the module (read, write, copy, transforms, etc.). This is needed so the control plane could pick the set of modules that can be used together in the same blueprint and fulfill all application and governance requirements. 


For any module picked by the control plane to be part of a blueprint, the control plane needs to be able to install/remove/upgrade an instance of the module. For example, if it's a service then the control plane needs to deploy and configure the service. This is done by installing the deployable **module package**.


Once deployed, we have an instance of the module's resources deployed to the cluster. Depending on the type of module, this can eventually be some code running in Kubernetes pods. This is the **module code**.


This document describes how {{< name >}} currently implements this concept using `M4DModule` CRD and Helm as well as listing requirements from the module code.

# M4DModule

A `M4DModule` CRD describes a single module:

```yaml
apiVersion: app.m4d.ibm.com/v1alpha1
kind: M4DModule
metadata:
  name: module name
  namespace: m4d-system  # control plane namespace
  labels:
    name: module name
    version: 1.0.0  # semantic version 
spec:
  chart: helm chart reference # e.g. docker.io/username/chartname:chartversion
  type: job/service/configuration # one of
  dependencies:
    - type: module # currently must be "module"
      name: module name # the `metadata.name` field of another applied M4DModule
  flows:
    - read  # optional
    - write # optional
    - copy  # optional
  capabilities:
    ... # omitted for brevity
```

An administrator registers a module with a control plane by applying a `M4DModule` resource in the namespace of the control plane (i.e., `m4d-system`). Hence, the module repository is the set of all `M4DModule` resources in that namespace.

The `spec.chart` field contains a reference to a Helm chart stored in an OCI image registry. This is similar to how a Kubernetes `Pod` contains a reference to a container image. The Helm chart is the module package that the control plane can dynamically install/remove/upgrade as needed as part of a blueprint. See [Module Helm chart](#module-helm-chart) for details.

The reminder of the `spec` field lists the module details: 
- **`type`**: specifies whether the module package deploys a job, a service, or just configuration. This is necessary for understanding what to expect from the status of the deployed module. Note that `service` is not necessarily a Kubernetes `Service` but rather sanything that runs without completion (e.g., a Kubernetes `Deployment` or `StatefulSet`). 
- **`dependencies`**: lists the requirements for using this module in a blueprint. Currently only a dependency on another modules is supported. When another module is listed as a dependency, it means that it must be part of the same blueprint. For modules of type `configuration` the first dependency indicates the parent module that the configuration applies to.
- **`flows`**: {{< name >}} currently supports three data flows: read for enabling an application to read data, write for enabling an application to write data, and copy for performing an implicit data copy on behalf of the application. A module is associated with one or more data flow based on its functionality.
- **`capabilities`**: lists the capabilities of the module as described next in [Module capabilities](#module-capabilities) 

# Module capabilities

The `spec.capabilities` field is defined as follows:

```yaml
capabilities:
  api:
    protocol: protocol
    dataformat: format # optional
  supportedInterfaces:
    - flow: read/write/copy
      source: #optional
        protocol: protocol
        dataformat: file format
      sink: #optional
        protocol: protocol
        dataformat: file format
  actions:
    - id: enforcement action identifier 
      level: dataset/column
```

* **`api`** specifies what client-facing API the module exposes to a caller that invokes read or write commands 
* **`supportedInterfaces`** lists the supported data services that the module can read data from and write data to
* **`actions`** lists the enforcement actions (e.g., transforms) supported by the module. The actions are taken from a defined [Enforcement Actions Taxonomy](about:blank)

A `protocol` field can take a value such as `kafka`, `s3`, `jdbc-db2`, `m4d-arrow-flight`, etc.

A `format` field can take a value such as `avro`, `parquet`, `json`, or `csv`.

Note that a module that targets copy flows will omit the `api` field and contain just `source` and `sink`, a module that only supports reading data assets will omit the `sink` field and only contain `api` and `source`, a module that does not perform any enforcement actions can omit the `actions` field, etc.

# Module Helm chart

Module is packed as a Helm chart to be installed dynamically by the control plane whenever needed. The Helm chart must be stored in an [OCI registry](https://helm.sh/docs/topics/registries/) capable of holding Helm charts. The project provides tooling for pushing modules to such Helm repositories.


Since the Helm package is installed by the control plane, it can't assume that arbitrary values will be passed. Instead, the project defines the exact values that are available to the control plane and that the module *should* accept as template values. The name, namespace, labels and service account must always be templated.


Context | Name | Type | Description
--------|------|------|------------
Root | name | String | Instance name
Root | namespace | String | Instance namespace
Root | labels | String[] | Instance labels
Root | serviceAccount | String | Service account
Root | arguments | Arguments | Arguments
Arguments | flow | String | copy, read or write
Arguments | copy | Copy | Copy arguments
Copy | source | DataStore | Copy source 
Copy | destination | DataStore | Copy sink
Copy | transformations | EnforcementAction[] | Actions on copy
Arguments | read | Read[] | Read arguments
Read | source | DataStore | Read datastore
Read | transformations | EnforcementAction[]      | Actions for read item
Arguments | write | Write[] | Write arguments
Write | destination | DataStore | Write datastore
Write | transformations | EnforcementAction[] | Actions on write item
