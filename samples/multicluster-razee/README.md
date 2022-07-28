# Notebook sample in multicluster

This sample shows how Fybrik enables a Jupyter notebook workload to access a dataset across multicluster. It demonstrates how policies are seamlessly applied when accessing the dataset classified as financial data and how the dataset is transfered across multicluster based on policies.

In this sample you play multiple roles:

As a data owner you upload a dataset and register it in a data catalog
As a data steward you setup data governance policies
As a data user you specify your data usage requirements and use a notebook to consume the data.

## Before you begin

Ensure that you have the following installed on your machine:

- git
- make
- jq
- unzip
- [Go](https://go.dev/) 1.16 or greater 
- [Helm](https://helm.sh/) 3.7 or greater
- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.18 or greater
- [Kind](http://kind.sigs.k8s.io/)
- A web browser

## Setup the multicluster environment

We leverage the scripts in the Fybrik repository to setup a [Razee](https://razee.io/) multicluster for this sample. First we download the repository:

```bash
git clone https://github.com/fybrik/fybrik.git
cd fybrik
```

We want to have a use case that the dataset is transfered accross different clusters. So change the region of the data plane cluster `kind-kind` from `homeoffice` to `theshire`:

```bash
sed -i 's/homeoffice/theshire/g' ./charts/fybrik/kind-kind.values.yaml
```

Check whether the regions of the two clusters, `kind-control` and `kind-kind` are different:

```bash
grep 'region:' ./charts/fybrik/kind-control.values.yaml
grep 'region:' ./charts/fybrik/kind-kind.values.yaml
```

Remove the use of the mockup images in control plane to trigger the built-in catalog and Open Policy Agent (OPA):

```bash
sed -i '/Connector:/d' ./charts/fybrik/kind-control.values.yaml
sed -i '/-mock/d' ./charts/fybrik/kind-control.values.yaml
```

Trigger the docker-build instead of building the mockup images:

```bash
sed -i 's/docker-minimal-it/docker-build docker-push/g' ./hack/setup-local-multi-cluster.sh
```

Add the [externed policies](https://fybrik.io/dev/concepts/config-policies/#extended-policies) to meet advanced deployment requirements, such as where read or transform modules should run, what should be the scope of module deployments, and more.

```bash
cp ./samples/adminconfig/quickstart_policies.rego ./charts/fybrik/files/adminconfig/
```

Now we can execute the script to setup a Razee multicluster automatically:

```bash
sh ./hack/setup-local-multi-cluster.sh
```

You can also manually execute the commands in this script one by one.

This script will setup two Kind clusters, install and configure Razee, Vault, Cert-manager, Datashim to them. After the script finishes running, we will get a multicluster environment consisting of two clusters:

- kind-control
- kind-kind

## Prepare the operator and modules

Next, install the [Data Movement Operator](https://github.com/fybrik/data-movement-operator) on the data plane cluster to orchestrate the data movement.

Get the code:

```bash
git clone https://github.com/fybrik/data-movement-operator
cd data-movement-operator
```

Install the operator on the cluster `kind-kind`:

```bash
export DOCKER_HOSTNAME=localhost:5000
export DOCKER_NAMESPACE=fybrik-system
export VALUES_FILE=charts/data-movement-operator/integration-tests.values.yaml

kubectl config use-context kind-kind
make docker-build docker-push
make deploy
```

Register [Implicit Copy Batch Module](https://github.com/fybrik/data-movement-operator/tree/master/modules/fybrik-implicit-copy-batch) and [Implicit Copy Stream Module](https://github.com/fybrik/data-movement-operator/tree/master/modules/fybrik-implicit-copy-stream) as copy modules on the control plane:

```bash
kubectl config use-context kind-control
kubectl apply -f modules/implicit-copy-batch-module.yaml -n fybrik-system
kubectl apply -f modules/implicit-copy-stream-module.yaml -n fybrik-system
```

Register the [Arrow Flight Module](https://github.com/fybrik/arrow-flight-module) as a read module on the control plane:

```bash
kubectl config use-context kind-control
kubectl apply -f https://raw.githubusercontent.com/fybrik/arrow-flight-module/master/module.yaml -n fybrik-system
```

## Create a namespace for the sample

Create a new Kubernetes namespace and set it as the active namespace in two clusters:

```bash
kubectl config use-context kind-control
kubectl create namespace fybrik-notebook-sample
kubectl config set-context --current --namespace=fybrik-notebook-sample

kubectl config use-context kind-kind
kubectl create namespace fybrik-notebook-sample
kubectl config set-context --current --namespace=fybrik-notebook-sample
```

## Prepare a dataset to be accessed by the notebook

This sample uses the [Synthetic Financial Datasets For Fraud Detection](https://www.kaggle.com/ealaxi/paysim1) dataset[^1] as the data that the notebook needs to read. Download and extract the file to your machine. You should now see a file named `PS_20174392719_1491204439457_log.csv`. Alternatively, use a sample of 100 lines of the same dataset by downloading [`PS_20174392719_1491204439457_log.csv`](https://raw.githubusercontent.com/fybrik/fybrik/master/samples/notebook/PS_20174392719_1491204439457_log.csv) from GitHub.

[^1]: Created by NTNU and shared under the ***CC BY-SA 4.0*** license.

Upload the CSV file to an object storage of your choice such as AWS S3, IBM Cloud Object Storage or Ceph.

For experimentation we install localstack to cluster instead of using a cloud service.

Setup localstack on both clusters, `kind-control` and `kind-kind`.

1. Define variables for access key and secret key

    ```bash
    export ACCESS_KEY="myaccesskey"
    export SECRET_KEY="mysecretkey"
    ```

2. Install localstack to the currently active namespace and wait for it to be ready:

    ```bash
    kubectl config use-context kind-control
    helm repo add localstack-charts https://localstack.github.io/helm-charts
    helm install localstack localstack-charts/localstack --set startServices="s3" --set service.type=ClusterIP
    kubectl wait --for=condition=ready --all pod -n fybrik-notebook-sample --timeout=120s

    kubectl config use-context kind-kind
    helm repo add localstack-charts https://localstack.github.io/helm-charts
    helm install localstack localstack-charts/localstack --set startServices="s3" --set service.type=ClusterIP
    kubectl wait --for=condition=ready --all pod -n fybrik-notebook-sample --timeout=120s
    ```

Upload the CSV file to localstack only on cluster `kind-kind`

1. Create a port-forward to communicate with localstack server:

    ```bash
    kubectl config use-context kind-kind
    kubectl port-forward svc/localstack 4566:4566 &
    ```

2. Use [AWS CLI](https://aws.amazon.com/cli/) to upload the dataset to a new created bucket in the localstack server:

    ```bash
    export ENDPOINT="http://127.0.0.1:4566"
    export BUCKET="demo"
    export OBJECT_KEY="PS_20174392719_1491204439457_log.csv"
    export FILEPATH="/path/to/PS_20174392719_1491204439457_log.csv"
    aws configure set aws_access_key_id ${ACCESS_KEY} && aws configure set aws_secret_access_key ${SECRET_KEY} && aws --endpoint-url=${ENDPOINT} s3api create-bucket --bucket ${BUCKET} && aws --endpoint-url=${ENDPOINT} s3api put-object --bucket ${BUCKET} --key ${OBJECT_KEY} --body ${FILEPATH}
    ```

Register the credentials required for accessing the dataset. Replace the values for `access_key` and `secret_key` with the values from the object storage service that you used and run:

```bash
kubectl config use-context kind-control

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

To enable pods on cluster `kind-kind` to access the localstack on cluster `kind-control`, we setup a NodePort service to export the localstack service on the cluster `kind-control`:

```bash
kubectl config use-context kind-control

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: localstack-kind-control-np
  namespace: fybrik-notebook-sample
spec:
  type: NodePort
  selector:
    app.kubernetes.io/instance: localstack
    app.kubernetes.io/name: localstack
  ports:
  - protocol: TCP
    port: 4566
    targetPort: 4566
    nodePort: 30566
EOF
```

Register a `FybrikStorageAccount` on the cluster `kind-control` for creating AWS S3 objects in localstack of region `homeoffice`:

```bash
kubectl config use-context kind-control

cat << EOF | kubectl apply -f -
apiVersion:   app.fybrik.io/v12
kind:         FybrikStorageAccount
metadata:
  name: storage-account
  namespace: fybrik-system
spec:
  id: homeoffice
  endpoints:
    homeoffice: "http://control-control-plane:30566"
  secretRef:  bucket-creds
EOF
```

Register the credentials required by `FybrikStorageAccount/storage-account`.

```bash
kubectl config use-context kind-kind

cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: bucket-creds
  namespace: fybrik-system
type: Opaque
stringData:
  access_key: "${ACCESS_KEY}"
  accessKeyID: "${ACCESS_KEY}"
  secret_key: "${SECRET_KEY}"
  secretAccessKey: "${SECRET_KEY}"
EOF
```

And then, register the data asset itself in the catalog on control plane. Replace the values for `endpoint`, `bucket` and `object_key` with values from the object storage service that you used and run:

```bash
kubectl config use-context kind-control

cat << EOF | kubectl apply -f -
apiVersion: katalog.fybrik.io/v12
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
      finance: true
    columns:
      - name: nameOrig
        tags:
          PII: true
      - name: oldbalanceOrg
        tags:
          PII: true
      - name: newbalanceOrig
        tags:
          PII: true
EOF
```

Notice the `metadata` field above. It specifies the dataset geography and tags. These attributes can later be used in policies.

In the yaml above, the `geography` is set to `theshire`, it is same with the region of cluster `kind-kind`, but is different with the region of cluster `kind-control`.

you can get the region of cluster with the following command:

```shell
kubectl get configmap cluster-metadata -n fybrik-system -o 'jsonpath={.data.Region}'
```

## Define data access policies

Define an [Open Policy Agent](https://www.openpolicyagent.org/) policy to redact the columns tagged as `PII` for datasets tagged with `finance`. Below is the policy (written in [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/#what-is-rego) language):

```rego
package dataapi.authz

rule[{"action": {"name":"RedactAction", "columns": column_names}, "policy": description}] {
  description := "Redact columns tagged as PII in datasets tagged with finance = true"
  input.action.actionType == "read"
  input.resource.metadata.tags.finance
  column_names := [input.resource.metadata.columns[i].name | input.resource.metadata.columns[i].tags.PII]
  count(column_names) > 0
}

rule[{"action": {"name":"RedactAction", "columns": column_names}, "policy": description}] {
  description := "Redact columns tagged as PII in datasets tagged with finance = true"
  input.action.actionType == "write"
  input.resource.metadata.tags.finance
  column_names := [input.resource.metadata.columns[i].name | input.resource.metadata.columns[i].tags.PII]
  count(column_names) > 0
}
```

In this sample only the policy above is applied. Copy the policy to a file named `sample-policy.rego` and then run:

```bash
kubectl config use-context kind-control
kubectl -n fybrik-system create configmap sample-policy --from-file=sample-policy.rego
kubectl -n fybrik-system label configmap sample-policy openpolicyagent.org/policy=rego
while [[ $(kubectl get cm sample-policy -n fybrik-system -o 'jsonpath={.metadata.annotations.openpolicyagent\.org/policy-status}') != '{"status":"ok"}' ]]; do echo "waiting for policy to be applied" && sleep 5; done
```

You can similarly apply a directory holding multiple rego files.

## Deploy a Jupyter notebook

In this sample a Jupyter notebook is used as the user workload and its business logic requires reading the asset that we registered (e.g., for creating a fraud detection model). Deploy a notebook to the cluster `kind-control`:

1. Deploy JupyterLab:

    ```bash
    kubectl config use-context kind-control
    kubectl create deployment my-notebook --image=jupyter/base-notebook --port=8888 -- start.sh jupyter lab --LabApp.token=''
    kubectl set env deployment my-notebook JUPYTER_ENABLE_LAB=yes
    kubectl label deployment my-notebook app.kubernetes.io/name=my-notebook
    kubectl wait --for=condition=available --timeout=120s deployment/my-notebook
    kubectl expose deployment my-notebook --port=8888 --target-port=8888
    ```

2. Create a port-forward to communicate with JupyterLab:

    ```bash
    kubectl port-forward svc/my-notebook 8888:8888 &
    ```

3. Open your browser and go to [http://localhost:8888/](http://localhost:8888/).

4. Create a new notebook in the server

## Create a `FybrikApplication` resource for the notebook

Create a [`FybrikApplication`](https://fybrik.io/dev/reference/crds/#fybrikapplication) resource to register the notebook workload to the control plane of Fybrik:

```yaml
kubectl config use-context kind-control

cat <<EOF | kubectl apply -f -
apiVersion: app.fybrik.io/v12
kind: FybrikApplication
metadata:
  name: my-notebook
  labels:
    app: my-notebook
spec:
  selector:
    clusterName: kind-control
    workloadSelector:
      matchLabels:
        app: my-notebook
  appInfo:
    intent: Fraud Detection
  data:
    - dataSetID: "fybrik-notebook-sample/paysim-csv"
      requirements:
        interface: 
          protocol: fybrik-arrow-flight
EOF
```

Notice that:

* The `selector/clusterName` field specifies the name of cluster which our Jupyter notebook workload is running on.
* The `selector/workloadSelector` field matches the labels of our Jupyter notebook workload.
* The `data` field includes a `dataSetID` that matches the asset identifier in the catalog.
* The `protocol` indicates that the developer wants to consume the data using Apache Arrow Flight. For some protocols a `dataformat` can be specified as well (e.g., `s3` protocol and `parquet` format).

Run the following command to wait until the `FybrikApplication` is ready:

```bash
while [[ $(kubectl get fybrikapplication my-notebook -o 'jsonpath={.status.ready}') != "true" ]]; do echo "waiting for FybrikApplication" && sleep 5; done
while [[ $(kubectl get fybrikapplication my-notebook -o 'jsonpath={.status.assetStates.fybrik-notebook-sample/paysim-csv.conditions[?(@.type == "Ready")].status}') != "True" ]]; do echo "waiting for fybrik-notebook-sample/paysim-csv asset" && sleep 5; done
```

## Read the dataset from the notebook

In your **terminal**, run the following command to print the endpoint to use for reading the data. It fetches the code from the `FybrikApplication` resource:

```bash
kubectl config use-context kind-control
ENDPOINT_SCHEME=$(kubectl get fybrikapplication my-notebook -o jsonpath={.status.assetStates.fybrik-notebook-sample/paysim-csv.endpoint.fybrik-arrow-flight.scheme})
ENDPOINT_HOSTNAME=$(kubectl get fybrikapplication my-notebook -o jsonpath={.status.assetStates.fybrik-notebook-sample/paysim-csv.endpoint.fybrik-arrow-flight.hostname})
ENDPOINT_PORT=$(kubectl get fybrikapplication my-notebook -o jsonpath={.status.assetStates.fybrik-notebook-sample/paysim-csv.endpoint.fybrik-arrow-flight.port})
printf "${ENDPOINT_SCHEME}://${ENDPOINT_HOSTNAME}:${ENDPOINT_PORT}\n"
```

The next steps use the endpoint to read the data in a python notebook

1. Insert a new notebook cell to install pandas and pyarrow packages:

    ```python
    %pip install pandas pyarrow
    ```

2. Insert a new notebook cell to read the data using the endpoint value extracted from the `FybrikApplication` in the previous step:

    ```python
    import json
    import pyarrow.flight as fl
    import pandas as pd

    # Create a Flight client
    client = fl.connect('<ENDPOINT>')

    # Prepare the request
    request = {
        "asset": "fybrik-notebook-sample/paysim-csv",
        # To request specific columns add to the request a "columns" key with a list of column names
        # "columns": [...]
    }

    # Send request and fetch result as a pandas DataFrame
    info = client.get_flight_info(fl.FlightDescriptor.for_command(json.dumps(request)))
    reader: fl.FlightStreamReader = client.do_get(info.endpoints[0].ticket)
    df: pd.DataFrame = reader.read_pandas()
    ```

3. Insert a new notebook cell with the following command to visualize the result:

    ```python
    df
    ```

4. Execute all notebook cells and notice that the `nameOrig`, `oldbalanceOrg`and `newbalanceOrig` columns appear redacted.

## Cleanup

When you finish this sample, clean it up:

1. Stop `kubectl port-forward` processes (e.g., using `pkill kubectl`)

2. Delete the namespaces created for this sample:

    ```bash
    kubectl config use-context kind-control
    kubectl delete namespace fybrik-notebook-sample

    kubectl config use-context kind-kind
    kubectl delete namespace fybrik-notebook-sample
    ```

3. Delete the resources created in fybrik-system namespace:

    ```bash
    kubectl config use-context kind-control
    kubectl -n fybrik-system delete configmap sample-policy
    kubectl -n fybrik-system delete FybrikStorageAccount storage-account
    kubectl -n fybrik-system delete Secret bucket-creds
    ```

## Under the hood

Now you have finished experimenting with the notebook sample, lets review this sample. We setup a multicluster environment consisting of two clusters, `kind-control` and `kind-kind`. They are located in the region `homeoffice` and `theshire` respectively.

We deploy [extended policies](https://fybrik.io/dev/concepts/config-policies/#extended-policies) for implict copy, locations of read and transformation operations, etc.

We upload a dataset to the localstack on the cluster `kind-kind` and register the dataset as a data asset with the geography `theshire` and the tag `finance`.

And then we define an [Open Policy Agent](https://www.openpolicyagent.org/) policy to redact the columns tagged as `PII` for datasets tagged with `finance`.

Finally we create a `FybrikApplication` which specifies the Jupyter notebook which runs on the cluster `kind-control` in the `homeoffice` region will consume the dataset.

At the runtime, the Jupyter notebook on cluster `kind-control` requires a data connection `fybrik-arrow-flight`. So the Fybrik control plane deploys an Arrow Flight Module on `kind-control`.

Due to the [extended policies](https://fybrik.io/dev/concepts/config-policies/#extended-policies), Arrow Flight Module is not allowed to read the data in `theshire`, so the Fybrik control plane allocates a storage space in localstack of `homeoffice` and deploys a copy module, Implicit Copy Batch Module on cluster `kind-kind`. The module generates a custom resource `BatchTransfer` to specify the source and destination of data replication.

[Data Movement Operator](https://github.com/fybrik/data-movement-operator) on the cluster `kind-kind` orchestrates the `BatchTransfers` and generates jobs.

These jobs trigger pods running [mover](https://github.com/fybrik/mover) to get the dataset, redact columns based on policies in `theshire`,  upload the dataset to `homeoffice` for Arrow Flight Module to read.

Finally the Jupyter notebook can read the data from the endpoints provided by the Arrow Flight Module.
