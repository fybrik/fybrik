# Notebook sample

This sample shows how Mesh for Data enables a Jupyter notebook workload to access a dataset.
It demonstrates how policies are seamlessly applied when accessing the dataset classified as financial data.

In this sample you play multiple roles:

1. As a data owner you upload a dataset and register it in a data catalog
2. As a data steward you setup data governance policies
3. As a data user you specify your data usage requirements and use a notebook to consume the data

## Before you begin

- Install Mesh for Data using the [Quick Start](../get-started/quickstart.md) guide.
  This sample assumes the use of the built-in catalog, Open Policy Agent (OPA) and flight module.
- A web browser.

## Create a namespace for the sample

Create a new Kubernetes namespace and set it as the active namespace:

```bash
kubectl create namespace m4d-notebook-sample
kubectl config set-context --current --namespace=m4d-notebook-sample
```

This enables easy [cleanup](#cleanup) once you're done experimenting with the sample.

## Prepare a dataset to be accessed by the notebook

This sample uses the [Synthetic Financial Datasets For Fraud Detection](https://www.kaggle.com/ealaxi/paysim1) dataset[^1] as the data that the notebook needs to read. Download and extract the file to your machine. You should now see a file named `PS_20174392719_1491204439457_log.csv`. Alternatively, use a sample of 100 lines of the same dataset by downloading [`PS_20174392719_1491204439457_log.csv`](https://raw.githubusercontent.com/mesh-for-data/mesh-for-data/master/samples/notebook/PS_20174392719_1491204439457_log.csv) from GitHub.

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
      kubectl wait --for=condition=ready --all pod -n m4d-notebook-sample --timeout=120s
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
      aws configure set aws_access_key_id ${ACCESS_KEY} && aws configure set aws_secret_access_key ${SECRET_KEY} && aws --endpoint-url=${ENDPOINT} s3api create-bucket --bucket ${BUCKET} && aws --endpoint-url=${ENDPOINT} s3api put-object --bucket ${BUCKET} --key ${OBJECT_KEY} --body ${FILEPATH}
      ```
## Register the dataset in a data catalog

Register the credentials required for accessing the dataset. Replace the values for `access_key` and `secret_key` with the values from the object storage service that you used and run:

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

Then, register the data asset itself in the catalog. Replace the values for `endpoint`, `bucket` and `objectKey` with values from the object storage service that you used and run:

```yaml
cat << EOF | kubectl apply -f -
apiVersion: katalog.m4d.ibm.com/v1alpha1
kind: Asset
metadata:
  name: paysim-csv
spec:
  secretRef: 
    name: paysim-csv
  assetDetails:
    dataFormat: csv
    connection:
      type: s3
      s3:
        endpoint: "http://localstack.m4d-notebook-sample.svc.cluster.local:4566"
        bucket: "demo"
        objectKey: "PS_20174392719_1491204439457_log.csv"
  assetMetadata:
    geography: theshire
    tags:
    - finance
    componentsMetadata:
      nameOrig: 
        tags:
        - PII
      oldbalanceOrg:
        tags:
        - sensitive
      newbalanceOrig:
        tags:
        - sensitive
EOF
```

The asset is now registered in the catalog. The identifier of the asset is `m4d-notebook-sample/paysim-csv` (i.e. `<namespace>/<name>`). You will use that name in the `M4DApplication` later.

Notice the `assetMetadata` field above. It specifies the dataset geography and tags. These attributes can later be used in policies.


## Define data access policies

Define an [OpenPolicyAgent](https://www.openpolicyagent.org/) policy to redact the `nameOrig` column for datasets tagged as `finance`. Below is the policy (written in [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/#what-is-rego) language):

```rego
package dataapi.authz

import data.data_policies as dp

transform[action] {
  description := "Redact sensitive columns in finance datasets"
  dp.AccessType() == "READ"
  dp.dataset_has_tag("finance")
  column_names := dp.column_with_any_name({"nameOrig"})
  action = dp.build_redact_column_action(column_names[_], dp.build_policy_from_description(description))
}
```

In this sample only the policy above is applied. Copy the policy to a file named `sample-policy.rego` and then run:

```bash
kubectl -n m4d-system create configmap sample-policy --from-file=sample-policy.rego
kubectl -n m4d-system label configmap sample-policy openpolicyagent.org/policy=rego
while [[ $(kubectl get cm sample-policy -n m4d-system -o 'jsonpath={.metadata.annotations.openpolicyagent\.org/policy-status}') != '{"status":"ok"}' ]]; do echo "waiting for policy to be applied" && sleep 5; done
```

You can similarly apply a directory holding multiple rego files.

## Deploy a Jupyter notebook

In this sample a Jupyter notebook is used as the user workload and its business logic requires reading the asset that we registered (e.g., for creating a fraud detection model). Deploy a notebook to your cluster:

=== "JupyterLab"

    1. Deploy JupyterLab:
        ```bash
        kubectl create deployment my-notebook --image=jupyter/base-notebook --port=8888 -- start.sh jupyter lab --LabApp.token=''
        kubectl set env deployment my-notebook JUPYTER_ENABLE_LAB=yes
        kubectl label deployment my-notebook app.kubernetes.io/name=my-notebook
        kubectl wait --for=condition=available --timeout=120s deployment/my-notebook
        kubectl expose deployment my-notebook --port=80 --target-port=8888
        ```
    1. Create a port-forward to communicate with JupyterLab:
        ```bash
        kubectl port-forward svc/my-notebook 8080:80 &
        ```
    1. Open your browser and go to [http://localhost:8080/](http://localhost:8080/).
    1. Create a new notebook in the server


=== "Kubeflow"

    1. Ensure that [Kubeflow](https://www.kubeflow.org/) is installed in your cluster
    1. Create a port-forward to communicate with Kubeflow:
        ```bash
        kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80 &
        ```
    1. Open your browser and go to [http://localhost:8080/](http://localhost:8080/).
    1. Click **Start Setup** and then **Finish** (use the `anonymous` namespace).
    1. Click **Notebook Servers** (in the left).
    1. In the notebooks page select in the top left the `anonymous` namespace and then click **New Server**.
    1. In the notebook server creation page, set `my-notebook` in the **Name** box and then click **Launch**. Wait for the server to become ready.
    1. Click **Connect** and create a new notebook in the server.


## Create a `M4DApplication` resource for the notebook

Create a [`M4DApplication`](../reference/crds.md#m4dapplication) resource to register the notebook workload to the control plane of Mesh for Data: 

<!-- TODO: role field removed but code still requires it -->
```yaml
cat <<EOF | kubectl apply -f -
apiVersion: app.m4d.ibm.com/v1alpha1
kind: M4DApplication
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
    intent: fraud-detection
  data:
    - dataSetID: "m4d-notebook-sample/paysim-csv"
      requirements:
        interface: 
          protocol: m4d-arrow-flight
          dataformat: arrow
EOF
```

Notice that:

* The `selector` field matches the labels of our Jupyter notebook workload.
* The `data` field includes a `dataSetID` that matches the asset identifier in the catalog.
* The `protocol` and `dataformat` indicate that the developer wants to consume the data using Apache Arrow Flight.


Run the following command to wait until the `M4DApplication` is ready:

```bash
while [[ $(kubectl get m4dapplication my-notebook -o 'jsonpath={.status.ready}') != "true" ]]; do echo "waiting for M4DApplication" && sleep 5; done
```

## Read the dataset from the notebook

In your **terminal**, run the following command to print the [endpoint](../../reference/crds/#m4dapplicationstatusreadendpointsmapkey) to use for reading the data. It fetches the code from the `M4DApplication` resource:
```bash
ENDPOINT_SCHEME=$(kubectl get m4dapplication my-notebook -o jsonpath={.status.readEndpointsMap.m4d-notebook-sample/paysim-csv.scheme})
ENDPOINT_HOSTNAME=$(kubectl get m4dapplication my-notebook -o jsonpath={.status.readEndpointsMap.m4d-notebook-sample/paysim-csv.hostname})
ENDPOINT_PORT=$(kubectl get m4dapplication my-notebook -o jsonpath={.status.readEndpointsMap.m4d-notebook-sample/paysim-csv.port})
printf "${ENDPOINT_SCHEME}://${ENDPOINT_HOSTNAME}:${ENDPOINT_PORT}"
```
The next steps use the endpoint to read the data in a python notebook

1. Insert a new notebook cell to install pandas and pyarrow packages:
  ```python
  %pip install pandas pyarrow
  ```
2. Insert a new notebook cell to read the data using the endpoint value extracted from the `M4DApplication` in the previous step:
  ```bash
  %pip install pandas pyarrow
  import json
  import pyarrow.flight as fl
  import pandas as pd

  # Create a Flight client
  client = fl.connect('<ENDPOINT>')

  # Prepare the request
  request = {
      "asset": "m4d-notebook-sample/paysim-csv",
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
5. Execute all notebook cells and notice that the `nameOrig` column appears redacted.


## Cleanup

When you’re finished experimenting with the notebook sample, clean it up:

1. Stop `kubectl port-forward` processes (e.g., using `pkill kubectl`)
1. Delete the namespace created for this sample:
    ```bash
    kubectl delete namespace m4d-notebook-sample
    ```
