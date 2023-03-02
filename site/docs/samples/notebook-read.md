# Notebook sample for the read flow

This sample demonstrates the following:

- how Fybrik enables a Jupyter notebook workload to access a cataloged dataset.

- how [arrow-flight module](https://github.com/fybrik/arrow-flight-module) is used for reading and transforming data.

- how policies regarding the use of personal information are seamlessly applied when accessing a dataset containing financial data.

In this sample you play multiple roles:

- As a data owner you upload a dataset and register it in a data catalog
- As a data steward you setup data governance policies
- As a data user you specify your data usage requirements and use a notebook to consume the data

## Prepare a dataset to be accessed by the notebook

This sample uses the [Synthetic Financial Datasets For Fraud Detection](https://www.kaggle.com/ealaxi/paysim1) dataset[^1] as the data that the notebook needs to read. Download and extract the file to your machine. You should now see a file named `PS_20174392719_1491204439457_log.csv`. Alternatively, use a sample of 100 lines of the same dataset by downloading [`PS_20174392719_1491204439457_log.csv`](https://raw.githubusercontent.com/fybrik/fybrik/master/samples/notebook/PS_20174392719_1491204439457_log.csv) from GitHub.

[^1]: Created by NTNU and shared under the ***CC BY-SA 4.0*** license.

Upload the CSV file to an object storage of your choice such as AWS S3, IBM Cloud Object Storage or Ceph.
Make a note of the service endpoint, bucket name, and access credentials. You will need them later.

??? tip "Setup and upload to localstack"

    For experimentation you can install localstack to your cluster instead of using a cloud service.
    
    1. Define variables for access key and secret key
      ```bash
      export ACCESS_KEY="myaccesskey"
      export SECRET_KEY="mysecretkey"
      ```
    2. Install localstack to the currently active namespace and wait for it to be ready:
      ```bash
      helm repo add localstack-charts https://localstack.github.io/helm-charts
      helm install localstack localstack-charts/localstack \
           --set startServices="s3" \
           --set service.type=ClusterIP \
           --set livenessProbe.initialDelaySeconds=25
      kubectl wait --for=condition=ready --all pod -n fybrik-notebook-sample --timeout=120s
      ```
    3. Create a port-forward to communicate with localstack server:
      ```bash
      kubectl port-forward svc/localstack 4566:4566 &
      ```
    3. Use [AWS CLI](https://aws.amazon.com/cli/) to upload the dataset to a new created bucket in the localstack server:
      ```bash
      export ENDPOINT="http://127.0.0.1:4566"
      export BUCKET="demo"
      export OBJECT_KEY="PS_20174392719_1491204439457_log.csv"
      export FILEPATH="/path/to/PS_20174392719_1491204439457_log.csv"
      export REGION=theshire
      aws configure set aws_access_key_id ${ACCESS_KEY} && aws configure set aws_secret_access_key ${SECRET_KEY}
      aws configure set region ${REGION}
      aws --endpoint-url=${ENDPOINT} s3api create-bucket --bucket ${BUCKET} --region ${REGION} --create-bucket-configuration LocationConstraint=${REGION}
      aws --endpoint-url=${ENDPOINT} s3api put-object --bucket ${BUCKET} --key ${OBJECT_KEY} --body ${FILEPATH}
      ```
## Register the dataset in a data catalog

In this step you are performing the role of the data owner, registering his data in the data catalog and registering the credentials for accessing the data in the credential manager.

=== "With OpenMetadata"
    Datasets can be registered either directly, through the OpenMetadata UI, or indirectly, through the data-catalog connector:

    === "Register an asset through the OpenMetadata UI"
        To register an asset directly through the OpenMetadata UI, follow the instructions [here](../../tasks/omd-discover-s3-asset/). These instructions also explain how to determine the asset ID.

        Store the asset ID in a `CATALOGED_ASSET` variable. For instance:
        ```bash
        CATALOGED_ASSET="openmetadata-s3.default.demo.\"PS_20174392719_1491204439457_log.csv\""
        ```

    === "Registering Dataset via Connector"
        We now explain how to register a dataset using the OpenMetadata connector.

        Begin by registering the credentials required for accessing the dataset as a kubernetes secret. Replace the values for `access_key` and `secret_key` with the values from the object storage service that you used and run:

        ```yaml
        cat << EOF | kubectl apply -f -
        apiVersion: v1
        kind: Secret
        metadata:
          name: paysim-csv
        type: Opaque
        stringData:
          access_key: "${ACCESS_KEY}"
          secret_key: "${SECRET_KEY}"
        EOF
        ```

        Next, register the data asset itself in the data catalog.
        We use port-forwarding to send asset creation requests to the OpenMetadata connector.
        ```bash
        kubectl port-forward svc/openmetadata-connector -n fybrik-system 8081:8080 &
        ```
        ```bash
        cat << EOF | curl -X POST localhost:8081/createAsset -d @-
        {
          "destinationCatalogID": "openmetadata",
          "destinationAssetID": "paysim-csv",
          "credentials": "/v1/kubernetes-secrets/paysim-csv?namespace=fybrik-notebook-sample",
          "details": {
            "dataFormat": "csv",
            "connection": {
              "name": "s3",
              "s3": {
                "endpoint": "http://localstack.fybrik-notebook-sample.svc.cluster.local:4566",
                "bucket": "demo",
                "object_key": "PS_20174392719_1491204439457_log.csv"
              }
            }
          },
          "resourceMetadata": {
            "name": "Synthetic Financial Datasets For Fraud Detection",
            "geography": "theshire ",
            "tags": {
              "Purpose.finance": "true"
            },
            "columns": [
              {
                "name": "nameOrig",
                "tags": {
                  "PII.Sensitive": "true"
                }
              },
              {
                "name": "oldbalanceOrg",
                "tags": {
                  "PII.Sensitive": "true"
                }
              },
              {
                "name": "newbalanceOrig",
                "tags": {
                  "PII.Sensitive": "true"
                }
              }
            ]
          }
        }
        EOF
        ```

        The response from the OpenMetadata connector should look like this:
        ```bash
        {"assetID":"openmetadata-s3.default.demo.\"PS_20174392719_1491204439457_log.csv\""}
        ```
        Store the asset ID in a `CATALOGED_ASSET` variable:
        ```bash
        CATALOGED_ASSET="openmetadata-s3.default.demo.\"PS_20174392719_1491204439457_log.csv\""
        ```

=== "With Katalog"
    We now explain how to register a dataset in the Katalog data catalog.

    Begin by registering the credentials required for accessing the dataset as a kubernetes secret. Replace the values for `access_key` and `secret_key` with the values from the object storage service that you used and run:

    ```yaml
    cat << EOF | kubectl apply -f -
    apiVersion: v1
    kind: Secret
    metadata:
      name: paysim-csv
    type: Opaque
    stringData:
      access_key: "${ACCESS_KEY}"
      secret_key: "${SECRET_KEY}"
    EOF
    ```

    Next, register the data asset itself in the data catalog.
    We use port-forwarding to send asset creation requests to the Katalog connector.
    ```bash
    cat << EOF | kubectl apply -f -
    apiVersion: katalog.fybrik.io/v1alpha1
    kind: Asset
    metadata:
      name: paysim-csv
    spec:
      secretRef:
        name: paysim-csv
      details:
        dataFormat: csv
        connection:
          name: s3
          s3:
            endpoint: "http://localstack.fybrik-notebook-sample.svc.cluster.local:4566"
            bucket: "demo"
            object_key: "PS_20174392719_1491204439457_log.csv"
      metadata:
        name: Synthetic Financial Datasets For Fraud Detection
        geography: theshire
        tags:
          Purpose.finance: true
        columns:
          - name: nameOrig
            tags:
              PII.Sensitive: true
          - name: oldbalanceOrg
            tags:
              PII.Sensitive: true
          - name: newbalanceOrig
            tags:
              PII.Sensitive: true
    EOF
    ```

    Store the asset name in a `CATALOGED_ASSET` variable:
    ```bash
    CATALOGED_ASSET="fybrik-notebook-sample/paysim-csv"
    ```

If you look at the asset creation request above, you will notice that in the `resourceMetadata` field, we request that the asset should be tagged with the `Purpose.finance` tag, and that three of its columns should be tagged with the `PII.Sensitive` tag. Those tags will be referenced below in the access policy rules. Tags are important because they are used to determine whether an application would be allowed to access a dataset, and if so, which transformations should be applied to it.

The asset is now registered in the catalog.

Notice the `resourceMetadata` field above. It specifies the dataset geography and tags. These attributes can later be used in policies.

For example, in the json above, the `geography` is set to `theshire`. You need make sure that it is same as the region of your fybrik control plane. You can get this information using the following command:

```shell
kubectl get configmap cluster-metadata -n fybrik-system -o 'jsonpath={.data.Region}'
```

[Quick Start](../get-started/quickstart.md) installs a fybrik control plane with the region `theshire` by default. If you change it or the `geography` in the json above, a [copy module](https://github.com/fybrik/mover) will be required by the policies, but we do not install any copy module in the [Quick Start](../get-started/quickstart.md).

## Define data access policies

Acting as the data steward, define an [OpenPolicyAgent](https://www.openpolicyagent.org/) policy to redact the columns tagged as `PII.Sensitive` for datasets tagged with `Purpose.finance`. Below is the policy (written in [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/#what-is-rego) language):

```rego
package dataapi.authz

rule[{"action": {"name":"RedactAction", "columns": column_names}, "policy": description}] {
  description := "Redact columns tagged as PII.Sensitive in datasets tagged with Purpose.finance = true"
  input.action.actionType == "read"
  input.resource.metadata.tags["Purpose.finance"]
  column_names := [input.resource.metadata.columns[i].name | input.resource.metadata.columns[i].tags["PII.Sensitive"]]
  count(column_names) > 0
}
```

In this sample only the policy above is applied. Copy the policy to a file named `sample-policy.rego` and then run:

```bash
kubectl -n fybrik-system create configmap sample-policy --from-file=sample-policy.rego
kubectl -n fybrik-system label configmap sample-policy openpolicyagent.org/policy=rego
while [[ $(kubectl get cm sample-policy -n fybrik-system -o 'jsonpath={.metadata.annotations.openpolicyagent\.org/policy-status}') != '{"status":"ok"}' ]]; do echo "waiting for policy to be applied" && sleep 5; done
```

You can similarly apply a directory holding multiple rego files.

## Deploy a Jupyter notebook

In this sample a Jupyter notebook is used as the user workload and its business logic requires reading the asset that we 
registered (e.g., for creating a fraud detection model). Deploy a notebook to your cluster:

1. Deploy JupyterLab:
```bash
kubectl create deployment my-notebook --image=jupyter/base-notebook --port=8888 -- start.sh jupyter lab --LabApp.token=''
kubectl set env deployment my-notebook JUPYTER_ENABLE_LAB=yes
kubectl label deployment my-notebook app.kubernetes.io/name=my-notebook
kubectl wait --for=condition=available --timeout=120s deployment/my-notebook
kubectl expose deployment my-notebook --port=80 --target-port=8888
```
2. Create a port-forward to communicate with JupyterLab:
```bash
kubectl port-forward svc/my-notebook 8080:80 &
```
3. Open your browser and go to [http://localhost:8080/](http://localhost:8080/).
4. Create a new notebook in the server


## Create a `FybrikApplication` resource for the notebook

Create a [`FybrikApplication`](../reference/crds.md#fybrikapplication) resource to register the notebook workload to the control plane of Fybrik. The value you place in the `dataSetID` field is your asset ID, as explained above. If you registered your dataset through the data catalog connector, enter the `assetID` which was returned to you by the connector, e.g. `"openmetadata-s3.default.demo.\"PS_20174392719_1491204439457_log.csv\""`.

<!-- TODO: role field removed but code still requires it -->
```yaml
cat <<EOF | kubectl apply -f -
apiVersion: app.fybrik.io/v1beta1
kind: FybrikApplication
metadata:
  name: my-notebook
  labels:
    app: my-notebook
spec:
  selector:
    workloadSelector:
      matchLabels:
        app: my-notebook
  appInfo:
    intent: Fraud Detection
  data:
    - dataSetID: ${CATALOGED_ASSET}
      requirements:
        interface: 
          protocol: fybrik-arrow-flight
EOF
```

Notice that:

* The `selector` field matches the labels of our Jupyter notebook workload.
* The `data` field includes a `dataSetID` that matches the asset identifier in the catalog.
* The `protocol` indicates that the developer wants to consume the data using Apache Arrow Flight. For some protocols a `dataformat` can be specified as well (e.g., `s3` protocol and `parquet` format).


Run the following command to wait until the `FybrikApplication` is ready and set the `CATALOGED_ASSET_MODIFIED` environment variable:

```bash
while [[ $(kubectl get fybrikapplication my-notebook -o 'jsonpath={.status.ready}') != "true" ]]; do echo "waiting for FybrikApplication" && sleep 5; done
CATALOGED_ASSET_MODIFIED=$(echo $CATALOGED_ASSET | sed 's/\./\\\./g')
while [[ $(kubectl get fybrikapplication my-notebook -o "jsonpath={.status.assetStates.${CATALOGED_ASSET_MODIFIED}.conditions[?(@.type == 'Ready')].status}") != "True" ]]; do echo "waiting for ${CATALOGED_ASSET} asset" && sleep 5; done
```

## Read the dataset from the notebook

In your **terminal**, run the following command to print the [endpoint](../../reference/crds/#fybrikapplicationstatusreadendpointsmapkey) to use for reading the data. It fetches the code from the `FybrikApplication` resource:
```bash
ENDPOINT_SCHEME=$(kubectl get fybrikapplication my-notebook -o "jsonpath={.status.assetStates.${CATALOGED_ASSET_MODIFIED}.endpoint.fybrik-arrow-flight.scheme}")
ENDPOINT_HOSTNAME=$(kubectl get fybrikapplication my-notebook -o "jsonpath={.status.assetStates.${CATALOGED_ASSET_MODIFIED}.endpoint.fybrik-arrow-flight.hostname}")
ENDPOINT_PORT=$(kubectl get fybrikapplication my-notebook -o "jsonpath={.status.assetStates.${CATALOGED_ASSET_MODIFIED}.endpoint.fybrik-arrow-flight.port}")
printf "\n${ENDPOINT_SCHEME}://${ENDPOINT_HOSTNAME}:${ENDPOINT_PORT}\n\n"
```
The next steps use the endpoint to read the data in a python notebook

1. Insert a new notebook cell to install pandas and pyarrow packages:
  ```python
  %pip install pandas pyarrow==7.0.*
  ```
2. Insert a new notebook cell to read the data. You need to replace both the `ENDPOINT` and the `CATALOGED_ASSET` values, which were obtained in previous steps:
  ```bash
  import json
  import pyarrow.flight as fl
  import pandas as pd

  # Create a Flight client
  client = fl.connect('<ENDPOINT>')

  # Prepare the request
  request = {
      "asset": '<CATALOGED_ASSET>',
      # To request specific columns add to the request a "columns" key with a list of column names
      # "columns": [...]
  }

  # Send request and fetch result as a pandas DataFrame
  info = client.get_flight_info(fl.FlightDescriptor.for_command(json.dumps(request)))
  reader: fl.FlightStreamReader = client.do_get(info.endpoints[0].ticket)
  df: pd.DataFrame = reader.read_pandas()
  ```
4. Insert a new notebook cell with the following command to visualize the result:
  ```
  df
  ```
5. Execute all notebook cells and notice that some of the columns appear redacted.
