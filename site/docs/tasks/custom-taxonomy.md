# Using a Custom Taxonomy for Resource Validation

## Background

Fybrik acts as an orchestrator of independent components.  For example, the author of the data governance policy manager, which provides the governance decisions, and the components that enforce those decisions are not necessarily the same.  Thus, there is no common terminology between them.  Similarly, the data formats and protocols defined in the data catalog may be defined differently than the components used for reading/writing data.

In order to enable all these independent components to be used in a single architecture, Fybrik provides a taxonomy.  It provides a mechanism for all the components to interact using a common dialect.

The project defines a set of immutable structural JSON schemas, or "taxonomies" for resources deployed in Fybrik. 
However, since the taxonomy is meant to be configurable, a `taxonomy.json` file is referenced from these schemas for any definition that is customizable.

The `taxonomy.json` file is generated from a base taxonomy and zero or more taxonomy layers:

- The base taxonomy is maintained by the project and includes all of the structural definitions that are subject to customization (e.g.: tags, actions). 

- The taxonomy layers are maintained by users and external systems that add customizations over the base taxonomy (e.g., defining specific tags, actions).


This task describes how to deploy Fybrik with a custom `taxonomy.json` file that is generated with the Taxonomy Compile CLI tool. 

## Taxonomy Compile CLI tool 

<!--
{% set TaxonomyCliVersion = TaxonomyCliVersion %}
-->

