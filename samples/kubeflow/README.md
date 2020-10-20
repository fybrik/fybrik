# Kubeflow Jupyter notebook sample

This sample demonstrate the use of The Mesh for Data with Kubeflow notebooks.
The following sample and instructions have been tested on kind v0.10.0 and OpenShift 4.3 clusters, and against Kubeflow v1.0.2.
The instructions below assume that they are issued from the root directory of the project, and you have created a Kubernets cluster in advance (either Kind or OpenShift).

## Connectors stack

This sample uses the default stack:
- Credentials manager: Hashicorp Vault (aka Vault)
- Data catalog: ODPi Egeria (aka Egeria)
- Policy manager: Open Policy Agent (aka OPA)

## Actors and flows

A _data owner_ uploads a transactions.csv file to a s3 compatible object storage (e.g., Ceph, minio, IBM Cloud Object Storage, Amazon S3, etc.). Then, the data owner registers the asset in Egeria data catalog and adds tags to the asset to mark it as finance data. Finally, the data owner registers the credentials to access this asset in Vault credentials manager.

A _data steward_ creates policies in OPA policy manager. These policies set transformations to perform on data columns when accessed for the purpose of fraud-detection.

A _data user_ creates a notebook server in Kubeflow and a notebook with the business logic for creating a fraud detection model. The data user also creates a `M4DApplication` resource that references the transactions.csv asset from the data catalog. This allows the code in the notebook to load the data (after policies are applied to it).

## Installation

Run the following script to install The Mesh for Data core components and third party dependencies required for this sample.
Please note that it may take a while for the M4D containers to be ready as the container images may need to be downloaded.

```bash
kubectl config set-context --current --namespace=m4d-system
```

Install on OpenShift:
```bash
WITHOUT_OPENSHIFT=false ./hack/install.sh
```

**Or**

Install on Kind:
```bash
./hack/install.sh
```

Install the Arrow-Flight module
```
kubectl apply -f https://raw.githubusercontent.com/IBM/the-mesh-for-data-flight-module/master/module.yaml
```

Install kubeflow on your cluster.
On a Kind cluster for example you can use the following script to install it:
```bash
cd samples/kubeflow/install/kubeflow
./install_kubeflow.sh

cd -
```

## Data owner instructions

1. Upload [data.csv](data.csv) to an object-storage of your choice. For example, IBM [Cloud Object Storage](https://cloud.ibm.com/catalog/services/cloud-object-storage).

`data.csv` contains the first 100 rows from the following [data set](https://www.kaggle.com/ntnu-testimon/paysim1/data) created by NTNU, and it is shared under the ***CC BY-SA 4.0*** license.

2. Alter the `fullPath` field in the json `example_transactions.csv.json` to contain the details of your object storage location, etc.
See the comments in [third_pary/egeria/usage/create_new_asset.sh](../../third_party/egeria/usage/create_new_asset.sh) for more details.

3. Register the asset using the following command:

```bash

cd third_party/egeria/usage

# Wait for the port-forward to take effect before proceeding to the next command
kubectl port-forward -n egeria-catalog svc/lab-core 9443:9443 &


# path-to-transactions.csv.json can be for example ../../../samples/kubeflow/example_transactions.csv.json for using the example file
./create_new_asset.sh <path-to-transactions.csv.json> 'finance'

cd -
kill $!
```

4. Save the asset ID (example for asset-id is 5de27155-48d3-4d78-8767-73e7b264e394).
Export it to environment variable:
```
export ASSET_ID=<asset-id>
```

5. Register in Vault the S3 credentials `access_key` (aka `access_key_id`) and `secret_key` (aka `secret_access_key`) required to access the dataset.
Currently, only hmac credentials are supported.

In order to communicate with Vault you first need to create a port-forward:
```bash
kubectl port-forward -n m4d-system svc/vault 8200:8200 &
```

you can use your browser and Vault's UI to upload the credentials:
- Open `http://localhost:8200` in your browser and login using username `data_provider` and password `password`.
- Click `/external` and then `Create secret`
- Create the following secret:
    - Path for this secret: `{"ServerName":"cocoMDS3","AssetGuid":"<asset ID>"}`. For example, `{"ServerName":"cocoMDS3","AssetGuid":"5de27155-48d3-4d78-8767-73e7b264e394"}`
    - Secret data (shown here as JSON): `{"access_key": "<hmac-access-key-id>", "secret_key": "<hmac-secret-access-key>"}`
- Click `save`

Finally, kill the port-forward
```bash
kill $!
```

## Data steward instructions

TODO: currently the policies are hard coded with the OPA deployment


## Data user instructions


1. Create a port-forward to communicate with Kubeflow:
```bash
kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80 &

cd samples/kubeflow/
```

2. Upload the notebook:
- Open your browser in `http://localhost:8080`.
- Click **Start Setup** and then **Finish** (use the `anonymous` namespace).
- Click **Notebook Servers** (in the left).
- In the notebooks page select in the top left the `anonymous` namespace and then click **New Server**.
- In the notebook server creation page, set `kf-notebook` in the **Name** box and then click **Launch**. Wait for the server to become ready.
- Click **Connect** and upload `kfM4DPolicySample.ipynb` notebook to the server.

3. Create the `M4DApplication` resource by running the following:
```bash

cat m4dapplication.yaml | sed "s/ASSET_ID/$ASSET_ID/g" | kubectl -n anonymous apply -f -

cd -
```

4. Before running the notebook you need to modify the following statements in the `Get Data` cell in the notebook:
```python
...
client = fl.connect("grpc://<arrow-flight-module-service>.<arrow-flight-module-ns>.svc.cluster.local:80")

request = {
    "asset": "<bucket-name>/<file-name>.csv", 
    "columns": ["step", "type", "amount", "nameOrig", "oldbalanceOrg", "newbalanceOrig", "nameDest", "oldbalanceDest", "newbalanceDest", "isFraud", "isFlaggedFraud"]
}
...
``` 

Edit the `client = fl.connect(...)` command to point to the right service and namespace of the arrow-flight-module.
You can get these by running:
```bash
# Get the ns and the service name
kubectl get svc -l app.kubernetes.io/name=arrow-flight-module --all-namespaces
```

Edit `"asset": "<bucket-name>/<file-name>.csv"` in the second command to point to your bucket and the name of the file you uploaded.
5. run the notebook!
If everything worked according to plan you should see in the cel `Get Data` the data in the file you uploaded.


6. Finally, kill the port-forward
```bash
kill $!
```



