---
title: "Data Distribution Controller v1.0"
date: 2020-05-03T21:20:45+03:00
draft: false
weight: 100
---



# Overview of Requirements and Functionality

The **Data Distribution Controller** (DDC) handles the movement of data between data stores.
{{< name >}} uses the DDC to perform an action called "implicit copy", i.e. the movement
of a data set from one data store to another with possibly some unitary transform applied
to that data set. It corresponds to Step 8 in the Architecture Document (Add a link here)

Data can be copied from data store to data store in a large variety of different ways, depending on the types of the data store (e.g. COS, Relational DB) and nature of the data capture (Streamed, Snapshot).
This document defines the functionality as well as the boundary conditions of the data distribution controller.


## Goals

This document introduces fundamental concepts of the data
distribution component and describes a high-level API for invoking data distributions.
The initial focus is on structured (tabular) data. One goal of the data distribution component is to maximize congruence
across different data stores and formats by preserving not only the data content but also the structure of the data as
faithfully as possible. Fully unstructured data such (e.g. "binary content") will also be supported but that is not the focus of the initial version. Semi-structured data will be supported on a case-by-case basis.

## Non-Goals

The focus is on how to invoke data distribution and not the if and when. This document doesn't describe the control component that is required to decide whether, when and how often data should be copied across storage systems.
Neither does the data distribution perform any policy enforcement. This is done by the component that controls the data distribution system.

# High-Level Design

Before providing an outline of the API functionality, some fundamental concepts are defined.

## Data Sets and Data Assets

The following definition is aligned with the terminology used in the Watson Knowledge Catalog.
Data is organized into data sets and data assets. A data set is a collection of data assets that is administered by a single body using a set of policies. Both data sets and data assets are uniquely identifiable. A data set is a collection of data assets with the same structure. Some examples:

- A data set is data that resides in a relational database where the database tables or views form the data asset.
- A data set consisting of objects that reside in a COS bucket where object prefix paths that have a common format are data assets. E.g. a set of partitioned parquet files with the same schema.
- A data set may be formed by a set of Kafka topics where each topic contains messages in compatible format. A data asset is represented by the content of the topic.

The unit of data distribution is the data asset.

## Data Stores

A data store allows access to data sets and data assets. Each store allows to individually access data through a data store specific API, e.g. S3 API or JDBC. Additional properties that are relevant for data distribution:

* Granularity of data access for reading: Some systems provide access to entire data assets only. (e.g. single unpartitioned files on COS). Other storage systems support queries to retrieve a sub-set (a selection and/or projection) of an individual data assets. (e.g. queries on Db2 or partitioned/bucketed prefixes on COS)

