# Kubeflow Jupyter notebook demo

This demo demonstrate the use of The Mesh for Data with Kubeflow notebooks.

## Connectors stack

This demo uses the default stack:
- Credentials manager: Hashicorp Vault (aka Vault)
- Data catalog: ODPi Egeria (aka Egeria)
- Policy manager: Open Policy Agent (aka OPA)

## Actors and flows

A _data owner_ uploads a transactions.csv file to an s3 compatible object storage (e.g., Ceph, minio, IBM Cloud Object Storage, Amazon S3, etc.). Then, the data owner registers the asset in Egeria data catalog and adds tags to the asset to mark it as finance data. Finally, the data owner registers the credentials to access this asset in Vault credentials manager.

A _data steward_ creates policies in OPA policy manager. These policies set transformations to perform on data columns when accessed for the purpose of fraud-detection.

A _data user_ creates a notebook server in Kubeflow and a notebook with the business logic for creating a fraud detection model. The data user also creates a `M4DApplication` resource that references the transactions.csv asset from the data catalog. This allows the code in the notebook to load the data (after policies are applied to it).

## Installation

Run the following script to install The Mesh for Data and all dependencies required for this demo. It will take a while for all containers to be ready due to image downloads.


```bash
kind create cluster
./install/install.sh
```

## Data owner instructions

We already have transactions.csv stored in COS (TODO: don't assume that).
Register the asset using the following command:

```bash
kubectl port-forward -n egeria-catalog svc/lab-core 9443:9443 &
sleep 5
cd dataowner
./create_new_asset.sh transactions.csv.json 'finance'
cd -
kill $!
```

Note the asset ID that is listed next to "Read asset info: ". You will need it later:
```
export ASSET_ID=<asset-id>
```

The credentials for the asset need to be registered in Vault. 
You can do that using the Vault UI. First use this command to be able to open the UI from the browser:

```bash
kubectl port-forward -n m4d-system svc/vault 8200:8200
```

Now open "http://localhost:8200" and login using username `data_provider` and password `password`. Click **/external** and then **Create secret**. In the screen opened add the following:
1. **Path for this secret**: `{"ServerName":"cocoMDS3","AssetGuid":"<asset ID>"}`
1. **Secret data** (shown here as JSON): `{"access_key_id": "<ask me>", "secret_access_key": "<ask me>"}`


## Data steward instructions

TODO: currently the policies are hard coded with the OPA deployment


## Data user instructions

Port forward the istio ingress gateway to be able to open Kubeflow home.

```bash
kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80
```

Then open you browser in "http://localhost:8080". Click **Start Setup** and then **Finish**. Then, click **Notebook Servers**. In the notebooks page select in the top the anonymous namespace and then click **New Server**.

In the notebook server creation page set `kf-notebook` in the **Name** box and then click **Launch**. Wait for the server to become ready.

Click **Connect** and upload `notebook.ipynb` to the server.

Create the `M4DApplication` resource by running the following (replace `<asset ID>` with the actual asset guid:

```
echo $ASSET_ID
cat m4dapplication.yaml | sed "s/ASSET_ID/$ASSET_ID/g" | kubectl -n anonymous apply -f -
```