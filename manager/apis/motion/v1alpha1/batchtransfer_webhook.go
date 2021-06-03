// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"
	log "log"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/robfig/cron"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	validationutils "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *BatchTransfer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// +kubebuilder:webhook:admissionReviewVersions=v1;v1beta1,sideEffects=None,path=/mutate-motion-m4d-ibm-com-v1alpha1-batchtransfer,mutating=true,failurePolicy=fail,groups=motion.m4d.ibm.com,resources=batchtransfers,verbs=create;update,versions=v1alpha1,name=mbatchtransfer.kb.io

var _ webhook.Defaulter = &BatchTransfer{}

const DefaultFailedJobHistoryLimit = 5
const DefaultSuccessfulJobHistoryLimit = 5

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *BatchTransfer) Default() {
	log.Printf("Defaulting batchtransfer %s", r.Name)
	if r.Spec.Image == "" {
		// TODO check if can be removed after upgrading controller-gen to 0.5.0
		r.Spec.Image = "ghcr.io/mesh-for-data/mover:latest"
	}

	if r.Spec.ImagePullPolicy == "" {
		// TODO check if can be removed after upgrading controller-gen to 0.5.0
		r.Spec.ImagePullPolicy = v1.PullIfNotPresent
	}

	if r.Spec.SecretProviderURL == "" {
		if env, b := os.LookupEnv("SECRET_PROVIDER_URL"); b {
			r.Spec.SecretProviderURL = env
		}
	}

	if r.Spec.SecretProviderRole == "" {
		if env, b := os.LookupEnv("SECRET_PROVIDER_ROLE"); b {
			r.Spec.SecretProviderRole = env
		}
	}

	if r.Spec.FailedJobHistoryLimit == 0 {
		r.Spec.FailedJobHistoryLimit = DefaultFailedJobHistoryLimit
	}

	if r.Spec.SuccessfulJobHistoryLimit == 0 {
		r.Spec.SuccessfulJobHistoryLimit = DefaultSuccessfulJobHistoryLimit
	}

	if r.Spec.Spark != nil {
		if r.Spec.Spark.Image == "" {
			r.Spec.Spark.Image = r.Spec.Image
		}

		if r.Spec.Spark.ImagePullPolicy == "" {
			r.Spec.Spark.ImagePullPolicy = r.Spec.ImagePullPolicy
		}
	}

	if env, b := os.LookupEnv("NO_FINALIZER"); b {
		if parsedBool, err := strconv.ParseBool(env); err != nil {
			panic(fmt.Sprintf("Cannot parse boolean value %s: %s", env, err.Error()))
		} else {
			r.Spec.NoFinalizer = parsedBool
		}
	}

	defaultDataStoreDescription(&r.Spec.Source)
	defaultDataStoreDescription(&r.Spec.Destination)

	if r.Spec.WriteOperation == "" {
		r.Spec.WriteOperation = Overwrite
	}

	if r.Spec.DataFlowType == "" {
		r.Spec.DataFlowType = Batch
	}

	if r.Spec.ReadDataType == "" {
		r.Spec.ReadDataType = LogData
	}

	if r.Spec.WriteDataType == "" {
		r.Spec.WriteDataType = LogData
	}
}

