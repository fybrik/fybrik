# Sample for the delete flow

This sample demonstrate how to delete an S3 object from a bucket.

## Install module

To apply the latest development version of the delete-module:
```bash
kubectl apply -f https://raw.githubusercontent.com/fybrik/delete-module/main/module.yaml -n fybrik-system
```

## Prepare dataset

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

Before we delete the object, we make sure it's been created.
You can check with the object storage serive that you used or with [AWS CLI](https://aws.amazon.com/cli/):
```
aws --endpoint-url=${ENDPOINT} s3api list-objects --bucket=${BUCKET}
```
You should see the new created object:
```
{
    "Contents": [
        {
            "Key": "PS_20174392719_1491204439457_log.csv",
            "LastModified": "2022-06-06T07:12:16.000Z",
            "ETag": "\"9a34903326938d8c33c29f4a1170a7b1\"",
            "Size": 6551,
            "StorageClass": "STANDARD",
            "Owner": {
                "DisplayName": "webfile",
                "ID": "75aa57f09aa0c8caeab4f8c24e99d10f8e7faeebf76c078efc7c6caea54ba06a"
            }
        }
    ]
}
```

## Register the dataset in a data catalog

In this step you are performing the role of the data owner, registering his data in the data catalog and registering the credentials for accessing the data in the credential manager.

In this tutorial, we assume that OpenMetadata is used as the data catalog.

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

The asset is now registered in the catalog.

Notice the `resourceMetadata` field above. It specifies the dataset geography and tags. These attributes can later be used in policies.

For example, in the json above, the `geography` is set to `theshire`. You need make sure that it is same as the region of your fybrik control plane. You can get this information using the following command:

```shell
kubectl get configmap cluster-metadata -n fybrik-system -o 'jsonpath={.data.Region}'
```

[Quick Start](../get-started/quickstart.md) installs a fybrik control plane with the region `theshire` by default. If you change it or the `geography` in the json above, a [copy module](https://github.com/fybrik/mover) will be required by the policies, but we do not install any copy module in the [Quick Start](../get-started/quickstart.md).

## Define data access policy

Acting as the data steward, define an [OpenPolicyAgent](https://www.openpolicyagent.org/) policy.
In this sample we only specify the action taken.
Below is the policy (written in [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/#what-is-rego) language):

```rego
package dataapi.authz

rule[{}] {
  description := "allow the delete operation"
  input.action.actionType == "delete"
}
```

In this sample only the policy above is applied. Copy the policy to a file named `sample-policy.rego` and then run:

```bash
kubectl -n fybrik-system create configmap sample-policy --from-file=sample-policy.rego
kubectl -n fybrik-system label configmap sample-policy openpolicyagent.org/policy=rego
while [[ $(kubectl get cm sample-policy -n fybrik-system -o 'jsonpath={.metadata.annotations.openpolicyagent\.org/policy-status}') != '{"status":"ok"}' ]]; do echo "waiting for policy to be applied" && sleep 5; done
```

You can similarly apply a directory holding multiple rego files.

## Create a `FybrikApplication` resource

Create a [`FybrikApplication`](../reference/crds.md#fybrikapplication) resource to register the notebook workload to the control plane of Fybrik:

<!-- TODO: check if works without role field -->
```yaml
cat <<EOF | kubectl apply -f -
apiVersion: app.fybrik.io/v1beta1
kind: FybrikApplication
metadata:
  name: delete-app
  namespace: fybrik-notebook-sample
spec:
  selector:
   workloadSelector:
     matchLabels: {}
  appInfo:
    intent: Fraud Detection
    role: Security
  data:
    - dataSetID: ${CATALOGED_ASSET}
      flow: delete
      requirements: {}
EOF
```

Notice that the `data` field includes a `dataSetID` that matches the asset identifier in the catalog.

Run the following command to wait until the `FybrikApplication` is ready:

```bash
while [[ $(kubectl get fybrikapplication delete-app -o 'jsonpath={.status.ready}') != "true" ]]; do echo "waiting for FybrikApplication" && sleep 5; done
CATALOGED_ASSET_MODIFIED=$(echo $CATALOGED_ASSET | sed 's/\./\\\./g')
while [[ $(kubectl get fybrikapplication delete-app -o "jsonpath={.status.assetStates.${CATALOGED_ASSET_MODIFIED}.conditions[?(@.type == 'Ready')].status}") != "True" ]]; do echo "waiting for ${CATALOGED_ASSET} asset" && sleep 5; done
```

## Ensure the object is deleted

Now the object should be deleted. We can check again with [AWS CLI](https://aws.amazon.com/cli/):
```
aws --endpoint-url=${ENDPOINT} s3api list-objects --bucket=${BUCKET}
```
Now you should see that the object is no longer in the list (or that the bucket is empty).
