// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"github.com/mesh-for-data/mesh-for-data/manager/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// BatchTransferSpec defines the state of a BatchTransfer.
// The state includes source/destination specification, a schedule and the means by which
// data movement is to be conducted. The means is given as a kubernetes job description.
// In addition, the state also contains a sketch of a transformation instruction.
// In future releases, the transformation description should be specified in a separate CRD.
type BatchTransferSpec struct {
	// Source data store for this batch job
	Source DataStore `json:"source"`

	// Destination data store for this batch job
	Destination DataStore `json:"destination"`

	// Transformations to be applied to the source data before writing to destination
	Transformation []Transformation `json:"transformation,omitempty"`

	// Optional Spark configuration for tuning
	// +optional
	Spark *Spark `json:"spark,omitempty"`

	// Cron schedule if this BatchTransfer job should run on a regular schedule.
	// Values are specified like cron job schedules.
	// A good translation to human language can be found here https://crontab.guru/
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// Image that should be used for the actual batch job. This is usually a datamover
	// image. This property will be defaulted by the webhook if not set.
	// +optional
	Image string `json:"image"`

	// Image pull policy that should be used for the actual job.
	// This property will be defaulted by the webhook if not set.
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`

	// Secret provider url that should be used for the actual job.
	// This property will be defaulted by the webhook if not set.
	// +optional
	SecretProviderURL string `json:"secretProviderURL,omitempty"`

	// Secret provider role that should be used for the actual job.
	// This property will be defaulted by the webhook if not set.
	// +optional
	SecretProviderRole string `json:"secretProviderRole,omitempty"`

	// If this batch job instance is run on a schedule the regular schedule can be suspended with this property.
	// This property will be defaulted by the webhook if not set.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// Maximal number of failed retries until the batch job should stop trying.
	// This property will be defaulted by the webhook if not set.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	MaxFailedRetries int `json:"maxFailedRetries,omitempty"`

	// Maximal number of successful Kubernetes job objects that should be kept.
	// This property will be defaulted by the webhook if not set.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=20
	SuccessfulJobHistoryLimit int `json:"successfulJobHistoryLimit,omitempty"`

	// Maximal number of failed Kubernetes job objects that should be kept.
	// This property will be defaulted by the webhook if not set.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=20
	FailedJobHistoryLimit int `json:"failedJobHistoryLimit,omitempty"`

	// If this batch job instance should have a finalizer or not.
	// This property will be defaulted by the webhook if not set.
	// +optional
	NoFinalizer bool `json:"noFinalizer,omitempty"`

	// Data flow type that specifies if this is a stream or a batch workflow
	// +optional
	DataFlowType DataFlowType `json:"flowType,omitempty"`

	// Data type of the data that is read from source (log data or change data)
	// +optional
	ReadDataType DataType `json:"readDataType,omitempty"`

	// Data type of how the data should be written to the target (log data or change data)
	// +optional
	WriteDataType DataType `json:"writeDataType,omitempty"`

	// Write operation that should be performed when writing (overwrite,append,update)
	// Caution: Some write operations are only available for batch and some only for stream.
	// +optional
	WriteOperation WriteOperation `json:"writeOperation,omitempty"`
}

// A datastore has a name and can be one of the following objects.
// Note that the objects are pointers which allows for a fast check
// for the presence of a specific data store type.
// The validator makes sure that exactly one datastore definition is
// given.
type DataStore struct {
	// Description of the transfer in human readable form that is displayed in the kubectl get
	// If not provided this will be filled in depending on the datastore that is specified.
	// +optional
	Description string `json:"description,omitempty"`

	// Database data store. For the moment only Db2 is supported.
	// +optional
	Database *Database `json:"database,omitempty"`

	// An object store data store that is compatible with S3.
	// This can be a COS bucket.
	// +optional
	S3 *S3 `json:"s3,omitempty"`

	// Kafka data store. The supposed format within the given Kafka topic
	// is a Confluent compatible format stored as Avro.
	// A schema registry needs to be specified as well.
	// +optional
	Kafka *Kafka `json:"kafka,omitempty"`

	// IBM Cloudant. Needs cloudant legacy credentials.
	// +optional
	Cloudant *Cloudant `json:"cloudant,omitempty"`
}

// A minimalistic database connection definition.
// Will be extended as needed.
type Database struct {
	// URL to Db2 instance in JDBC format
	// Supported SSL certificates are currently certificates signed with IBM Intermediate CA
	// or cloud signed certificates.
	Db2URL string `json:"db2URL"`

	// Table to be read
	Table string `json:"table"`

	// Database user. Can be retrieved from vault if specified in vault parameter and is thus optional.
	// +optional
	User string `json:"user,omitempty"`

	// Database password. Can be retrieved from vault if specified in vault parameter and is thus optional.
	// +optional
	Password string `json:"password,omitempty"`

	// Define a secret import definition.
	// +optional
	SecretImport *string `json:"secretImport,omitempty"`

	// Define secrets that are fetched from a Vault instance
	// +optional
	Vault *v1alpha1.Vault `json:"vault,omitempty"`
}

// A minimalistic database connection definition.
// Will be extended as needed.
type Cloudant struct {
	// Host of cloudant instance
	Host string `json:"host"`

	// Database to be read from/written to
	Database string `json:"database"`

	// Cloudant user. Can be retrieved from vault if specified in vault parameter and is thus optional.
	// +optional
	User string `json:"username,omitempty"`

	// Cloudant password. Can be retrieved from vault if specified in vault parameter and is thus optional.
	// +optional
	Password string `json:"password,omitempty"`

	// Define a secret import definition.
	// +optional
	SecretImport *string `json:"secretImport,omitempty"`

	// Define secrets that are fetched from a Vault instance
	// +optional
	Vault *v1alpha1.Vault `json:"vault,omitempty"`
}

// An S3/COS endpoint. Besides the mandatory parameters such as
// endpoint, region, bucket, access- & secret key, the object key
// allows to define a prefix for the COS objects.
// The dataformat specifies the object format such as Parquet or ORC.
type S3 struct {
	// Endpoint of S3 service
	Endpoint string `json:"endpoint"`

	// Region of S3 service
	// +optional
	Region string `json:"region,omitempty"`

	// Bucket of S3 service
	Bucket string `json:"bucket"`

	// Access key of the HMAC credentials that can access the given bucket.
	// Can be retrieved from vault if specified in vault parameter and is thus optional.
	// +optional
	AccessKey string `json:"accessKey,omitempty"`

	// Secret key of the HMAC credentials that can access the given bucket.
	// Can be retrieved from vault if specified in vault parameter and is thus optional.
	// +optional
	SecretKey string `json:"secretKey,omitempty"`

	// Object key of the object in S3. This is used as a prefix!
	// Thus all objects that have the given objectKey as prefix will be used as input!
	ObjectKey string `json:"objectKey"`

	// Data format of the objects in S3. e.g. parquet or csv. Please refer to struct for allowed values.
	DataFormat string `json:"dataFormat,omitempty"`

	// Partition by partition (for target data stores)
	// Defines the columns to partition the output by for a target data store.
	// +optional
	PartitionBy *[]string `json:"partitionBy,omitempty"`

	// Define a secret import definition.
	// +optional
	SecretImport *string `json:"secretImport,omitempty"`

	// Define secrets that are fetched from a Vault instance
	// +optional
	Vault *v1alpha1.Vault `json:"vault,omitempty"`
}

// An extended kafka endpoint for storing KTables that also
// includes the schema registry.
type Kafka struct {
	// Kafka broker URLs as a comma separated list.
	KafkaBrokers string `json:"kafkaBrokers"`

	// URL to the schema registry. The registry has to be Confluent schema registry compatible.
	SchemaRegistryURL string `json:"schemaRegistryURL"`

	// Kafka security protocol one of (PLAINTEXT, SASL_PLAINTEXT, SASL_SSL, SSL)
	// Default SASL_SSL will be assumed if not specified
	// +optional
	SecurityProtocol string `json:"securityProtocol,omitempty"`

	// SASL Mechanism to be used (e.g. PLAIN or SCRAM-SHA-512)
	// Default SCRAM-SHA-512 will be assumed if not specified
	// +optional
	SaslMechanism string `json:"saslMechanism,omitempty"`

	// Kafka user name.
	// Can be retrieved from vault if specified in vault parameter and is thus optional.
	// +optional
	User string `json:"user,omitempty"`

	// Kafka user password
	// Can be retrieved from vault if specified in vault parameter and is thus optional.
	// +optional
	Password string `json:"password,omitempty"`

	// A truststore or certificate encoded as base64.
	// The format can be JKS or PKCS12.
	// A truststore can be specified like this or in a predefined Kubernetes secret
	// +optional
	SslTruststore string `json:"sslTruststore,omitempty"`

	// Kubernetes secret that contains the SSL truststore.
	// The format can be JKS or PKCS12.
	// A truststore can be specified like this or as
	// +optional
	SslTruststoreSecret string `json:"sslTruststoreSecret,omitempty"`

	// SSL truststore location.
	// +optional
	SslTruststoreLocation string `json:"sslTruststoreLocation,omitempty"`

	// SSL truststore password.
	// +optional
	SslTruststorePassword string `json:"sslTruststorePassword,omitempty"`

	// Kafka topic
	KafkaTopic string `json:"kafkaTopic"`

	// Deserializer to be used for the keys of the topic
	// +optional
	KeyDeserializer string `json:"keyDeserializer,omitempty"`

	// Deserializer to be used for the values of the topic
	// +optional
	ValueDeserializer string `json:"valueDeserializer,omitempty"`

	// If a snapshot should be created of the topic.
	// Records in Kafka are stored as key-value pairs. Updates/Deletes for the same key are appended
	// to the Kafka topic and the last value for a given key is the valid key in a Snapshot.
	// When this property is true only the last value will be written. If the property is false all values
	// will be written out.
	// As a CDC example:
	// If the property is true a valid snapshot of the log stream will be created.
	// If the property is false the CDC stream will be dumped as is like a change log.
	// +optional
	CreateSnapshot bool `json:"createSnapshot,omitempty"`

	// Data format of the objects in S3. e.g. parquet or csv. Please refer to struct for allowed values.
	// +optional
	DataFormat string `json:"dataFormat,omitempty"`

	// Define a secret import definition.
	// +optional
	SecretImport *string `json:"secretImport,omitempty"`

	// Define secrets that are fetched from a Vault instance
	// +optional
	Vault *v1alpha1.Vault `json:"vault,omitempty"`
}

// to be refined...
type Transformation struct {
	// Name of the transaction. Mainly used for debugging and lineage tracking.
	Name string `json:"name,omitempty"`

	// Transformation action that should be performed.
	Action Action `json:"action,omitempty"`

	// Columns that are involved in this action. This property is optional as for some actions
	// no columns have to be specified. E.g. filter is a row based transformation.
	// +optional
	Columns []string `json:"columns,omitempty"`

	// Additional options for this transformation.
	// +optional
	//+kubebuilder:pruning:PreserveUnknownFields
	Options map[string]string `json:"options,omitempty"`
}

// +kubebuilder:validation:Enum=RemoveColumns;EncryptColumns;DigestColumns;RedactColumns;SampleRows;FilterRows
type Action string

// to be refined...
const (
	RemoveColumns  Action = "RemoveColumns"
	EncryptColumns Action = "EncryptColumns"
	DigestColumns  Action = "DigestColumns" // md5, sha1, crc32, sha256, sha512, xxhash32, xxhash64, murmur32
	RedactColumns  Action = "RedactColumns" // random, fixed, formatted, etc
	SampleRows     Action = "SampleRows"
	FilterRows     Action = "FilterRows"
)

// +kubebuilder:validation:Enum=Batch;Stream
type DataFlowType string

// to be refined...
const (
	Batch  DataFlowType = "Batch"
	Stream DataFlowType = "Stream"
)

// +kubebuilder:validation:Enum=LogData;ChangeData
type DataType string

// to be refined...
const (
	LogData    DataType = "LogData"
	ChangeData DataType = "ChangeData"
)

// +kubebuilder:validation:Enum=Overwrite;Append;Update
type WriteOperation string

// to be refined...
const (
	Overwrite WriteOperation = "Overwrite"
	Append    WriteOperation = "Append"
	Update    WriteOperation = "Update"
)

type Spark struct {
	// Name of the transaction. Mainly used for debugging and lineage tracking.
	// +optional
	AppName string `json:"appName,omitempty"`

	// Number of cores that the driver should use
	// +optional
	DriverCores int `json:"driverCores,omitempty"`

	// Memory that the driver should have
	// +optional
	DriverMemory int `json:"driverMemory,omitempty"`

	// Number of executors to be started
	// +optional
	NumExecutors int `json:"numExecutors,omitempty"`

	// Number of cores that each executor should have
	// +optional
	ExecutorCores int `json:"executorCores,omitempty"`

	// Memory that each executor should have
	// +optional
	ExecutorMemory string `json:"executorMemory,omitempty"`

	// Image to be used for executors
	// +optional
	Image string `json:"image,omitempty"`

	// Image pull policy to be used for executor
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`

	// Number of shuffle partitions for Spark
	// +optional
	ShufflePartitions int `json:"shufflePartitions,omitempty"`

	// Additional options for Spark configuration.
	// +optional
	//+kubebuilder:pruning:PreserveUnknownFields
	AdditionalOptions map[string]string `json:"options,omitempty"`
}