func defaultDataStoreDescription(dataStore *DataStore) {
	if len(dataStore.Description) == 0 {
		switch {
		case dataStore.Database != nil:
			dataStore.Description = dataStore.Database.Db2URL + "/" + dataStore.Database.Table
		case dataStore.Kafka != nil:
			dataStore.Description = "kafka://" + dataStore.Kafka.KafkaTopic
		case dataStore.S3 != nil:
			dataStore.Description = "s3://" + dataStore.S3.Bucket + "/" + dataStore.S3.ObjectKey
		case dataStore.Cloudant != nil:
			dataStore.Description = "cloudant://" + dataStore.Cloudant.Host + "/" + dataStore.Cloudant.Database
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,admissionReviewVersions=v1;v1beta1,sideEffects=None,path=/validate-motion-m4d-ibm-com-v1alpha1-batchtransfer,mutating=false,failurePolicy=fail,groups=motion.m4d.ibm.com,resources=batchtransfers,versions=v1alpha1,name=vbatchtransfer.kb.io

var _ webhook.Validator = &BatchTransfer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *BatchTransfer) ValidateCreate() error {
	log.Printf("Validating batchtransfer %s for creation", r.Name)
	return r.validateBatchTransfer()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *BatchTransfer) ValidateUpdate(old runtime.Object) error {
	log.Printf("Validating batchtransfer %s for update", r.Name)

	return r.validateBatchTransfer()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *BatchTransfer) ValidateDelete() error {
	log.Printf("Validating batchtransfer %s for deletion", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *BatchTransfer) validateBatchTransfer() error {
	var allErrs field.ErrorList
	specField := field.NewPath("spec")
	if err := r.validateBatchTransferSpec(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := validateDataStore(specField.Child("source"), &r.Spec.Source); err != nil {
		allErrs = append(allErrs, err...)
	}
	if err := validateDataStore(specField.Child("destination"), &r.Spec.Destination); err != nil {
		allErrs = append(allErrs, err...)
	}
	if r.Spec.SuccessfulJobHistoryLimit < 0 || r.Spec.SuccessfulJobHistoryLimit > 20 {
		allErrs = append(allErrs, field.Invalid(specField.Child("successfulJobHistoryLimit"),
			r.Spec.SuccessfulJobHistoryLimit, "'successfulJobHistoryLimit' has to be between 0 and 20!"))
	}
	if r.Spec.FailedJobHistoryLimit < 0 || r.Spec.FailedJobHistoryLimit > 20 {
		allErrs = append(allErrs, field.Invalid(specField.Child("failedJobHistoryLimit"),
			r.Spec.FailedJobHistoryLimit, "'failedJobHistoryLimit' has to be between 0 and 20!"))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "motion.m4d.ibm.com", Kind: "BatchTransfer"},
		r.Name, allErrs)
}

func validateDataStore(path *field.Path, store *DataStore) []*field.Error {
	var allErrs []*field.Error

	if store.Database != nil {
		var db = store.Database
		databasePath := path.Child("database")
		if len(db.Password) != 0 && db.Vault != nil {
			allErrs = append(allErrs, field.Invalid(databasePath, db.Vault, "Can only set vault or password!"))
		}

		match, _ := regexp.MatchString("^jdbc:[a-z0-9]+://", db.Db2URL)
		if !match {
			allErrs = append(allErrs, field.Invalid(databasePath.Child("db2URL"), db.Db2URL, "Invalid JDBC string!"))
		}
		r := regexp.MustCompile("^jdbc:[a-z0-9]+://([a-z0-9.-]+):")
		host := r.FindStringSubmatch(db.Db2URL)[1]

		if msgs := validationutils.IsDNS1123Subdomain(host); len(msgs) != 0 {
			allErrs = append(allErrs, field.Invalid(databasePath.Child("db2URL"), db.Db2URL, "Invalid database host!"))
		}

		if len(db.Table) == 0 {
			allErrs = append(allErrs, field.Invalid(path, db.Table, "Table cannot be empty!"))
		}
	}

	if store.S3 != nil {
		s3Path := path.Child("s3")
		_, err := url.Parse(store.S3.Endpoint)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(s3Path.Child("endpoint"), store.S3.Endpoint, "Invalid endpoint! Expecting a endpoint URL!"))
		}

		if len(store.S3.Bucket) == 0 {
			allErrs = append(allErrs, field.Invalid(s3Path.Child("bucket"), store.S3.Bucket, validationutils.EmptyError()))
		}

		if len(store.S3.ObjectKey) == 0 {
			allErrs = append(allErrs, field.Invalid(s3Path.Child("objectKey"), store.S3.ObjectKey, validationutils.EmptyError()))
		}

		if (len(store.S3.AccessKey) != 0 || len(store.S3.SecretKey) != 0) && store.S3.Vault != nil {
			allErrs = append(allErrs, field.Invalid(s3Path, store.S3.Vault, "Can only set vault or accessKey/secretKey!"))
		}
	}

	if store.Kafka != nil {
		kafkaPath := path.Child("kafka")

		// Validate Kafka brokers
		kafkaBrokers := strings.Split(store.Kafka.KafkaBrokers, ",")
		if len(kafkaBrokers) == 0 {
			allErrs = append(allErrs, field.Invalid(kafkaPath.Child("kafkaBrokers"), store.Kafka.KafkaBrokers, "Could not parse kafka brokers!"))
		}
		for i, broker := range kafkaBrokers {
			errs := validateHostPort(broker)
			for _, err := range errs {
				errMsg := fmt.Sprintf("Invalid broker at position %d (%s) %s", i, broker, err)
				allErrs = append(allErrs, field.Invalid(kafkaPath.Child("kafkaBrokers"), store.Kafka.SchemaRegistryURL, errMsg))
			}
		}

		// Validate Kafka schema registry url
		schemaRegistryURL, err := url.Parse(store.Kafka.SchemaRegistryURL)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(kafkaPath.Child("schemaRegistryUrl"), store.Kafka.SchemaRegistryURL, "Could not parse url!"))
		}
		if schemaRegistryURL != nil {
			errs := validateHostPort(schemaRegistryURL.Host)
			for _, err := range errs {
				errMsg := "Invalid host: " + err
				allErrs = append(allErrs, field.Invalid(kafkaPath.Child("schemaRegistryUrl"), store.Kafka.SchemaRegistryURL, errMsg))
			}
		}

		// Validate Kafka topic
		if len(store.Kafka.KafkaTopic) == 0 {
			allErrs = append(allErrs, field.Invalid(kafkaPath.Child("kafkaTopic"), store.Kafka.KafkaTopic, validationutils.EmptyError()))
		}

		if len(store.Kafka.Password) != 0 && store.Kafka.Vault != nil {
			allErrs = append(allErrs, field.Invalid(kafkaPath, store.Kafka.Vault, "Can only set vault or password!"))
		}

		if store.Kafka.DataFormat != "" && store.Kafka.DataFormat != "avro" && store.Kafka.DataFormat != "json" {
			allErrs = append(allErrs, field.Invalid(kafkaPath, store.Kafka.DataFormat, "Currently only 'avro' and 'json' are supported as Kafka dataFormat!"))
		}
	}

	return allErrs
}

