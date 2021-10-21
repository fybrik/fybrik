// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"testing"

	apiv1alpha1 "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/test"
	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/flight"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ArrowRequest struct {
	Asset   string   `json:"asset,omitempty"`
	Columns []string `json:"columns,omitempty"`
}

// TODO Refactor this test to be called from within the suite
func TestS3Notebook(t *testing.T) {
	if s, ok := os.LookupEnv("VALUES_FILE"); !ok || s != "charts/fybrik/notebook-tests.values.yaml" {
		t.Skip("Only executed for notebook tests")
	}
	RegisterFailHandler(Fail)
	defer GinkgoRecover()

	// Copy data.csv file to S3
	// S3 is assumed to be exposed on localhost at port 9090
	region := "theshire"
	endpoint := "http://localhost:9090"
	bucket := "bucket1"
	key1 := "data.csv"
	filename := "../../../samples/kubeflow/data.csv"
	s3credentials := credentials.NewStaticCredentials("ak", "sk", "")

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials:      s3credentials,
		Endpoint:         &endpoint,
		Region:           &region,
		S3ForcePathStyle: aws.Bool(true),
	}))
	svc := s3.New(sess)
	object, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key1,
	})
	if err != nil { // Could not retrieve object. Assume it does not exist
		uploader := s3manager.NewUploader(sess)

		f, ferr := os.Open(filename)
		Expect(ferr).To(BeNil(), "Opening local test data file")

		// Upload the file to S3.
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key1),
			Body:   f,
		})
		Expect(err).To(BeNil(), "S3 upload")
		if result != nil {
			log.Println(fmt.Sprintf("file uploaded to, %s\n", result.Location))
		}
	} else {
		Expect(object).ToNot(BeNil())
		log.Println("Object already exists in S3!")
	}

	err = apiv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme.Scheme})
	Expect(err).To(BeNil())

	// Create Kubernetes objects for test
	// - namespace (in setup before)
	// - asset (in setup before)
	// - asset secret (in setup before)
	// - arrow flight module (in setup before)
	// - rego policy (in setup before)

	// Push read module and application

	// Module installed by setup script directly from remote arrow-flight-module repository
	// Installing application
	application := &apiv1alpha1.FybrikApplication{}
	Expect(readObjectFromFile("../../testdata/notebook/fybrikapplication.yaml", application)).ToNot(HaveOccurred())
	applicationKey := client.ObjectKeyFromObject(application)

	// Create FybrikApplication and FybrikModule
	By("Expecting application creation to succeed")
	Expect(k8sClient.Create(context.Background(), application)).Should(Succeed())

	// Ensure getting cleaned up after tests finish
	defer func() {
		application := &apiv1alpha1.FybrikApplication{ObjectMeta: metav1.ObjectMeta{Namespace: applicationKey.Namespace, Name: applicationKey.Name}}
		_ = k8sClient.Get(context.Background(), applicationKey, application)
		_ = k8sClient.Delete(context.Background(), application)
	}()

	By("Expecting application to be created")
	Eventually(func() error {
		return k8sClient.Get(context.Background(), applicationKey, application)
	}, timeout, interval).Should(Succeed())
	By("Expecting plotter to be constructed")
	Eventually(func() *apiv1alpha1.ResourceReference {
		_ = k8sClient.Get(context.Background(), applicationKey, application)
		return application.Status.Generated
	}, timeout, interval).ShouldNot(BeNil())

	// The plotter has to be created
	plotter := &apiv1alpha1.Plotter{}
	plotterObjectKey := client.ObjectKey{Namespace: application.Status.Generated.Namespace, Name: application.Status.Generated.Name}
	By("Expecting plotter to be fetchable")
	Eventually(func() error {
		return k8sClient.Get(context.Background(), plotterObjectKey, plotter)
	}, timeout, interval).Should(Succeed())

	By("Expecting application to be ready")
	Eventually(func() bool {
		err := k8sClient.Get(context.Background(), applicationKey, application)
		if err != nil {
			return false
		}
		return application.Status.Ready
	}, timeout, interval).Should(Equal(true))

	blueprintNamespace := utils.GetBlueprintNamespace()
	fmt.Printf("blueprint namespace notebook test: %s\n", blueprintNamespace)

	// Forward port of arrow flight service to local port
	svcName := strings.Replace(application.Status.AssetStates["fybrik-notebook-sample/data-csv"].Endpoint.Hostname, "."+blueprintNamespace, "", 1)
	port := application.Status.AssetStates["fybrik-notebook-sample/data-csv"].Endpoint.Port

	By("Starting kubectl port-forward for arrow-flight")
	listenPort, err := test.RunPortForward(blueprintNamespace, svcName, int(port))
	Expect(err).To(BeNil())

	// Reading data via arrow flight
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())
	flightClient, err := flight.NewFlightClient(net.JoinHostPort("localhost", listenPort), nil, opts...)
	Expect(err).To(BeNil(), "Connect to arrow-flight service")
	defer flightClient.Close()

	request := ArrowRequest{
		Asset: "fybrik-notebook-sample/data-csv",
	}

	marshal, err := json.Marshal(request)
	Expect(err).To(BeNil())

	info, err := flightClient.GetFlightInfo(context.Background(), &flight.FlightDescriptor{
		Type: flight.FlightDescriptor_CMD,
		Cmd:  marshal,
	})
	Expect(err).To(BeNil())

	stream, err := flightClient.DoGet(context.Background(), info.Endpoint[0].Ticket)
	Expect(err).To(BeNil())

	reader, err := flight.NewRecordReader(stream)
	Expect(err).To(BeNil())
	for reader.Next() {
		record := reader.Record()
		defer record.Release()
		Expect(record.ColumnName(0)).To(Equal("step"))
		Expect(record.ColumnName(1)).To(Equal("type"))
		Expect(record.ColumnName(3)).To(Equal("nameOrig"))
		column := record.Column(3) // Check out the third 4th column that should be nameOrig and redacted

		// Check that data of nameOrig column is the correct size and all records are redacted
		dt := column.Data().DataType()
		Expect(column.Data().Len()).To(Equal(100))
		Expect(dt.ID()).To(Equal(arrow.STRING))
		Expect(dt.Name()).To(Equal((&arrow.StringType{}).Name()))
		data := array.NewStringData(column.Data())
		for i := 0; i < data.Len(); i++ {
			Expect(data.Value(i)).To(Equal("XXXXX"))
		}
	}
}
