# Storage Management by Fybrik

## Background
In several use-cases Fybrik needs to allocate storage for data. One use case is implicit copy of a dataset in read scenarios made for performance, cost or governance sake. A second scenario is when a new dataset is created by the workload. In this case Fybrik allocates the storage and registers the new dataset in the data catalog. A third use case is explicit copy - i.e. the user indicates that a copy of an existing dataset should be made. As in the second use case, here too Fybrik allocates storage for the data and registers the new dataset in the data catalog.

When we say that Fybrik allocates storage, we actually mean that Fybrik allocates a portion of an existing storage account for use by the given dataset. Fybrik must be informed what storage accounts are available, and how to access them. This information is currently provided via the FybrikStorageAccount CRD.  

Modules that write data to this storage, receive from Fybrik a connection that holds the relevant information about the storage (e.g. endpoint and write credentials).

Currently, only S3 storage is supported. Both allocation and deletion of the storage (if temporary) is done using [Datashim](https://datashim.io/).
Business logic related to storage management is hard-coded in Fybrik.

## Gaps / Requirements

- Support additional connection types (e.g. [MySQL](https://www.mysql.com/), [Google Sheets](https://learn.microsoft.com/en-us/connectors/googlesheet/))

- Business logic should not be hard-coded in Fybrik. 

- Storage manager should use the common connection taxonomy.

- Deployment of FybrikStorageAccount CRD should be configurable - to be discussed: [issue 1717](https://github.com/fybrik/fybrik/issues/1717) 

- Fybrik manages storage life cycle of temporary data copies.

- A clear error indication should be provided if the requested storage type is not supported, or an operation has failed.

- IT admin should be able to express config policies related to storage allocation, based on storage dynamic attributes (e.g., cost) as well as storage properties such as type, geography and others. 

- Optimizer needs to ensure that the allocated storage type matches the connection that the module uses.

- The selected storage should not necessarily match the source dataset connection in case of copying an existing asset.

- The data user should be able to leave a choice of a storage type to Fybrik and organization policies.  

- [Future enhancement] The data user should be able to request a specific storage type, or to specify some of the connection properties (e.g., bucket name) inside FybrikApplication. 


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

- Fybrik defines **connection taxonomy layer** with schema definition for supported connections (in pkg/taxonomy/layers). Quickstart deploys Fybrik using this layer. Users can modify the connection layer when deploying Fybrik.

In **Phase1** we define a connection taxonomy layer for all connection types that are supported today by open-source modules, such as `s3`, `db2`, `kafka`, `arrow-flight`. (Revisit taxonomy definition used by Airbyte module.)

See [connection taxonomy](https://github.com/fybrik/fybrik/blob/master/samples/taxonomy/example/catalog/connection.yaml) for an example of a connection taxonomy layer.

- Selection of modules is done based on connection `name`.

In **Phase2** we add an optional `Category` field to `Connection`. Category represents a wider set of connection types that can be supported by a module.
For example, `Generic S3` can represent `AWS S3`, `IBM Cloud Object Storage`, and so on.
A module yaml will be able to specify either name or category as a connection type.  Optimizer will do the matching.

### FybrikStorageAccount 
Today, FybrikStorageAccount spec defines properties of s3 storage, e.g.:
```
spec:
  id: theshire-object-store
  secretRef: credentials-theshire
  region: theshire
  endpoint: http://s3.eu.cloud-object-storage.appdomain.cloud
```

We suggest to add `type` (connection name), `geography` and the appropriate connection properties taken from the taxonomy.
Example:
```
spec:
  id: theshire-object-store
  type: s3
  secretRef: credentials-theshire
  geography: theshire
  s3:
    region: eu
    endpoint: http://s3.eu.cloud-object-storage.appdomain.cloud
```

Dynamic information about performance, amount free, costs, etc., are detailed in the separate Infrastructure Attributes JSON file.

## StorageManager

StorageManager is responsible for allocating storage in known storage accounts (as declared in FybrikStorageAccount CRDs) and for freeing the allocated storage.  



### Interfaces

StorageManager runs as a new container in the manager pod. A default Fybrik deployment uses its open-source implementation as a docker image specified in Fybrik `values.yaml`. This implementation can be replaced as long as the alternative obeys the following APIs:

#### AllocateStorage

Storage is allocated after the appropriate storage account has been selected by the optimizer. 

`AllocateStorage` request includes properties of the selected storage account, asset name (and additional properties) defined in FybrikApplication, prefix for name generation based on application uuid, attributes defined by IT config policies, e.g., bucket_name. 
Upon a successful allocation, a connection object is returned.
In case of an error, a detailed error message is returned. Examples of errors: credentials are not provided, access to cloud object storage is forbidden.

#### DeleteStorage

The allocated storage is freed after FybrikApplication is deleted or a dataset is no longer required in the spec, and the storage is not persistent.

`DeleteStorage` request receives the `Connection` object to be deleted, and configuration options defined by IT config policies, e.g., delete_empty_bucket.

It returns the operation status (success/failure), and a detailed error message, e.g. access is denied, the specified bucket does not exist, etc.

#### GetSupportedConnectionTypes

Returns a list of supported connection types. Optimizer will use this list to constrain selection of storage accounts. 


## Architecture

Storage management functionality will be defined in Fybrik repo under `pkg/storage`. 

The folder will include:

- FybrikStorageAccount types to generate the CRD
- Open-source implementation of StorageManager APIs based on a some / all of the connection types in the [taxonomy](#taxonomy-and-structures)

Architecture of StorageManager is based on [Design pattern](https://eli.thegreenplace.net/2019/design-patterns-in-gos-databasesql-package)

`agent` defines the interface for `AllocateStorage`/`DeleteStorage`:

```
package agent

type AgentInterface interface {
    func AllocateStorage...
    func DeleteStorage...
}
```

`registrator` registers `agents` implementing the interface:
```
package registrator

import "registrator/agent"

var (
	agentsMu sync.RWMutex
	agents   = make(map[string]agent.Agent)
)

func Register(name string, worker agent.Agent) error {
	agentsMu.Lock()
	defer agentsMu.Unlock()
	if worker == nil {
		// return error
	}
	if _, dup := agents[name]; dup {
		// return error
	}
	agents[name] = worker
}
```

Each agent implements the interface and registers the connection type it supports in init(). 
```
package s3Agent

import "registrator"
import "registrator/agent"

func init() {
    registrator.Register("s3", &S3Agent{})
}

func AllocateStorage...
func DeleteStorage...
```

StorageManager invokes the appropriate agent based on the registered connection type.
```
package storageManager
// import registrator
import "registrator"
// import agents
import "s3Agent"

... within AllocateStorage
// get agent by type
agent = registrator.GetAgent(type)
// invoke the agent
agent.AllocateStorage...
```

## How to support a new connection type

### Development

- Add a new connection schema and compile `taxonomy.json`.

- StorageManager: implement `AllocateStorage`/`DeleteStorage` for the new type and register it.

- Create a new docker image of StorageManager

- Optionally extend catalog-connector to support the new connection.

- Re-install Fybrik release using StorageManager image and the new taxonomy schema. No change to fybrik_crd is required.

### Deployment

- Ensure existence of modules that are able to write/copy to this connection. Update the capabilities in module yamls accordingly.

- Prepare FybrikStorageAccount resources with the shared storage information.

- Update infrastructure attributes related to the storage accounts, e.g., cost.

- Optionally update IT config policies to specify when the new storage can/should be selected

## Fybrik deployment configuration

In [values.yaml](https://github.com/fybrik/fybrik/blob/master/charts/fybrik/values.yaml) add a section `storageManager` with `image` of StorageManager. Modify manager deployment to bring up a new container running in the manager pod.


## Changes to Optimizer and storage selection

Currently, the only available storage connection type is S3. It has been hard-coded inside manager as the default connection type used in the write flow of a new dataset. Now, the following changes are required (both to optimizer and the manager naive algorithm):

- add the constraint of connection type/category matching the module protocol

- do not specify the desired connection type and determine it later from the selected storage type

## Backwards compatibility:

- FybrikStorageAccount CRD will be changed without preserving backwards compatibility

- No changes to connectors or connector APIs

- Changes to AirByte module chart are required after the connection layer is defined, no changes to other modules

## To consider in the future:

- Changes to FybrikApplication (add user requirements for storage type, connection details such as bucket name, etc. to the dataset entry)

- Use information about the amount of storage available and amount of data to be written/copied to influence storage selection.

- Extend IT config policies with options for storage management

## Development plan 

### Phase1

- Provide connection taxonomy layer for s3, db2, kafka, arrow-flight, what else?

- Changes to FybrikStorageAccount CR

- Implement StorageManager with the defined API for s3 using minio sdk.

- Remove dependency on datashim

- Update documentation accordingly

- Changes to Airbyte module to adapt the suggested taxonomy


### Phase2

- Add Category field to `Connection`, modify matching criteria in the optimizer/ non-CSP algorithm. 

- Lift the requirement for the default S3 storage, add constraints to the optimizer. Change the non-CSP algorithm as well.

- Support MySQL

### Phase3

- IT config policies for configuration options

### Phase4

- Support additional connection types - DB2, Kafka, Google Sheets, 
