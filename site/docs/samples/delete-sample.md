# Sample for the delete flow

This sample demonstrate how to delete an S3 object from a bucket.

## Install module

To apply the latest development version of arrow-flight-module:
```bash
kubectl apply -f -n fybrik-system https://raw.githubusercontent.com/fybrik/delete-module/main/module.yaml
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

Register the credentials required for accessing the dataset as a kubernetes secret. Replace the values for `access_key` and `secret_key` with the values from the object storage service that you used and run:

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

Then, register the data asset itself in the data catalog `katalog` used for samples. Replace the values for `endpoint`, `bucket` and `object_key` with values from the object storage service that you used and run:

```yaml
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
      finance: true
EOF
```

The asset is now registered in the catalog. The identifier of the asset is `fybrik-notebook-sample/paysim-csv` (i.e. `<namespace>/<name>`). You will use that name in the `FybrikApplication` later.

Notice the `metadata` field above. It specifies the dataset geography and tags. These attributes can later be used in policies.

For example, in the yaml above, the `geography` is set to `theshire`, you need make sure it is same with the region of your fybrik control plane, you can get the information with the below command:

```shell
kubectl get configmap cluster-metadata -n fybrik-system -o 'jsonpath={.data.Region}'
```

[Quick Start](../get-started/quickstart.md) installs a fybrik control plane with the region `theshire` by default. If you change it or the `geography` in the yaml above, a [copy module](https://github.com/fybrik/mover) will be required by the policies, but we do not install any copy module in the [Quick Start](../get-started/quickstart.md).



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
    - dataSetID: 'fybrik-notebook-sample/paysim-csv'
      flow: delete
      requirements: {}
EOF
```

Notice that the `data` field includes a `dataSetID` that matches the asset identifier in the catalog.

Run the following command to wait until the `FybrikApplication` is ready:

```bash
while [[ $(kubectl get fybrikapplication delete-app -o 'jsonpath={.status.ready}') != "true" ]]; do echo "waiting for FybrikApplication" && sleep 5; done
while [[ $(kubectl get fybrikapplication delete-app -o 'jsonpath={.status.assetStates.fybrik-notebook-sample/paysim-csv.conditions[?(@.type == "Ready")].status}') != "True" ]]; do echo "waiting for fybrik-notebook-sample/paysim-csv asset" && sleep 5; done
```

## Ensure the object is deleted

Now the object should be deleted. We can check again with [AWS CLI](https://aws.amazon.com/cli/):
```
aws --endpoint-url=${ENDPOINT} s3api list-objects --bucket=${BUCKET}
```
Now you should see that the object is no longer in the list (or no list at all if the bukcet is empty).
