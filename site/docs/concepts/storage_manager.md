# Storage manager

In several use-cases Fybrik needs to allocate storage for data. One use case is implicit copy of a dataset in read scenarios made for performance, cost or governance sake. A second scenario is when a new dataset is created by the workload. In this case Fybrik allocates the storage and registers the new dataset in the data catalog. A third use case is explicit copy - i.e. the user indicates that a copy of an existing dataset should be made. As in the second use case, here too Fybrik allocates storage for the data and registers the new dataset in the data catalog.

When we say that Fybrik allocates storage, we actually mean that Fybrik allocates a portion of an existing [storage account](#storage-account) for use by the given dataset. Fybrik must be informed what storage accounts are available, and how to access them. This information is currently provided via the FybrikStorageAccount CRD.

Storage manager is a Fybrik component responsible for allocating storage in known storage accounts and for freeing previously allocated storage. Upon a successful storage allocation, the component forms a [connection](../reference/connectors-storagemanager/Models/Connection.md) object that is passed to modules that write data to this storage.


## Deployment

Storage manager runs as a container in the manager pod. A default Fybrik deployment uses its open-source implementation as a docker image specified in Fybrik [values.yaml](https://github.com/fybrik/fybrik/blob/master/charts/fybrik/values.yaml) as `storageManager.image`.


## Can I write my own storage manager implementation?

Yes. Custom implementations are required to support the interface described in the [Storage manager API documentation](../reference/connectors-storagemanager/README.md). Next, generate a custom docker image and replace the default image used by Fybrik, as described [here](#deployment).

## Storage account

An instance of [FybrikStorageAccount](../reference/crds.md#appfybrikiov1beta2) defines properties of the shared storage for a specific connection type.
Example of a storage account for S3:
```
spec:
  id: <storage account id>
  type: s3
  secretRef: <name of the secret that holds credentials to the object store>
  geography: <storage location>
  s3:
    region: <region>
    endpoint: <endpoint>
```

## What storage types are supported?

The current implementation supports `S3` and `MySQL` storage.

Storage allocation results in creating a new S3 bucket or MySQL database. When storage is de-allocated, the dataset is deleted, and the generated bucket/database is deleted. In the future, the deletion of a bucket/database will be controlled by IT configuration policies.

In the future other storage types might be supported as well. We strongly encourage contributions to extend the supported types.

## How to support a new storage type?

### Development

- Add a new connection schema to the [taxonomy](../tasks/custom-taxonomy.md).

- Support the new type according to [Storage manager API documentation](../reference/connectors-storagemanager/README.md) and create a new docker image.

When adding a new type to the existing open-source implementation, a new package should be created in `pkg/storage/impl` and imported inside `pkg/storage/handler.go`. For example, after adding support for `kafka` in `pkg/storage/impl/kafka`:

```
_ "fybrik.io/fybrik/pkg/storage/impl/kafka"
```

### Deployment

- Specify the storage manager image in Fybrik [values.yaml](https://github.com/fybrik/fybrik/blob/master/charts/fybrik/values.yaml) as `storageManager.image`.

- Ensure existence of [FybrikModule instances](../reference/crds.md#fybrikmodule) that are able to write/copy to this storage. Update the capabilities in module yamls accordingly.

Example:

```
supportedInterfaces:
        - source:
            protocol: <new storage type>
```

- Prepare FybrikStorageAccount resources with the shared storage information.

Example:

```
apiVersion: app.fybrik.io/v1beta2
kind: FybrikStorageAccount
metadata:
  name: theshire-account
spec:
  id: theshire-storage-account
  type: <new storage type>
  secretRef: theshire-credentials
  geography: theshire
  <new storage type>:
    <attribute1>: <value1>
    <attribute2>: <value2>

```

- Update [infrastructure attributes](../tasks/infrastructure.md) related to the storage accounts, e.g., cost.

- Optionally update [IT config policies](config-policies.md) to specify when the new storage can/should be selected

- Redeploy Fybrik with the changes.