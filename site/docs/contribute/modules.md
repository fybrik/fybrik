# Module Development

This page describes what must be provided when contributing a [module](../concepts/modules.md).

## Steps for creating a module

1. Implement the logic of the module you are contributing. The implementation can either be directly in the [Module Workload](#module-workload) or in an external component.  If the logic is in an external component, then the module workload should act as a client - i.e. receiving paramaters from the control plane and passing them to the external component.
1. Create and publish the [Module Helm Chart](#module-helm-chart) that will be used by the control plane to deploy the module workload, update it, and delete it as necessary.
1. Create the [M4DModule YAML](#m4dmodule-yaml) which describes the capabilities of the module workload, in which flows it should be considered for inclusion, its supported interfaces, and the link to the module helm chart.
1. [Test](#test) the new module

These steps are described in the following sections in more detail, so that you can create your own modules for use by Mesh for Data.  Note that a new module is maintained in its own git repository, separate from the [mesh-for-data](https://github.com/mesh-for-data/mesh-for-data) repository.

## Module Workload

The module workload is associated with a specific user workload and is deployed by the control plane.  It may implement the logic required itself, or it may be a client interface to an external component.  

### Credential management

Modules that access or write data need credentials in order to access the data store. The credentials are retrieved from [HashiCorp Vault](https://www.vaultproject.io/). The parameters to [login](https://www.vaultproject.io/api-docs/auth/kubernetes#login) to vault and to [read secret](https://www.vaultproject.io/api/secret/kv/kv-v1#read-secret) are passed as part of the [arguments](../reference/crds.md#blueprintspecflowstepsindexarguments) to the module Helm chart.


An example for Vault Login API call which uses the Vault parameters is as follows:

```
$ curl -v --request POST <address>/<authPath> -H "Content-Type: application/json" --data '{"jwt": <module service account token>, "role": <role>}'
```

An example for Vault Read Secret API call which uses the Vault parameters is as follows:

```
$ curl --header "X-Vault-Token: ..." https://<address>/<secretPath>
```

## Module Helm Chart

For any module chosen by the control plane to be part of the data path, the control plane needs to be able to install/remove/upgrade an instance of the module. Mesh for Data uses [Helm](https://helm.sh/docs/intro/using_helm/) to provide this functionality. Follow the Helm [getting started](https://helm.sh/docs/chart_template_guide/getting_started/) guide if you are unfamiliar with Helm. Note that Helm 3.3 or above is required.

Because the chart is installed by the control plane, the input `values` to the chart must match the relevant type of [arguments](../reference/crds.md#blueprintspecflowstepsindexarguments). 
<!-- TODO: expand this when we support setting values in the M4DModule YAML: https://github.com/mesh-for-data/mesh-for-data/pull/42 -->

If the module workload needs to return information to the user, that information should be written to the `NOTES.txt` of the helm chart.

For a full example see the [Arrow Flight Module chart](https://github.com/mesh-for-data/arrow-flight-module/tree/cd168bb6cdf666c2ec1df960395c0dc1c8feeaa9/helm/afm).

### Publishing the Helm Chart

Once your Helm chart is ready, you need to push it to a [OCI-based registry](https://helm.sh/docs/topics/registries/) such as [ghcr.io](https://ghcr.io). This allows the control plane of Mesh for Data to later pull the chart whenever it needs to be installed.

You can use the [hack/make-rules/helm.mk](https://github.com/mesh-for-data/mesh-for-data/blob/master/hack/make-rules/helm.mk) Makefile, or manually push the chart:

```bash
HELM_EXPERIMENTAL_OCI=1 
helm registry login -u <username> <registry>
helm chart save <chart folder> <registry>/<path>:<version>
helm chart push <registry>/<path>:<version>
```

## M4DModule YAML

`M4DModule` is a kubernetes Custom Resource Definition (CRD) which describes to the control plane the functionality provided by the module.  The M4DModule CRD has no controller. The specification of the `M4DModule` Kubernetes CRD is available in the [API documentation](../reference/crds.md#m4dmodule). 

The YAML file begins with standard Kubernetes metadata followed by the `M4DModule` specification:
```yaml
apiVersion: app.m4d.ibm.com/v1alpha1 # always this value
kind: M4DModule # always this value
metadata:
  name: "<module name>" # the name of your new module
  namespace: m4d-system  # control plane namespace. Always m4d-system
spec:
   ...
```

The child fields of `spec` are described next. 

### `spec.chart`

This is a link to a the Helm chart stored in the [image registry]( https://helm.sh/docs/topics/registries/). This is similar to how a Kubernetes `Pod` references a container image. See [Module Helm chart](#module-helm-chart) for more details.

```
spec:
  chart: "<helm chart link>" # e.g.: ghcr.io/username/chartname:chartversion
```

### `spec.statusIndicators`

Used for tracking the status of the module in terms of success or failure. In many cases this can be omitted and the status will be detected automatically.

if the Helm chart includes standard Kubernetes resources such as Deployment and Service, then the status is automatically detected. If however Custom Resource Definitions are used, then the status may not be automatically detected and statusIndicators should be specified.

```yaml
statusIndicators:
    - kind: "<module name>"
      successCondition: "<condition>" # ex: status.status == SUCCEEDED
      failureCondition: "<condition>" # ex: status.status == FAILED
      errorMessage: "<field path>" # ex: status.error
```


### `spec.dependencies`

A dependency has a `type` and a `name`. Currently `dependencies` of type `module` are supported, indicating that another module must also be installed for this module to work.
```yaml
dependencies:
    - type: module #currently the only option is a dependency on another module deployed by the control plane
      name: <dependent module name>
```


### `spec.flows`

The `flows` field indicates the types of capabilities supported by the module. Currently supported are three data flows: `read` for enabling an application to read data or prepare data for being read, `write` for enabling an application to write data, and `copy` for performing an implicit data copy on behalf of the application. A module is associated with one or more data flow based on its functionality.

```yaml
flows: # Indicate the data flow(s) in which the control plane should consider using this module 
- read  # optional
- write # optional
- copy  # optional
```

### `spec.capabilities`

`capabilites.supportedInterfaces` lists the supported data services from which the module can read data and to which it can write 
* `flow` field can be `read`, `write` or `copy`
* `protocol` field can take a value such as `kafka`, `s3`, `jdbc-db2`, `m4d-arrow-flight`, etc.
* `format` field can take a value such as `avro`, `parquet`, `json`, or `csv`.
Note that a module that targets copy flows will omit the `api` field and contain just `source` and `sink`, a module that only supports reading data assets will omit the `sink` field and only contain `api` and `source`

`capabilites.api` describes the api exposed by the module for reading or writing data from the user's workload:
* `protocol` field can take a value such as `kafka`, `s3`, `jdbc-db2`, `m4d-arrow-flight`, etc 
* `dataformat` field can take a value such as `parquet`, `csv`, `arrow`, etc
* `endpoint` field describes the endpoint exposed the module

`capabilites.api.endpoint` describes the endpoint from a networking perspective:
* `hostname` field is the hostname to be used when accessing the module. Equals the release name. Can be omitted.
* `port` field is the port of the service exposed by the module.
* `scheme` field can take a value such as `http`, `https`, `grpc`, `grpc+tls`, `jdbc:oracle:thin:@`, etc

An example for a module that copies data from a db2 database table to an s3 bucket in parquet format.

```yaml
capabilities:
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

```yaml
capabilities:
    api:
      protocol: m4d-arrow-flight
      dataformat: arrow
      endpoint:
        port: 80
        scheme: grpc
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

```yaml
capabilities:
    actions:
    - id: "redact-ID"
      level: 2 # column
      args:
        column_name: column_value
    - id: "removed-ID"
      level: 2 # column
      args:
        column_name: column_value
    - id: "encrypt-ID"
      level: 2 # column
```

### Full Examples 

The following are examples of YAMLs from fully implemented modules:

* An example YAML for a module that [copies from db2 to s3](https://github.com/mesh-for-data/mesh-for-data/blob/master/manager/testdata/e2e/module-implicit-copy-db2wh-to-s3.yaml) and includes transformation actions 
* And an example [arrow flight read module](https://github.com/mesh-for-data/arrow-flight-module/blob/master/module.yaml) YAML, also with transformation support

## Test

1. [Register the module](../concepts/modules.md#registering-a-module) to make the control plane aware of it.
1. Create an `M4DApplication` YAML for a user workload, ensuring that the data set and other parameters included in it, together with the governance policies defined in the policy manager, will result in your module being chosen based on the [control plane logic](../concepts/modules.md#control-plane-choice-of-modules).
1. Apply the `M4DApplication` YAML.
1. View the `M4DApplication status`.
1. Run the user workload and review the results to check if they are what is expected.
