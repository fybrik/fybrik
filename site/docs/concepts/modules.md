# Modules

The project currently has two extension mechanisms, namely connectors and modules. 
This page describes what modules are and how they are leveraged by the control plane to build the data plane flow.  

## What are modules?

As described in the [architecture](./architecture.md) page, the control plane generates a description of a data plane based on policies and application requirements. This is known as a blueprint, and includes components that are deployed by the control plane to fulfill different data-centric requirements.  For example, a component that can mask data can be used to enforce a data masking policy, or a component that copies data may be used to create a local data copy to meet performance requirements, etc. 

Modules are the way to describe such data plane components and make them available to the control plane. A module is packaged as a [Helm](https://helm.sh/) chart that the control plane can install to a workload's data plane. To make a module available to the control plane it must be [registered](#registering-a-module) by applying a [`M4DModule`](../reference/crds.md#m4dmodule) CRD.

The functionality described by the module may be deployed (a) per workload, or (b) it may be composed of one or more components that run independent of the workload and its associated control plane.  In the case of (a), the control plane handles the deployment of the functional component. In the case of (b) where the functionality of the module runs independently and handles requests from multiple workloads, a client module is what is deployed by the control plane.  This client module passes parameters to the external component(s) and monitors the status and results of the requests to the external component(s). 
<!-- TODO: Add "which are declared as a dependencies in the module yaml"  when we support it-->

The following diagram shows an example with an Arrow Flight module that is fully deployed by the control plane and a second module where the client is deployed by the control plane but the ETL component providing the functionality has been independently deployed and supports multiple workloads.

![Example](../static/module_arch.png)

## Components that make up a module

There are several parts to a module:

1. **Optional** external component(s): deployed and managed independently of Mesh for Data.
1. [Module Workload](../contribute/modules.md#module-workload): the workload that runs once the Helm chart is installed by the control plane.
Can be a client to the external component(s) or be independent.
1. [Module Helm Chart](../contribute/modules.md#module-helm-chart): the package containing the module workload that the control plane installs as part of a data plane.
1. [M4DModule YAML](../contribute/modules.md#m4dmodule-yaml): describes the functional capabilities, supported interfaces, and has links to the Module Helm chart.

## Registering a module

To make the control plane aware of the module so that it can be included in appropriate workload data flows, the administrator must apply the M4DModule YAML in the `m4d-system` namespace.  This makes the control plane aware of the existence of the module.  Note that it **does not** check that the module's helm chart exists.

For example, the following registers the `arrow-flight-module`:
```bash
kubectl apply -f https://raw.githubusercontent.com/mesh-for-data/arrow-flight-module/master/module.yaml -n m4d-system
```

## When is a module used?

There are three main data flows in which modules may be used:
* Read - preparing data to be read and/or actually reading the data
* Write - writing a new data set or appending data to an existing data set
* Copy - for performing an implicit data copy on behalf of the application.  The decision to do an implicit copy is made by the control plane, typically for performance or governance reasons.

A module may be used in one or more of these flows, as is indicated in the module's yaml file.

## Control plane choice of modules

A user workload description `M4DApplicaton` includes a list of the data sets required, the technologies that will be used to read them, and information about the location and reason for the use of the data.  This information together with input from data and enterprise policies, determine which modules are chosen by the control plane. Currently the logic for choosing the modules for the data plane is as follows:
1. If the user is requesting to read data, find all the read flow related modules
1. If the data set protocol/format and the protocol/format requested by the user do not match, then make an implicit copy of the data, storing it such that it is readable via the protocol/format requested by the user.
1. If the governance action(s) required on the data set are not supported by the read module, and it is supported by the implicit copy module ... then make an implicit copy. Otherwise no need for implicit copy, and read will be done from the source directly.

<!-- TODO: Update to address multi-cluster logic -->

## Available modules

The table below lists the currently available modules: 

Name | Description | M4DModule 
---  | ---         | ---      
[arrow-flight-module](https://github.com/mesh-for-data/arrow-flight-module) | reading datasets while performing data transformations | https://raw.githubusercontent.com/mesh-for-data/arrow-flight-module/master/module.yaml

<!-- implicit-copy-module is not listed because it's still only available as part of the project tests -->

## Contributing

Read  [Module Development](../contribute/modules.md) for details on the components that make up a module and how to create a module.