func (r *BatchTransfer) validateBatchTransferSpec() *field.Error {
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.
	if len(r.Spec.Schedule) > 0 {
		return validateScheduleFormat(
			r.Spec.Schedule,
			field.NewPath("spec").Child("schedule"))
	}

	if r.Spec.DataFlowType == Stream {
		return field.Invalid(field.NewPath("spec").Child("dataFlowType"), r.Spec.DataFlowType, "'dataFlowType' must be 'Batch' for a BatchTransfer!")
	}

	return nil
}

func validateScheduleFormat(schedule string, fldPath *field.Path) *field.Error {
	if _, err := cron.ParseStandard(schedule); err != nil {
		return field.Invalid(fldPath, schedule, err.Error())
	}
	return nil
}

// Validates a host port combination e.g localhost:8080
// Validates if the host is a valid domain and if the port is a valid port number
func validateHostPort(hostPort string) []string {
	var errs []string
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		errs = append(errs, err.Error())
	}

	if msgs := validationutils.IsDNS1123Subdomain(host); len(msgs) != 0 {
		errs = append(errs, msgs...)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		errs = append(errs, fmt.Sprintf("Could not parse port %s as an integer", portStr))
	}
	if msgs := validationutils.IsValidPortNum(port); len(msgs) != 0 {
		for _, err := range msgs {
			errs = append(errs, "Port "+err)
		}
	}

	return errs
}
