# Storage Management by Fybrik

## Background
In several use-cases Fybrik needs to allocate storage for data. One use case is implicit copy of a dataset in read scenarios made for performance, cost or governance sake. A second scenario is when a new dataset is created by the workload. In this case Fybrik allocates the storage and registers the new dataset in the data catalog. A third use case is explicit copy - i.e. the user indicates that a copy of an existing dataset should be made. As in the second use case, here too Fybrik allocates storage for the data and registers the new dataset in the data catalog.

When we say that Fybrik allocates storage, we actually mean that Fybrik allocates a portion of an existing storage account for use by the given dataset. Fybrik must be informed what storage accounts are available, and how to access them. This information is currently provided via the FybrikStorageAccount CRD.  

Modules that write data to this storage, receive from Fybrik a connection that holds the relevant information about the storage (e.g. endpoint and write credentials).

Currently, only S3 storage is supported. Both allocation and deletion of the storage (if temporary) is done using Datashim.

## Gaps / Requirements

- Support additional connection types (e.g. MySQL, googlesheets)

- Business logic should not be hard-coded in Fybrik. 

- A single connection taxonomy should be used by modules, catalog conector and storage manager.

- Deployment of FybrikStorageAccount CRD should be configurable - to be discussed: [issue 1717](https://github.com/fybrik/fybrik/issues/1717) 

- Fybrik manages storage life cycle of temporary data copies.

- A clear error indication should be provided if the requested storage type is not supported, or an operation has failed.

- IT admin should be able to express config policies based on storage dynamic attributes (e.g., cost) as well as storage properties such as type, geography and others. 

- Optimizer needs to ensure that the allocated storage type matches the connection that the module uses.

- The selected storage should not necessarily match the source dataset connection in case of copying an existing asset.

- [Future enhancement] The data user should be able to request a specific storage type, or to specify some of the connection properties (e.g., bucket name) inside FybrikApplication. 

- The data user should be able to leave a choice of a storage type to Fybrik and organization policies.  

## Goals 

- Provide modules with a connection object for writing data

- Share organization storage to store temporary or persistent data while hiding the details (credentials, server URLs, service endpoints, etc.) from the data user. This means that data users will store data in organization accounts created by IT administrators.  

- Govern the use of the shared storage for the given workload according to the compliance, capacity and other factors.

- Optimize the shared storage for the given workload (by cost, latency to the workload)

- Manage storage life cycle of the shared storage (e.g., delete temporary data after not being used, delete empty buckets, etc.) 

# High Level Design

## Taxonomy and structures

It is crucial that the modules and the storage allocation component have a base common connection structure.  Otherwise, the components will not work together correctly when deployed.  Thus, we propose the following:

- **Connection** is defined in **base** taxonomy within Fybrik repository, as it is done today:

```
// +kubebuilder:pruning:PreserveUnknownFields
// Name of the connection to the data source
type Connection struct {
	// Name of the connection to the data source
	Name                 ConnectionType   `json:"name"`
	AdditionalProperties serde.Properties `json:"-"`
}
```

- Fybrik defines **taxonomy layers** with schema definition for supported connections (in pkg/taxonomy/layers). Quickstart deploys Fybrik using these layers. Users can add/replace layers when deploying Fybrik.

In **Phase1** we define layers for all connection types that are supported today by open-source modules, such as `s3`, `db2`, `kafka`, `arrow-flight`. (Revisit taxonomy layers used by Airbyte module.)

Example of a taxonomy layer for `s3`:
```
  s3:
    description: Connection information for S3 compatible object store
    type: object
    properties:
      bucket:
        type: string
      endpoint:
        type: string
      object_key:
        type: string
      region:
        type: string
    required:
    - bucket
    - endpoint
    - object_key
```
- Selection of modules is done based on connection `name`.

In **Phase2** we add an optional `Category` field to `Connection`. 
A module yaml will specify as a connection type either name or category. Optimizer will do the matching.

### FybrikStorageAccount 
Today, FybrikStorageAccount spec defines properties of s3 storage, e.g.:
```
spec:
  id: theshire-object-store
  secretRef: credentials-theshire
  region: theshire
  endpoint: http://s3.eu.cloud-object-storage.appdomain.cloud
```

We suggest to add `type` (connection name), `geography` and `properties` defining the appropriate connection properties.
Example:
```
spec:
  id: theshire-object-store
  type: s3
  secretRef: credentials-theshire
  geography: theshire
  properties:
    s3:
        region: eu
        endpoint: http://s3.eu.cloud-object-storage.appdomain.cloud
```

Dynamic information about performance, amount free, costs, etc., are detailed in the separate Infrastructure Attributes JSON file.

## StorageManager

StorageManager is responsible for allocating storage in known storage accounts (as declared in FybrikStorageAccount CRDs) and for freeing the allocated storage.  



### Architecture and interfaces

StorageManager runs as a new container in the manager pod. A default Fybrik deployment uses its open-source implementation as a docker image specified in Fybrik `values.yaml`. This implementation can be replaced as long as the alternative obeys the following APIs:

#### AllocateStorage

Storage is allocated after the appropriate storage account has been selected by the optimizer. 

`AllocateStorage` request includes properties of the selected storage account, asset name (and additional properties) defined in FybrikApplication, prefix for name generation based on application uuid, attributes defined by IT config policies, e.g., bucket_name. Upon a successful allocation, a connection object will be returned.

#### DeleteStorage

The allocated storage is freed after FybrikApplication is deleted or a dataset is no longer required in the spec, and the storage is not persistent.

`DeleteStorage` request receives the `Connection` object to be deleted, and configuration options defined by IT config policies, e.g., delete_empty_bucket.

#### GetSupportedConnectionTypes

Returns a list of supported connection types. Optimizer will use this list to constrain selection of storage accounts. 


## Architecture

As the first step, storage management functionality will be defined in Fybrik repo under `pkg/storage`. 
To consider moving to another repository in the future.

The folder will include:

- FybrikStorageAccount types to generate the CRD
- Open-source implementation of StorageManager APIs

Architecture of StorageManager is based on [Design pattern](https://eli.thegreenplace.net/2019/design-patterns-in-gos-databasesql-package)

It defines `main` that registers various connection types and `plugins` that implement the interface for `AllocateStorage`/`DeleteStorage`. Each plugin registers the connection type it supports in init(). StorageManager invokes the appropriate plugin method based on the registered connection type.


## How to support a new connection type

### StorageManager

- Add a new plugin with implementation of `AllocateStorage`/`DeleteStorage` and register it in the main process.

- Create a new docker image of StorageManager

### Fybrik core

- Add a new taxonomy layer describing the connection schema and compile `taxonomy.json`.

- Ensure existence of modules that are able to write/copy to this connection. Update the capabilities in module yamls accordingly.

- Prepare FybrikStorageAccount resources with the shared storage information.

- Update infrastructure attributes related to the storage accounts, e.g., cost.

- Optionally update IT config policies to specify when the new storage can/should be selected

- Optionally extend catalog-connector to support the new connection.

- Deploy new/modified yamls and re-install Fybrik release using StorageManager image and the new taxonomy schema. No change to fybrik_crd is required.


## Deployment configuration

In values.yaml add another section `storageManager` with `image` of StorageManager. Modify manager deployment.


## Changes to Optimizer and storage type selection

Earlier, the only available storage type was S3. It was hard-coded inside manager as the default connection type used in the write flow of a new dataset. Now, the following changes are required (both to optimizer and the manager naive algorithm):

- add the constraint of storage type/category matching the module protocol

- do not specify the desired connection type and determine it later from the selected storage type


## To consider in the future:

- Changes to FybrikApplication (add requirements to the dataset entry)

- Use information about the amount of storage available and amount of data to be written/copied to influence storage selection.

- Extend IT config policies with options for storage management

## Development plan 

### Phase1

- Provide layers for s3, db2, kafka, arrow-flight, what else?

- Changes to FybrikStorageAccount CR

- Implement StorageManager with the defined API for s3 using minio sdk.

- Remove dependency on datashim

- Update documentation accordingly

- Changes to Airbyte module to adapt the suggested taxonomy


### Phase2

- Lift the requirement for the default S3 storage, add constraints to the optimizer. Change the non-CSP algorithm as well.

### Phase3

- IT config policies for configuration options

### Phase4

- Support additional connection types - to be decided what types and what priorities. 