* Granularity of data access for writing: Fine-granular write access is required to apply delta-updates of individual data assets, i.e. update and insert (_upsert_) operations as well as deletes on record level are needed to process streams of changes.  Systems that support fine-granular updates are relational database systems, elastic search indexes, and in-memory caches. Other systems such as traditional parquet files stored on COS or HDFS only allow data assets to be updated in their entirety. More sophisticated storage formats such as [Delta Lake](https://delta.io/), [Apache Iceberg](https://iceberg.apache.org/) or [Hudi](https://hudi.incubator.apache.org/) extend the capabilities of parquet.

* Fidelity of the type system: Data stores use various different typing systems and have different data models that require type conversions as data is distributed between these systems.  For example, when moving the content of an elastic search index
into a relational database we are moving between two entirely different data models.
In order to minimize loss of information, type specific metadata (technical metadata) may need to be preserved as separate entities. In addition, schema inference might be needed to support certain data distributions.

The invoker of the DDC is assumed to have knowledge of the technical metadata present
at the source data asset and of the *desired* technical metadata of that data asset at the target.
If the invoker does not specify this the DDC will attempt to infer it where possible.
In both cases the source and target technical metadata are returned as part of the result of the data
distribution. If the passed source or target technical metadata is inconsistent with the data asset
at the source, then the data distribution fails.

The version 1.0 of the DDC supports the following data stores:

- Db2 LUW v10.5 or newer
- Apache Kafka v2.2 or newer (Raw + Confluent KTable format serialized in JSON or Avro)
- IBM COS with Parquet, JSON and ORC (using a Stocator based approach)


## Transformations

The data distribution supports simple transformations and filtering such as:
- Filtering of individual rows or columns based on condition clauses.
- Masking of specific columns.
- Encrypting/hashing of specific columns.
- Sampling of a subset of rows.

This is specifically for creating a derived version of a specific data asset and is **NOT** to enrich or combine
data assets, i.e. this is a not a general purpose computation environment.

## Data Life-cycle

The DDC moves a data asset from a source to a target data store. The copy of the data asset will
be retained at the target until explicitly removed by the invoker via the DDC API.


# API High-level Description

The API follows the custom resource definition approach (CRD) for Kubernetes. The following basic CRD types exist:
- _BatchTransfer_: One-time or periodic transfer of a single data asset from a source data store to a destination data store. This is also called snapshotting. This is similar to a job in K8s and will inherit many features from it, e.g. the
state is kept in K8s after the batch transfer has completed and must be deleted manually.
- _SyncTransfer_: Continuous synchronization of a single data asset from a source data store to a destination data store. The main use-case is to continuously update a destination data asset as it
is typically used in a streaming or change-data-capture scenario. This CRD is similar to a stateful set in K8s.


Both transfer types will have the same API concerning the core transfer definitions such as:
- The source data store including connection details and data asset.
- The path (in Vault) to the credentials required to access the source data store.
- The destination data store including connection details and data asset.
- The path (in Vault) to the credentials required to access the destination data store.
- Transfer properties that define parameters such as schedule, retries, transformations etc.

The difference is that _SyncTransfer_ is running continuously, _BatchTransfer_ requires a schedule or is a one-time transfer.

Initially we will limit _SyncTransfer_ to the movement of data from Kafka to COS or from Kafka to Db2.

The status of the CRD is continuously updated with the state of the data distribution. It is used to detect both success or error situations as well as freshness. It also provides transfer statistics.

Using the status of the CRD a user may examine:
- where data assets have been moved
- when this was last successfully completed (for _BatchTransfer_s)
- statistics, i.e. how long this took, how many bytes, rows etc. were transferred
- what technical metadata about the data was used at the source/destination

Other K8s controllers can watch the objects and subscribe to statistics or technical metadata updates and forward these changes e.g. in dashboards or WKT.

## Secret management

The data distribution API should not define any secrets in the CRD spec in a production environment. For development and
testing direct definitions can be used but in a production environment credentials shall be retrieved from the [secret provider](credentials_release.md).

The secret provider can be accessed via a REST API using a role and a secret name. This secret name refers to a path in vault.
At the movement operator shall not create any secrets in Kubernetes that contain any credentials and credentials shall only be maintained
in memory. The fetching of secrets will be executed by the [datamover]({{< mover_github_url >}}) component.
The datamover component retrieves configuration from a JSON file that is passed on as a Kubernetes secret. 
The goal is that vault paths can be specified in this JSON configuration file and will be substituted by values retrieved from the
secret provider. The following example illustrates this mechanism:

Given the example configuration file:
```
{
  "db2URL": "jdbc:db2//host1:1234/MYDB",
  "user": "myuser"
  "vaultPath": "/myvault/db2Password"
}
```
 and the following string in vault:
```{"password": "mypassword"}```

The substitution in the datamover will find a JSON field called `vaultPath` and look up the value using the secret
provider. The substitution happens at the same level as the `vaultPath` field was found. This works whenever
the data that is stored in vault is a JSON object itself. The advantage is that the in-memory configuration will be the same
as in a dev/test environment after the substitution. The result of the given example after substitution will be:
```
{
  "db2URL": "jdbc:db2//host1:1234/MYDB",
  "user": "myuser"
  "password": "mypassword"
}
```

This credential substitution can also be used in the `options` field of transformations.


## Error handling

The data distribution API is using validation hooks to do simple checks when a CRD is created or updated. This is a first
kind of error that will result in an error when creating/updating the CRD. It will specify an error message about which
fields are not valid. (e.g. an invalid cron pattern for the schedule property) 
As validation errors are checked before objects are created they return an error via the Kubernetes API.

If an error occurred during a `BatchTransfer` the status of the CRD will be set to `FAILED` and a possible error reason
will show in the `error` field. The error messages will differ depending on the type of exception that is thrown in the internal
datamover process. 
The internal datamover process will communicate errors to Kubernetes via a [termination message](https://kubernetes.io/docs/tasks/debug-application-cluster/determine-reason-pod-failure/). 
The content of the termination message will be written into the `error` field of the `BatchTransfer`. 
The error message shall describe the error as good as possible without any stack traces to keep it readable and displayable in a short form.

Actions for possible error states:
* Pending - Nothing to do. Normal process
* Running - Nothing to do. Normal process
* Succeeded - Possibly execute on succeeded actions (e.g. updating a catalog, ...)
* Failed - Operator will try to recover. 
* Fatal - Operator could not recover. Possibly recreate CRD to resolve and investigate error further.


## Events

In addition to errors the datamover application that is called by the data distribution api will publish Kubernetes
events for the CRD in order to give feedback for errors and successes. Errors will contain the error message.
Successful messages will contain additional metrics such as number of transferred rows or technical metadata information.

# API Specification

The formalism to use to describe this is to be decided, possibilities are Go using kubebuilder OR CRD directly. As the definition of transfer specific parameters is the same for _BatchTransfer_ kind and _SyncTransfer_ kind the definition below focusses on the _BatchTransfer_ kind. (Think of it like a pod template definition that is the same for a job or a deployment)

A possible but not complete list of Go structs using kubebuilder is:

```
// BatchTransferSpec defines the desired state of BatchTransfer
type BatchTransferSpec struct {
	Source         DataStore        `json:"source"`
	Destination    DataStore        `json:"destination"`
	Transformation []Transformation `json:"transformation,omitempty"`
	Schedule string `json:"schedule,omitempty"`
	Image string `json:"image"`                                 // Has default value from webhook
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`  // Has default value from webhook
	SecretProviderURL string `json:"secretProviderURL"`         // Has default value from webhook
	SecretProviderRole string `json:"secretProviderRole"`       // Has default value from webhook
	Suspend bool `json:"suspend,omitempty"`                     // Has default value from webhook
	MaxFailedRetries int `json:"maxFailedRetries,omitempty"`    // Has default value from webhook
	SuccessfulJobHistoryLimit int `json:"successfulJobHistoryLimit,omitempty"` // Has default value from webhook
	FailedJobHistoryLimit int `json:"failedJobHistoryLimit,omitempty"` // Has default value from webhook
}

type DataStore struct {
	DataAsset string    `json:"dataAsset"`
	Database  *Database `json:"database,omitempty"`
	S3        *S3       `json:"s3,omitempty"`
	Kafka     *Kafka    `json:"kafka,omitempty"`
}

type Database struct {
	Db2URL   string `json:"db2URL"`
	User     string `json:"user"`
	Password *string `json:"password,omitempty"`   // Please use for dev/test only!
	VaultPath *string `json:"vaultPath,omitempty"`
}

type S3 struct {
	Endpoint   string `json:"endpoint"`
	Region     string `json:"region,omitempty"`
	Bucket     string `json:"bucket"`
	AccessKey *string `json:"accessKey,omitempty"` // Please use for dev/test only!
	SecretKey *string `json:"secretKey,omitempty"` // Please use for dev/test only!
	VaultPath *string `json:"vaultPath,omitempty"`
	ObjectKey  string `json:"objectKey"`
	DataFormat string `json:"dataFormat,omitempty"`
}

type Kafka struct {
	KafkaBrokers          string `json:"kafkaBrokers"`
	SchemaRegistryURL     string `json:"schemaRegistryURL"`
	User                  string `json:"user"`
	Password             *string `json:"password,omitempty"` // Please use for dev/test only!
	VaultPath            *string `json:"vaultPath,omitempty"`
	SslTruststoreLocation string `json:"sslTruststoreLocation,omitempty"`
	SslTruststorePassword string `json:"sslTruststorePassword,omitempty"`
	KafkaTopic            string `json:"kafkaTopic"`
	CreateSnapshot        bool   `json:"createSnapshot,omitempty"`
}

type Transformation struct {
	Name string `json:"name,omitempty"`
	Action Action `json:"action,omitempty"`
	Columns []string `json:"columns,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

type Action string

const (
	RemoveColumn Action = "RemoveColumn"
	Filter       Action = "Filter"
	Encrypt      Action = "Encrypt"
	Sample       Action = "Sample"
	Digest       Action = "Digest" // md5, sha1, crc32, sha256, sha512, xxhash32, xxhash64, murmur32
	Redact       Action = "Redact" // random, fixed, formatted, etc
)

// BatchTransferStatus defines the observed state of BatchTransfer
type BatchTransferStatus struct {
	Active *corev1.ObjectReference `json:"active,omitempty"`
	Status Status `json:"status,omitempty"`
	Error string `json:"status,omitempty"`
	LastCompleted *corev1.ObjectReference `json:"lastCompleted,omitempty"`
	LastFailed *corev1.ObjectReference `json:"lastFailed,omitempty"`
	LastSuccessTime *metav1.Time `json:"lastSuccessTime,omitempty"`
	LastRecordTime *metav1.Time `json:"lastRecordTime,omitempty"`
	NumRecords int64 `json:"numRecords,omitempty"`
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
}

// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Fatal;ConfigurationError
type Status string

const (
	Pending            Status = "Pending" // Starting up transfers
	Running            Status = "Running" // Transfers are running
	Succeeded          Status = "Succeeded" // Transfers succeeded
	Failed             Status = "Failed" // Transfers failed (Maybe recoverable (e.g. temporary connection issues))
	Fatal              Status = "Fatal" // Fatal. Cannot recover. Manual intervention needed
)

// +kubebuilder:object:root=true

// BatchTransfer is the Schema for the batchtransfers API
type BatchTransfer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BatchTransferSpec   `json:"spec,omitempty"`
	Status BatchTransferStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BatchTransferList contains a list of BatchTransfer
type BatchTransferList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Assets           []BatchTransfer `json:"assets"`
}
```

## Examples

```
---
apiVersion: "m4d.ibm.com/v1"
kind: BatchTransfer
metadata:
  name: copy1
  namespace: myNamespace
spec:
  source:
    db:
      db2URL: "jdbc:db2://db1.psvc-dev.zc2.ibm.com:50602/DHUBS:sslConnection=true;"
      user: myuser
      password: "mypassword"
  destination:
    cos:
      endpoint: s3...
      bucket: myBucket
      accessKey: 0123
      secretKey: 0123
  transformation:
  - name: "Remove column A"
    action: "RemoveColumn"
    columns: ["A"]
  - name: "Digest column B"
    action: "Digest"
    columns: ["B"]
    options:
      algo: "md5"
  schedule: null # Cron schedule definition if needed
  maxFailedRetries: 3 # Maximum retries if failed
  suspend: false
  successfulJobsHistoryLimit: 2
  failedJobsHistoryLimit: 5
status:
  lastCompleted: corev1.ObjectReference # Reference to child K8s objects
  lastScheduledTime: 2018-01-01T00:00:00Z
  lastSuccessTime: 2018-01-01T00:00:00Z
  lastRecordTime: 2018-01-01T00:00:00Z # inspect data?
  numRecords: 23113
```

# External Dependencies

Data distribution will be implemented in different ways, depending on the distribution kind, on the source and destination data store technologies as well as depending on the requested transformations.

The control layer of the data distribution is implemented following the operator pattern of Kubernetes. In addition, the following technologies are relevant for specific distribution scenarios:
- Redhat Debezium for Change Data Capture
- IBM Event Streams (Apache Kafka) for SyncTransfer
- Apache Spark
- Db2 client
- COS client
- Reference to IBM Specific JDBC driver for streaming into a relation database.



# Relevant Code Repositories

The [data distribution core libraries]({{< mover_github_url >}}) that are Scala/Spark based

The [data distribution operator](https://{{< github_base >}}/{{< github_repo >}}) has been integrated into
the mesh for data code and is part of the manager.


# Roadmap

* Integration with Parquet Encryption + KeyProtect (As Target)
* Integration with Iceberg (As Target)
* Integration with Relational Databases (As Target)
* Integration with KTables (As Source)