// BatchTransferStatus defines the observed state of BatchTransfer
// This includes a reference to the job that implements the movement
// as well as the last schedule time.
// What is missing: Extended status information such as:
// - number of records moved
// - technical meta-data
type BatchTransferStatus struct {
	// A pointer to the currently running job (or nil)
	// +optional
	Active *corev1.ObjectReference `json:"active,omitempty"`

	// +optional
	Status BatchStatus `json:"status,omitempty"`

	// +optional
	Error string `json:"error,omitempty"`

	// +optional
	LastCompleted *corev1.ObjectReference `json:"lastCompleted,omitempty"`

	// +optional
	LastFailed *corev1.ObjectReference `json:"lastFailed,omitempty"`

	// +optional
	LastSuccessTime *metav1.Time `json:"lastSuccessTime,omitempty"`

	// +optional
	LastRecordTime *metav1.Time `json:"lastRecordTime,omitempty"`

	// +optional
	// +kubebuilder:validation:Minimum=0
	NumRecords int64 `json:"numRecords,omitempty"`

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
}

// +kubebuilder:validation:Enum=STARTING;RUNNING;SUCCEEDED;FAILED
type BatchStatus string

// to be refined...
const (
	Starting  BatchStatus = "STARTING"
	Running   BatchStatus = "RUNNING"
	Succeeded BatchStatus = "SUCCEEDED"
	Failed    BatchStatus = "FAILED"
)

