---
title: Kubeflow notebook sample
weight: 10
---

This sample shows how to run a Kubeflow notebook with {{< name >}} and demonstrates how polices are seamlessly applied when accessing a dataset.

## Before you begin

Ensure that you have the following:
- `kubectl` with access to a Kubernetes cluster (this guide was tested with kind v0.10.0 and OpenShift 4.3)
- {{< name >}} installed on your Kubernetes cluster
- S3 Object storage account (e.g., Ceph, Minio, IBM Cloud Object Storage)
- [Kubeflow](https://www.kubeflow.org/) installed on your cluster (this guide was tested with Kubeflow v1.0.2)
- Web browser


{{< tip >}}
You can install Kubeflow on Kind by running [install_kubeflow.sh](https://{{< github_base >}}/{{< github_repo >}}/blob/master/samples/kubeflow/install/kubeflow/install_kubeflow.sh). 
For OpenShift follow [KF_OPENSHIFT.md](https://{{< github_base >}}/{{< github_repo >}}/blob/master/samples/kubeflow/install/kubeflow/KF_OPENSHIFT.md).
{{</ tip >}}

## About this sample
In this sample guide you will run a Kubeflow notebook with {{< name >}} and demonstrate that data read polices that are defined in Open Policy Agent (OPA) are seamlessly applied when reading a dataset.

In this sample guide you will:
1. Prepare a dataset to be accessed by the notebook
1. Register the dataset in ODPi Egeria catalog
1. Register the dataset credentials in Vault
1. Deploy a Kubeflow notebook
1. Create a {{< name >}} runtime environment for the notebook
1. Read the dataset and observe policies applied seamlessly

## Getting started

Completed all the steps in [quick start guide]({{< baseurl >}}/docs/setup/quickstart/).

## Prepare the dataset for the sample notebook

1. Upload [data.csv](https://{{< github_base >}}/{{< github_repo >}}/blob/master/samples/kubeflow/data.csv) to an object storage of your choice
    `data.csv` contains the first 100 rows from the following [data set](https://www.kaggle.com/ntnu-testimon/paysim1/data) created by NTNU, and it is shared under the ***CC BY-SA 4.0*** license.
1. Update ```samples/kubeflow/example_transactions.csv.json``` with the location of the dataset, The location is encoded in the `fullPath` field as follows:
    - Bucket: Change the `bucket` value from `m4d-bucket-example` to the bucket name where the dataset reside.
    - endpoint: Change the S3 endpoint name from `s3.eu-de.cloud-object-storage.appdomain.cloud` to the object storage endpoint
    - object_key: If needed, change object_key from `data.csv` to the object name that you used in the previous step.
    For more information on the content please see the comments in [third_pary/egeria/usage/create_new_asset.sh](https://{{< github_base >}}/{{< github_repo >}}/blob/master//third_party/egeria/usage/create_new_asset.sh) for more details.
1. Register the dataset in the catalog

    - Setup port forwarding for communicating with Egeria

        ```bash
        cd third_party/egeria/usage
        kubectl port-forward -n egeria-catalog svc/lab-core 9443:9443 &
        ```
    - Wait for the port-forward to take effect.

    - Register the dataset in the catalog with the tag 'finance'.
    
        ```bash
        ./create_new_asset.sh ../../../samples/kubeflow/example_transactions.csv.json 'finance'
        ```
    - Cleanup the port forwarding using the following.
    
        ```bash
        cd -
        kill $!
        ```
1. Record the asset id for the dataset. It will be displayed as part of the output from the previous step and export it to environment variable. An example for an asset id is `5de27155-48d3-4d78-8767-73e7b264e394`
    ```
    export ASSET_ID=<asset-id>
    ```
1. Store the object dataset credentials in Vault to make them available for {{< name >}}. Currently only hmac credentials are supported, and the `access_key` (a.k.a `access_key_id`) and `secret_key` (a.k.a `secret_access_key`) should be associated with the asset id.

    You can register the credentials using a browser and Vault's UI to upload the credentials.

    - Setup port forwarding to communicate with Vault.
        ```bash
        kubectl port-forward -n m4d-system svc/vault 8200:8200 &
        ```
    - Open `http://localhost:8200` a your browser, select `method` as `username` and login using username `data_provider` and password `password`.

    - Click `/external` and then `Create secret`

    - Create the following secret:
        - Path for this secret: `{"ServerName":"cocoMDS3","AssetGuid":"<asset ID>"}`. For example, `{"ServerName":"cocoMDS3" , "AssetGuid":"5de27155-48d3-4d78-8767-73e7b264e394"}`
        - Secret data key: `<access-key-id>`
        - Secrey data value: `<secret-access-key>`
    - Click `save`
    
    Note: The path is a reference to the Egeria metadata server and asset id. In the default Egeria installation `cocoMDS3` is the metadata server name.

    - Finally, kill the port-forward

        ```bash
        kill $!
        ```

## Reviewing the policies for the dataset

Currently predefined policies are included as part of the OPA deployment.
Included are policies that are triggered for datasets that are tagged with 'finance' and have columns `nameOrig` and `nameDest`. The policies indicate that these columns must be redacted (masked) when data is read.

The policies can be found at `third_party/opa/opa-policy.rego`.

## Setup the notebook

Next you will create a Kubeflow notebook server and a notebook with the business logic for creating a fraud detection model.

1. Create a port-forward to communicate with Kubeflow:
    ```bash
    kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80 &
    cd samples/kubeflow/
    ```

1. Upload the notebook:
    - Open your browser and go to `http://localhost:8080`.
    - Click **Start Setup** and then **Finish** (use the `anonymous` namespace).
    - Click **Notebook Servers** (in the left).
    - In the notebooks page select in the top left the `anonymous` namespace and then click **New Server**.
    - In the notebook server creation page, set `kf-notebook` in the **Name** box and then click **Launch**. Wait for the server to become ready.
    - Click **Connect** and upload `kfM4DPolicySample.ipynb` notebook to the server.

## Run the notebook with {{< name >}}

Now you will deploy a {{< name >}} runtime environment for the notebook by creating a `M4DApplication` resource that references the `data.csv` data set that was registered in the data catalog.
This allows the code in the notebook to read the data and policies to seamlessly be applied before the data reaches the notebook server.

1. Create the `M4DApplication` resource which will deploy {{< name >}} runtime environment by running the following:
    ```bash
    cat m4dapplication.yaml | sed "s/ASSET_ID/$ASSET_ID/g" | kubectl -n anonymous apply -f -
    cd -
    ```

1. Before running the notebook you need to modify the following statements in the `Get Data` cell in the notebook:
    ```python
    ...
    client = fl.connect("grpc://<arrow-flight-module-service>.<arrow-flight-module-ns>.svc.  cluster.local:80")

    request = {
    "asset": "<bucket-name>/<file-name>.csv", 
    "columns": ["step", "type", "amount", "nameOrig", "oldbalanceOrg", "newbalanceOrig", "nameDest", "oldbalanceDest", "newbalanceDest", "isFraud", "isFlaggedFraud"]
    }
    ...
    ``` 

    - Edit the `client = fl.connect(...)` command to point to the right service and namespace of the arrow-flight-module.
    - To find the service and namespace run:
        ```bash
        kubectl get svc -l app.kubernetes.io/name=arrow-flight-module --all-namespaces
        ```

    - Edit `"asset": "<bucket-name>/<file-name>.csv"` in the second command to point to your bucket and the name of the dataset.

1. Run the notebook

    You should observe in the cell `Get Data` the data from the dataset.

1. Finally, kill the port-forward
    ```bash
    kill $!
    ```

# Next steps
You have completed an execution of a notebook with {{< name >}} and are now ready to continue exploring.