A CLI tool for compiling a base taxonomy and zero or more taxonomy layers is provided in the [taxonomy-cli repository](https://github.com/fybrik/taxonomy-cli), along with a Docker image to directly run the tool.

The base taxonomy can be found in [`charts/fybrik/files/taxonomy/taxonomy.json`](https://github.com/fybrik/fybrik/blob/master/charts/fybrik/files/taxonomy/taxonomy.json) and example layers can be found in [`samples/taxonomy/example`](https://github.com/fybrik/fybrik/tree/master/samples/taxonomy/example).

The following command can be used to run the Taxonomy Compile CLI tool from the provided Docker image.

Usage:
```bash
  docker run --rm --volume ${PWD}:/local --workdir /local/ ghcr.io/fybrik/taxonomy-cli:{{TaxonomyCliVersion}} compile --out <outputFile> --base <baseFile> [<layerFile> ...] [--codegen]
```

Flags:

- -b, --base string : File with base taxonomy definitions (required)

- --codegen : Best effort to make output suitable for code generation tools

- -o, --out string : Path for output file (default `taxonomy.json`)

This will generate a `taxonomy.json` file with the layers specified. 

Alternatively, the tool can be run from the root of the taxonomy-cli repository. Usage:
```bash
  go run main.go compile --out <outputFile> --base <baseFile> [<layerFile> ...] [--codegen]
```

## Deploy Fybrik with Custom Taxonomy

To deploy Fybrik with the generated `taxonomy.json` file, follow the [`quickstart guide`](https://fybrik.io/latest/get-started/quickstart/) but use the command below instead of `helm install fybrik fybrik-charts/fybrik -n fybrik-system --wait`:

```bash
helm install fybrik fybrik-charts/fybrik -n fybrik-system --wait --set-file taxonomyOverride=taxonomy.json
```
The `--set-file` flag will pass in your custom `taxonomy.json` file to use for taxonomy validation in Fybrik.
If this flag is not provided, Fybrik will use the default `taxonomy.json` file with no layers compiled into it. 


For an already deployed fybrik instance, it is possible to upgrade fybrik with an updated custom taxonomy file (`taxonomy.json`) with the following command:

```bash
helm upgrade fybrik fybrik-charts/fybrik -n fybrik-system --wait --set-file taxonomyOverride=taxonomy.json
```


## Examples of changing taxonomy

### Example 1: Add new intent for FybrikApplication

In this example we show how to update the application taxonomy. We show that when a FybrikApplication yaml containing a `Marketing` intent is submitted, it's validation fails because initially the application's taxonomy does not include `Marketing`. We then describe how to add `Marketing` to the taxonomy, enabling the validation to pass when we re-submit the FybrikApplication yaml.

Follow the [`quickstart guide`](https://fybrik.io/latest/get-started/quickstart/) but stop before the command `helm install fybrik fybrik-charts/fybrik -n fybrik-system --wait`
 (or `helm install fybrik charts/fybrik --set global.tag=master --set global.imagePullPolicy=Always -n fybrik-system --wait` in development mode).

The initial taxonomy to be used in this example is a base taxonomy that can be found in [`charts/fybrik/files/taxonomy/taxonomy.json`](https://github.com/fybrik/fybrik/blob/master/charts/fybrik/files/taxonomy/taxonomy.json) with the following taxonomy layer:

```yaml
definitions:
  AppInfo:
    properties:
      intent:
        type: string
        enum:
          - Customer Support
          - Fraud Detection
          - Customer Behavior Analysis
    required:
      - intent
```
Copy the taxonomy layer to a `taxonomy-layer.yaml` file.

The working directory is the fybrik repository.
In order to compile and merge the two taxonomies, the Taxonomy Compile CLI tool is used in the following way:

```bash
  docker run --rm --volume ${PWD}:/local --workdir /local/ ghcr.io/fybrik/taxonomy-cli:{{TaxonomyCliVersion}} compile \
      --out custom-taxonomy.json --base charts/fybrik/files/taxonomy/taxonomy.json taxonomy-layer.yaml
```

This command creates a `custom-taxonomy.json` file, which is included in the helm installation of fybrik using the following command:

```bash
helm install fybrik charts/fybrik --set global.tag=master --set global.imagePullPolicy=Always -n fybrik-system --wait --set-file taxonomyOverride=custom-taxonomy.json
```

Trying to deploy a fybrikapplication.yaml that has an intent of `Marketing` should fail validation beacuse there is no `Marketing` intent in the taxonomy. The following command should fail with a description of a validation error :

```yaml
cat << EOF | kubectl apply -f -
apiVersion: app.fybrik.io/v1beta1
kind: FybrikApplication
metadata:
  name: taxonomy-test
spec:
  selector:
   workloadSelector:
     matchLabels: {
       app: notebook
     }
  appInfo:
    intent: Marketing
    role: Business Analyst
  data:
    - dataSetID: "default/fake.csv"
      requirements:
        interface:
          protocol: s3
          dataformat: csv
EOF
```
The expected error is `The FybrikApplication "taxonomy-test" is invalid: spec.appInfo.intent: Invalid value: "Marketing": spec.appInfo.intent must be one of the following: "Customer Behavior Analysis", "Customer Support", "Fraud Detection"`. Thus, no FybrikApplication CRD was created.

To fix this, a new intent with `Marketing` value should be added to the taxonomy. Add a new value of "Marketing" in `custom-taxonomy.json` file in `intent` property as follows:

```
"intent": {
  "type": "string",
  "enum": [
    "Customer Behavior Analysis",
    "Customer Support",
    "Fraud Detection",
    "Marketing"
  ]
}
```

Now we upgrade the fybrik helm chart using the following command:

```bash
helm upgrade fybrik charts/fybrik --set global.tag=master --set global.imagePullPolicy=Always -n fybrik-system --wait --set-file taxonomyOverride=custom-taxonomy.json
```

After updating fybrik to get fybrikapplications with `Marketing` intent, the deployment of a fybrikapplication.yaml that has an intent of `Marketing` will succeed:

```yaml
cat << EOF | kubectl apply -f -
apiVersion: app.fybrik.io/v1beta1
kind: FybrikApplication
metadata:
  name: taxonomy-test
spec:
  selector:
   workloadSelector:
     matchLabels: {
       app: notebook
     }
  appInfo:
    intent: Marketing
    role: Business Analyst
  data:
    - dataSetID: "default/fake.csv"
      requirements:
        interface:
          protocol: s3
          dataformat: csv
EOF
```

The result is a FybrikApplication Custom Resource Definition instance called taxonomy-test.



### Example 2: Add new action for FybrikModule

In this example we show how to update the module taxonomy. We show that when a FybrikModule yaml containing a `FilterAction` action is submitted, it's validation fails because initially the module's taxonomy does not include `FilterAction`. We then describe how to add a new action `FilterAction` to the taxonomy, enabling the validation to pass when we re-submit the FybrikModule yaml.

Follow the [`quickstart guide`](https://fybrik.io/latest/get-started/quickstart/) but stop before the command `helm install fybrik fybrik-charts/fybrik -n fybrik-system --wait`
 (or `helm install fybrik charts/fybrik --set global.tag=master --set global.imagePullPolicy=Always -n fybrik-system --wait` in development mode).

The initial taxonomy to be used in this example is a base taxonomy that can be found in [`charts/fybrik/files/taxonomy/taxonomy.json`](https://github.com/fybrik/fybrik/blob/master/charts/fybrik/files/taxonomy/taxonomy.json) with the following taxonomy layer:

```yaml
definitions:
  Action:
    oneOf:
      - $ref: "#/definitions/RedactAction"
      - $ref: "#/definitions/RemoveAction"
      - $ref: "#/definitions/Deny"
  RedactAction:
    type: object
    properties:
      columns:
        items:
          type: string
        type: array
    required:
      - columns
  RemoveAction:
    type: object
    properties:
      columns:
        items:
          type: string
        type: array
    required:
      - columns
  Deny:
    type: object
    additionalProperties: false

```
Copy the taxonomy layer to a `taxonomy-layer.yaml` file.

The working directory is the fybrik repository.
In order to compile and merge the two taxonomies, the Taxonomy Compile CLI tool is used in the following way:

```bash
  docker run --rm --volume ${PWD}:/local --workdir /local/ ghcr.io/fybrik/taxonomy-cli:{{TaxonomyCliVersion}} compile \
      --out custom-taxonomy.json --base charts/fybrik/files/taxonomy/taxonomy.json taxonomy-layer.yaml
```

This command creates a `custom-taxonomy.json` file, which is included in the helm installation of fybrik using the following command:

```bash
helm install fybrik charts/fybrik --set global.tag=master --set global.imagePullPolicy=Always -n fybrik-system --wait --set-file taxonomyOverride=custom-taxonomy.json
```

Trying to deploy a fybrikmodule.yaml that has a `FilterAction` should fail validation beacuse there is no `FilterAction` in the taxonomy. The following command should fail with a description of a validation error :

```yaml
cat << EOF | kubectl apply -f -
apiVersion: app.fybrik.io/v1beta1
kind: FybrikModule
metadata:
  name: taxonomy-module-test
spec:
  type: service
  chart:
    name: ghcr.io/fybrik/fake
    values:
      image.tag: master
  capabilities:
    - capability: read
      scope: workload
      supportedInterfaces:
        - source:
            protocol: s3
            dataformat: parquet
        - source:
            protocol: s3
            dataformat: csv
      actions:
        - name: RedactAction
        - name: FilterAction
EOF
```
The expected error is `The FybrikModule "taxonomy-module-test" is invalid: spec.capabilities.0.actions.0.name: Invalid value: "FilterAction": spec.capabilities.0.actions.0.name must be one of the following: "Deny", "RedactAction", "RemoveAction"`. Thus, no FybrikModule CRD was created.

To fix this, a new action `FilterAction` should be added to the taxonomy. Add a new file `taxonomy-layer2.yaml` with the new action `FilterAction` as follows:

```yaml
definitions:
  Action:
    oneOf:
      - $ref: "#/definitions/FilterAction"
  FilterAction:
    type: object
    properties:
      threshold:
        type: integer
      operation:
        type: string
    required:
      - threshold
```
Now we create the `custom-taxonomy.json` file as before, by using the following command:

```bash
  docker run --rm --volume ${PWD}:/local --workdir /local/ ghcr.io/fybrik/taxonomy-cli:{{TaxonomyCliVersion}} compile \
      --out custom-taxonomy.json --base charts/fybrik/files/taxonomy/taxonomy.json taxonomy-layer.yaml taxonomy-layer2.yaml
```

Now we upgrade the fybrik helm chart using the following command:

```bash
helm upgrade fybrik charts/fybrik --set global.tag=master --set global.imagePullPolicy=Always -n fybrik-system --wait --set-file taxonomyOverride=custom-taxonomy.json
```

After updating fybrik to get fybrikmodule with `FilterAction`, the deployment of a fybrikmodule.yaml that has a `FilterAction` will succeed:

```yaml
cat << EOF | kubectl apply -f -
apiVersion: app.fybrik.io/v1beta1
kind: FybrikModule
metadata:
  name: taxonomy-module-test
spec:
  type: service
  chart:
    name: ghcr.io/fybrik/fake
    values:
      image.tag: master
  capabilities:
    - capability: read
      scope: workload
      supportedInterfaces:
        - source:
            protocol: s3
            dataformat: parquet
        - source:
            protocol: s3
            dataformat: csv
      actions:
        - name: RedactAction
        - name: FilterAction
EOF
```

The result is a FybrikModule Custom Resource Definition instance called taxonomy-module-test.