// the following to annotations are crucial as they allow to update the status of the BatchTransfer CRD
// as a sub-resource. If not provided, then the controller will fail...
// limit the scope of this CRD to a namespace. This is the default but we want to make it explicit.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Source",type=string,JSONPath=`.spec.source.description`
// +kubebuilder:printcolumn:name="Destination",type=string,JSONPath=`.spec.destination.description`
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:resource:scope=Namespaced

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
	Items           []BatchTransfer `json:"items"`
}

const (
	BatchtransferFinalizer       = "batchtransfer.finalizers.ibm.com"
	BatchtransferFinalizerBinary = "/finalizer"
	BatchtransferBinary          = "/mover"
	ConfigSecretVolumeName       = "conf-secret"
	ConfigSecretMountPath        = "/etc/mover"
)

// register above definition...
func init() {
	SchemeBuilder.Register(&BatchTransfer{}, &BatchTransferList{})
}

// +k8s:deepcopy-gen=false
type Transfer interface {
	IsBeingDeleted() bool
	HasStarted() bool
	HasFinalizer() bool
	AddFinalizer()
	RemoveFinalizer()
	FinalizerPodName() string
	FinalizerPodKey() client.ObjectKey
	ObjectKey() client.ObjectKey
	GetImage() string
	GetImagePullPolicy() corev1.PullPolicy

	runtime.Object
	metav1.Object
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (batchTransfer *BatchTransfer) IsBeingDeleted() bool {
	return !batchTransfer.ObjectMeta.DeletionTimestamp.IsZero()
}

func (batchTransfer *BatchTransfer) IsCronJob() bool {
	return batchTransfer.Spec.Schedule != ""
}

func (batchTransfer *BatchTransfer) HasStarted() bool {
	return batchTransfer.Status.Active != nil ||
		batchTransfer.Status.LastFailed != nil ||
		batchTransfer.Status.LastCompleted != nil
}

func (batchTransfer *BatchTransfer) HasFinalizer() bool {
	return controllerutil.ContainsFinalizer(batchTransfer, BatchtransferFinalizer)
}

func (batchTransfer *BatchTransfer) AddFinalizer() {
	controllerutil.AddFinalizer(batchTransfer, BatchtransferFinalizer)
}

func (batchTransfer *BatchTransfer) RemoveFinalizer() {
	controllerutil.RemoveFinalizer(batchTransfer, BatchtransferFinalizer)
}

func (batchTransfer *BatchTransfer) FinalizerPodName() string {
	return batchTransfer.Name + "-finalizer"
}

func (batchTransfer *BatchTransfer) FinalizerPodKey() client.ObjectKey {
	return client.ObjectKey{
		Namespace: batchTransfer.Namespace,
		Name:      batchTransfer.FinalizerPodName(),
	}
}

func (batchTransfer *BatchTransfer) ObjectKey() client.ObjectKey {
	return client.ObjectKey{
		Namespace: batchTransfer.Namespace,
		Name:      batchTransfer.Name,
	}
}

func (batchTransfer *BatchTransfer) GetImage() string {
	return batchTransfer.Spec.Image
}

func (batchTransfer *BatchTransfer) GetImagePullPolicy() corev1.PullPolicy {
	return batchTransfer.Spec.ImagePullPolicy
}
