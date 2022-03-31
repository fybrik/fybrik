# Notebook sample to demonstrate write flow

This sample shows how Fybrik enables a Jupyter notebook workload to read and write a dataset: a data that resides in an object storage is first been read into the notebook dataset and then written into a new location in the object storage. It is assumed that before the fybrik application is launched the new location already exists in the object store and the catalog holds the information about the new written asset.
This sample demonstrates how policies are seamlessly applied when accessing the dataset classified as financial data.

In this sample you play multiple roles:

1. As a data owner you upload a dataset and register it in a data catalog
2. As a data steward you setup data governance policies
3. As a data user you specify your data usage requirements and use a notebook to consume the data

## Before you begin

- Install Fybrik using the [Quick Start](../get-started/quickstart.md) guide.
  This sample assumes the use of the built-in catalog, Open Policy Agent (OPA) and flight module.
- A web browser.

## Create a namespace for the sample

Create a new Kubernetes namespace and set it as the active namespace:

```bash
kubectl create namespace fybrik-notebook-sample
kubectl config set-context --current --namespace=fybrik-notebook-sample
```

This enables easy [cleanup](#cleanup) once you're done experimenting with the sample.

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
      helm install localstack localstack-charts/localstack --set startServices="s3" --set service.type=ClusterIP
      kubectl wait --for=condition=ready --all pod -n fybrik-notebook-sample --timeout=120s
      ```
    3. Create a port-forward to communicate with localstack server:
      ```bash
      kubectl port-forward svc/localstack 4566:4566 &
      ```
    3. Use [AWS CLI](https://aws.amazon.com/cli/) to upload the dataset to a new created bucket in the localstack server:
      ```bash
      export ENDPOINT="http://127.0.0.1:4566"
      export ORIG_BUCKET="demo"
      export NEW_BUCKET="bucket1"
      export OBJECT_KEY="PS_20174392719_1491204439457_log.csv"
      export FILEPATH="/path/to/PS_20174392719_1491204439457_log.csv"
      aws configure set aws_access_key_id ${ACCESS_KEY} && aws configure set aws_secret_access_key ${SECRET_KEY} && aws --endpoint-url=${ENDPOINT} s3api create-bucket --bucket ${ORIG_BUCKET} && aws --endpoint-url=${ENDPOINT} s3api put-object --bucket ${ORIG_BUCKET} --key ${OBJECT_KEY} --body ${FILEPATH} && aws --endpoint-url=${ENDPOINT} s3api create-bucket --bucket ${NEW_BUCKET}
      ```
## Register the dataset in a data catalog

Register the credentials required for accessing the dataset. Replace the values for `access_key` and `secret_key` with the values from the object storage service that you used and run:

```yaml
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: paysim-secret
type: Opaque
stringData:
  access_key: "${ACCESS_KEY}"
  secret_key: "${SECRET_KEY}"
EOF
```

Then, register the data assets in the catalog. Replace the values for `endpoint`, `bucket` and `object_key` with values from the object storage service that you used and run:

```yaml
cat << EOF | kubectl apply -f -
apiVersion: katalog.fybrik.io/v1alpha1
kind: Asset
metadata:
  name: orig-paysim-csv
spec:
  secretRef: 
    name: paysim-secret
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
      finance: true
    columns:
      - name: nameOrig
        tags:
          PII: true
      - name: oldbalanceOrg
        tags:
          sensitive: true
      - name: newbalanceOrig
        tags:
          sensitive: true
EOF
```
```yaml
cat << EOF | kubectl apply -f -
apiVersion: katalog.fybrik.io/v1alpha1
kind: Asset
metadata:
  name: new-paysim-parquet
spec:
  secretRef:
    name: paysim-secret
  details:
    dataFormat: parquet
    connection:
      name: s3
      s3:
        endpoint: "http://localstack.fybrik-notebook-sample.svc.cluster.local:4566"
        bucket: "bucket1"
        object_key: ""
  metadata:
    name: Synthetic Financial Datasets For Fraud Detection
    geography: theshire
    tags:
      finance: true
    columns:
      - name: nameOrig
        tags:
          PII: true
      - name: oldbalanceOrg
        tags:
          sensitive: true
      - name: newbalanceOrig
        tags:
          sensitive: true
EOF
```
The assets are now registered in the catalog. The identifier of the original asset is `fybrik-notebook-sample/orig-paysim-csv` (i.e. `<namespace>/<name>`) and the identifier of the new asset is `fybrik-notebook-sample/new-paysim-parquet`. You will use these names in the `FybrikApplication` later.

Notice the `metadata` field above. It specifies the dataset geography and tags. These attributes can later be used in policies.

For example, in the yaml above, the `geography` is set to `theshire`, you need make sure it is same with the region of your fybrik control plane, you can get the information with the below command:

```shell
kubectl get configmap cluster-metadata -n fybrik-system -o 'jsonpath={.data.Region}'
```

[Quick Start](../get-started/quickstart.md) installs a fybrik control plane with the region `theshire` by default. If you change it or the `geography` in the yaml above, a [copy module](https://github.com/fybrik/mover) will be required by the policies, but we do not install any copy module in the [Quick Start](../get-started/quickstart.md).

## Define data access policies

Define an [OpenPolicyAgent](https://www.openpolicyagent.org/) policy to redact the columns tagged as `PII` in datasets tagged with `finance` when reading data and to redact the columns tagged as `sensitive` in datasets tagged with `finance` when writing data. Below is the policy (written in [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/#what-is-rego) language):

```rego
package dataapi.authz

rule[{"action": {"name":"RedactAction", "columns": column_names}, "policy": description}] {
  description := "Redact columns tagged as sensitive in datasets tagged with finance = true in write flow"
  input.action.actionType == "write"
  input.resource.metadata.tags.finance
  column_names := [input.resource.metadata.columns[i].name | input.resource.metadata.columns[i].tags.sensitive]
  count(column_names) > 0
}

rule[{"action": {"name":"RedactAction", "columns": column_names}, "policy": description}] {
  description := "Redact columns tagged as PII in datasets tagged with finance = true in read flow"
  input.action.actionType == "read"
  input.resource.metadata.tags.finance
  column_names := [input.resource.metadata.columns[i].name | input.resource.metadata.columns[i].tags.PII]
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

Refer to `Deploy a Jupyter notebook` section in [fybrik notebook sample](https://fybrik.io/v0.6/samples/notebook/).

## Create a `FybrikApplication` resource for the notebook

Create a [`FybrikApplication`](../reference/crds.md#fybrikapplication) resource to register the notebook workload to the control plane of Fybrik: 

<!-- TODO: role field removed but code still requires it -->
```yaml
cat <<EOF | kubectl apply -f -
apiVersion: app.fybrik.io/v1alpha1
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
    - dataSetID: "fybrik-notebook-sample/orig-paysim-csv"
      requirements:
        interface: 
          protocol: fybrik-arrow-flight
    - dataSetID: "fybrik-notebook-sample/new-paysim-parquet"
      flow: write
      requirements:
        interface:
          protocol: fybrik-arrow-flight
    - dataSetID: "fybrik-notebook-sample/new-paysim-parquet"
      requirements:
        interface:
          protocol: fybrik-arrow-flight

EOF
```

Notice that:

* The `selector` field matches the labels of our Jupyter notebook workload.
* The `data` field includes a `dataSetID` that matches the asset identifier in the catalog.
* The `protocol` indicates that the developer wants to consume the data using Apache Arrow Flight. For some protocols a `dataformat` can be specified as well (e.g., `s3` protocol and `parquet` format).


Run the following command to wait until the `FybrikApplication` is ready:

```bash
while [[ $(kubectl get fybrikapplication my-notebook -o 'jsonpath={.status.ready}') != "true" ]]; do echo "waiting for FybrikApplication" && sleep 5; done
```

## Read the dataset from the notebook

In your **terminal**, run the following command to print the [endpoints](../../reference/crds/#fybrikapplicationstatusreadendpointsmapkey)(*) to use for reading the data. It fetches the code from the `FybrikApplication` resource:
```bash
ENDPOINT_SCHEME=$(kubectl get fybrikapplication my-notebook -o jsonpath={.status.assetStates.fybrik-notebook-sample/orig-paysim-csv.endpoint.fybrik-arrow-flight.scheme})
ENDPOINT_HOSTNAME=$(kubectl get fybrikapplication my-notebook -o jsonpath={.status.assetStates.fybrik-notebook-sample/orig-paysim-csv.endpoint.fybrik-arrow-flight.hostname})
ENDPOINT_PORT=$(kubectl get fybrikapplication my-notebook -o jsonpath={.status.assetStates.fybrik-notebook-sample/orig-paysim-csv.endpoint.fybrik-arrow-flight.port})
printf "${ENDPOINT_SCHEME}://${ENDPOINT_HOSTNAME}:${ENDPOINT_PORT}"
```

* For the simplicity of the demonstration only the original dataset endpoint is been extracted.

The next steps use the endpoint to read the data in a python notebook

1. Insert a new notebook cell to install pandas and pyarrow packages:
  ```python
  %pip install pandas pyarrow numpy
  ```
2. Insert a new notebook cell to read the data using the endpoint value extracted from the `FybrikApplication` in the previous step:
  ```bash
  import json
import pyarrow.flight as fl
import pandas as pd
import pyarrow as pa
import numpy

# Create a Flight client
client = fl.connect('<ENDPOINT>')

# Prepare the request to read the original data
request = {
    "asset": "fybrik-notebook-sample/orig-paysim-csv",
    # To request specific columns add to the request a "columns" key with a list of column names
    # "columns": [...]
}
# Send request and fetch result
info = client.get_flight_info(fl.FlightDescriptor.for_command(json.dumps(request)))
reader: fl.FlightStreamReader = client.do_get(info.endpoints[0].ticket)
df: pd.DataFrame = reader.read_pandas()

# Prepare the request to write the new data
request = {
    "asset": "fybrik-notebook-sample/new-paysim-parquet",
    # To request specific columns add to the request a "columns" key with a list of column names
    # "columns": [...]
}
# drop masked column
df.pop("nameOrig")
data = pa.Table.from_pandas(df)

# Add a new column and write the new data
new_col = numpy.add(data.column(7).to_pylist(), data.column(8).to_pylist())
new_col = [n / 2 for n in new_col]
new_data = data.append_column('avg', pa.array(new_col, pa.float32()))
writer, _  = client.do_put(fl.FlightDescriptor.for_command(json.dumps(request)), new_data.schema)
writer.write_table(new_data, 1024)
writer.close()

# Prepare the request to read the new data
request = {
    "asset": "fybrik-notebook-sample/new-paysim-parquet",
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
5. Execute all notebook cells and notice that the `nameOrig`, `oldbalanceOrg`and `newbalanceOrig` columns appear redacted.


## Cleanup

When you’re finished experimenting with the notebook sample, clean it up:

1. Stop `kubectl port-forward` processes (e.g., using `pkill kubectl`)
1. Delete the namespace created for this sample:
    ```bash
    kubectl delete namespace fybrik-notebook-sample
    ```
1. Delete the policy created on fybrik-system namespace:
    ```bash
    kubectl -n fybrik-system delete configmap sample-policy
    ```
