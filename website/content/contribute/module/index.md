---
title: Contributing a Module
date: 2020-11-28T22:10:51+02:00
draft: false
weight: 1
---

This page describes what exactly must be provided when contributing a module, and how to make the control plane aware of the new module so that it can be used.



Steps for creating a module:
1. Implement the logic of the module you are contributing - either directly in the [Module Workload](#module-workload) or in an external component.  If the logic is in an external component, then the module workload should act as a client - i.e. receiving paramaters from the control plane and passing them to the external component.
2. [Create the Module Helm Chart](#module-helm-chart) that will be used by the control plane to deploy the module workload, update it, and delete it as necessary.
3. [Create the M4DModule YAML](#m4dmodule-yaml) which describes the capabilities of the module workload, in which flows it should be considered for inclusion, its supported interfaces, and the link to the module helm chart.
4. [Register the Module](#register-the-module) to make it available to the control plane.

These steps are described in the following sections in more detail, so that you can create your own modules for use by the Mesh for Data.  Note that a new module is maintained in its own git repository, separate from the core [Mesh for Data repository] (https://github.ibm.com/data-mesh-research/datamesh).


## Module Workload

The module workload is associated with a specific user workload and is deployed by the control plane.  It may implement the logic required itself, or it may be a client interface to an external component.  

In either case its input parameters must match the relevant type of `ModuleArguments`.  The details of the different types of arguments are defined in [blueprint_types.go](/Users/sima/Work/Dev/go/sima-m4d/the-mesh-for-data/manager/apis/app/v1alpha1/blueprint_types.go). 

```
type ModuleArguments struct {
	// CopyArgs are parameters specific to modules that do implicit copies of data from one data store to another.
	// +optional
	Copy *CopyModuleArgs `json:"copy,omitempty"`

	// ReadArgs are parameters that are specific to modules that enable user workloads to read data or are involved in preparing data to be read.
	// +optional
	Read []ReadModuleArgs `json:"read,omitempty"`

	// WriteArgs are parameters that are specific to modules that enable an application to write data
	// +optional
	Write []WriteModuleArgs `json:"write,omitempty"`
}
```

The status of the module workload is monitored by the control plane via standard kubernetes mechanisms.  
TODO: Expand on this or delete it?

If the module workload needs to return information to the user, that information should be written to the rel.Info.Note. 
TODO: Where should it be written to??

## Module Helm Chart
For any module chosen by the control plane to be part of the data path, the control plane needs to be able to install/remove/upgrade an instance of the module. Mesh for Data uses [Helm](https://helm.sh/docs/intro/using_helm/)
to provide this functionality.  Thus, a [helm chart](https://helm.sh/docs/chart_template_guide/getting_started/) should be provided for each `Module Workload`.

Example: [Arrow Flight Module helm chart](https://github.com/IBM/the-mesh-for-data-flight-module/tree/cd168bb6cdf666c2ec1df960395c0dc1c8feeaa9/helm/afm)


## M4DModule YAML
M4DModule is a kubernetes Custom Resource Definition (CRD) which describes to the control plane the functionality provided by the module.  The M4DModule CRD has no controller. The specification of the `M4DModule` Kubernetes CRD is available in the [API documentation]({{< baseurl >}}/docs/reference/api/generated/app/#k8s-api-github-com-ibm-the-mesh-for-data-manager-apis-app-v1alpha1-m4dmodule). 

### Metadata about the module
The first section in the YAML file contains general information about the module and where it will run.

```
apiVersion: app.m4d.ibm.com/v1alpha1 # always this value
kind: M4DModule # always this value
metadata:
  name: "<module name>" # the name of your new module
  namespace: m4d-system  # control plane namespace. Always m4d-system
```

### spec.chart
This is a link to a the Helm chart stored in an OCI image registry. This is similar to how a Kubernetes `Pod` references a container image. See [Module Helm chart](#module-helm-chart) for more details.

```
spec:
  chart: "<helm chart link>" # e.g.: ghcr.io/username/chartname:chartversion
```

### statusIndicators
Used for tracking the status of the module in terms of success or failure. In many cases this can be omitted and the status will be detected automatically.

TODO: In which cases is it automatically detected and when is it not?
```
statusIndicators:
    - kind: <module name>
      successCondition: status.status == SUCCEEDED
      failureCondition: status.status == FAILED
      errorMessage: status.error
```


### dependencies
A dependency has a `type` and a `name`. Currently dependencies of type `module` are supported, indicating that another module must also be installed for this module to work.
```
dependencies:
    - type: module #currently the only option is a dependency on another module deployed by the control plane
      name: <dependent module name>
```


### flows
This field indicates the types of capabilities supported by the module. Currently supported are three data flows: `read` for enabling an application to read data or prepare data for being read, `write` for enabling an application to write data, and `copy` for performing an implicit data copy on behalf of the application. A module is associated with one or more data flow based on its functionality.

```
flows: # Indicate the data flow(s) in which the control plane should consider using this module 
- read  # optional
- write # optional
- copy  # optional
```

### capabilities

`capabilites.credentials-managed-by` may be either `secret-provider` or `data-mesh-auto`.  Currently only `secret-provider` is supported, meaning that the module will receive a link to the data set credentials and must use the Secret Provider API to obtain the credentials. 
TODO: Add API link
TODO: /Users/sima/Work/Dev/go/sima-m4d/the-mesh-for-data/manager/apis/app/v1alpha1/m4dmodule_types.go has `automatic` andd not `data-mesh-auto`

`capabilites.supportedInterfaces` lists the supported data services from which the module can read data and to which it can write 
* `flow` field can be `read`, `write` or `copy`
* `protocol` field can take a value such as `kafka`, `s3`, `jdbc-db2`, `m4d-arrow-flight`, etc.
* `format` field can take a value such as `avro`, `parquet`, `json`, or `csv`.
Note that a module that targets copy flows will omit the `api` field and contain just `source` and `sink`, a module that only supports reading data assets will omit the `sink` field and only contain `api` and `source`

`capabilites.api` indicates the protocol and data format supported for reading or writing data from the user's workload.

An example for a module that copies data from a db2 database table to an s3 bucket in parquet format.
```
capabilities:
    credentials-managed-by: secret-provider
    supportedInterfaces:
    - flow: copy  
      source:
        protocol: jdbc-db2
        dataformat: table
      sink:
        protocol: s3
        dataformat: parquet
```

An example for a module that has an API for reading data, and supports reading both parquet and csv formats from s3.
```
capabilities:
    credentials-managed-by: secret-provider
    api:
      protocol: m4d-arrow-flight
      dataformat: arrow
    supportedInterfaces:
    - flow: read
      source:
        protocol: s3
        dataformat: parquet
    - flow: read
      source:
        protocol: s3
        dataformat: csv
```

`capabilites.actions`  are taken from a defined [Enforcement Actions Taxonomy](about:blank) 
a module that does not perform any transformation on the data may omit the `capabilities.actions` field.

The following is an example of how a module would declare that it knows how to redact, remove or encrypt data.  For each action there is a level indication, which can be data set level, column level, or row level.  In the example shown column level is indicated, and the actions arguments indicate the columns on which the transformation should be performed.

```
capabilities:
    actions:
    - id: "redact-ID"
      name: "redact"
      level: 2 # column
      args:
        column_name: column_value
    - id: "removed-ID"
      name: "removed"
      level: 2 # column
      args:
        column_name: column_value
    - id: "encrypt-ID"
      name: "encrypt"
      level: 2 # column
```

### Full Examples 

The following are examples of YAMLs from fully implemented modules:
* An example YAML for a module that [copies from db2 to s3](https://github.com/IBM/the-mesh-for-data/blob/master/manager/testdata/e2e/module-implicit-copy-db2wh-to-s3.yaml) and includes transformation actions 
* And an example [arrow flight read module](https://github.com/IBM/the-mesh-for-data-flight-module/blob/master/module.yaml) YAML, also with transformation support

## Register the Module
To make the control plane aware of the module so that it can be included in appropriate workload data flows, the administrator must apply the M4DModule YAML in the `m4d-system` namespace.  This makes the control plane aware of the existence of the module.  Note though that it *does not* check that the module's helm chart exists.

For example, the following registers the `arrow-flight-module`:
```bash
kubectl apply -f https://raw.githubusercontent.com/IBM/the-mesh-for-data-flight-module/master/module.yaml -n m4d-system
```








