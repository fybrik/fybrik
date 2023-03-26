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
	"strconv"
	"strings"
	"testing"

	"github.com/apache/arrow/go/v7/arrow"
	"github.com/apache/arrow/go/v7/arrow/array"
	"github.com/apache/arrow/go/v7/arrow/flight"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/pkg/test"
	fybrikUtils "fybrik.io/fybrik/pkg/utils"
)

const (
	readFlow      string = "charts/fybrik/notebook-test-readflow.values.yaml"
	readFlowTLS   string = "charts/fybrik/notebook-test-readflow.tls.values.yaml"
	readFlowTLSCA string = "charts/fybrik/notebook-test-readflow.tls-system-cacerts.yaml"
)

type ArrowRequest struct {
	Asset   string   `json:"asset,omitempty"`
	Columns []string `json:"columns,omitempty"`
}

func TestS3NotebookReadFlow(t *testing.T) {
	success := false
	valuesYaml, ok := os.LookupEnv("VALUES_FILE")
	if !ok || (valuesYaml != readFlow && valuesYaml != readFlowTLS && valuesYaml != readFlowTLSCA) {
		t.Skip("Only executed for notebook tests")
	}
	catalogedAsset, ok := os.LookupEnv("CATALOGED_ASSET")
	if !ok || catalogedAsset == "" {
		log.Printf("CATALOGED_ASSET should be defined.")
		t.FailNow()
	}
	gomega.RegisterFailHandler(Fail)

	g := gomega.NewGomegaWithT(t)
	defer GinkgoRecover()

	// Copy data.csv file to S3
	// S3 is assumed to be exposed on localhost at port 9090
	region := "theshire"
	endpoint := "http://localhost:9090"
	bucket := "bucket1"
	key1 := "data.csv"
	filename := "../../testdata/data.csv"
	s3credentials := credentials.NewStaticCredentials("ak", "sk", "")

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials:      s3credentials,
		Endpoint:         &endpoint,
		Region:           &region,
		S3ForcePathStyle: aws.Bool(true),
	}))
	s3Client := s3.New(sess)
	object, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key1,
	})
	if err != nil { // Could not retrieve object. Assume it does not exist
		uploader := s3manager.NewUploader(sess)

		f, ferr := os.Open(filename)
		g.Expect(ferr).To(gomega.BeNil(), "Opening local test data file")

		// Upload the file to S3.
		var result *s3manager.UploadOutput
		result, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key1),
			Body:   f,
		})
		g.Expect(err).To(gomega.BeNil(), "S3 upload")
		if result != nil {
			log.Printf("file uploaded to, %s\n", result.Location)
		}
	} else {
		g.Expect(object).ToNot(gomega.BeNil())
		log.Println("Object already exists in S3!")
	}
	err = fapp.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// creates the clientset
	// can not use controller-runtime client. see https://github.com/kubernetes-sigs/controller-runtime/issues/452
	clientset, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	g.Expect(err).NotTo(gomega.HaveOccurred())

	k8sClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme.Scheme}) //nolint:govet
	g.Expect(err).To(gomega.BeNil())

	// Create Kubernetes objects for test
	// - namespace (in setup before)
	// - asset (in setup before)
	// - asset secret (in setup before)
	// - arrow flight module (in setup before)
	// - rego policy (in setup before)

	// Push read module and application

	// Module installed by setup script directly from remote arrow-flight-module repository
	// Installing application
	application := &fapp.FybrikApplication{}
	plotter := &fapp.Plotter{}
	var applicationKey client.ObjectKey
	var plotterObjectKey client.ObjectKey
	var modulesNamespace string

	// Check allow-by-default in case the values is notebook-test-readflow.values.yaml
	if valuesYaml == readFlow {
		// Starting allow-by-default
		fmt.Println("Starting allow-by-default read")

		// Module installed by setup script directly from remote arrow-flight-module repository
		// Installing application

		g.Expect(readObjectFromFile("../../testdata/notebook/read-flow/fybrikapplication.yaml", application)).
			ToNot(gomega.HaveOccurred())
		application.ObjectMeta.Name += "-1"
		application.Spec.Data[0].DataSetID = catalogedAsset
		applicationKey = client.ObjectKeyFromObject(application)

		// Create FybrikApplication
		fmt.Println("Expecting application creation to succeed")
		g.Expect(k8sClient.Create(context.Background(), application)).Should(gomega.Succeed())

		fmt.Println("Expecting application to be created")
		g.Eventually(func() error {
			return k8sClient.Get(context.Background(), applicationKey, application)
		}, timeout, interval).Should(gomega.Succeed())
		fmt.Println("Expecting plotter to be constructed")
		g.Eventually(func() *fapp.ResourceReference {
			_ = k8sClient.Get(context.Background(), applicationKey, application)
			return application.Status.Generated
		}, timeout, interval).ShouldNot(gomega.BeNil())

		// The plotter has to be created

		plotterObjectKey = client.ObjectKey{Namespace: application.Status.Generated.Namespace,
			Name: application.Status.Generated.Name}
		fmt.Println("Expecting plotter to be fetchable")
		g.Eventually(func() error {
			return k8sClient.Get(context.Background(), plotterObjectKey, plotter)
		}, timeout, interval).Should(gomega.Succeed())

		fmt.Println("Expecting application to be ready")
		g.Eventually(func() bool {
			err = k8sClient.Get(context.Background(), applicationKey, application)
			if err != nil {
				return false
			}
			return application.Status.Ready
		}, timeout, interval).Should(gomega.Equal(true))

		modulesNamespace = plotter.Spec.ModulesNamespace
		fmt.Printf("data access module namespace notebook test: %s\n", modulesNamespace)
		g.Expect(application.Status.AssetStates[catalogedAsset].
			Conditions[ReadyConditionIndex].Status).To(gomega.Equal(v1.ConditionTrue))
		g.Expect(application.Status.AssetStates[catalogedAsset].Endpoint.Name).
			ToNot(gomega.BeEmpty())

		// cleanup of first application
		g.Eventually(func() error {
			return k8sClient.Delete(context.Background(), application)
		}, timeout, interval).Should(gomega.Succeed())

		// Deploy policy from a configmap
		piiReadConfigMap := &v1.ConfigMap{}
		// Create reduct PII policy
		g.Expect(readObjectFromFile("../../testdata/notebook/read-flow/pii-policy-cm.yaml", piiReadConfigMap)).ToNot(gomega.HaveOccurred())
		piiReadConfigMapKey := client.ObjectKeyFromObject(piiReadConfigMap)
		g.Expect(k8sClient.Create(context.Background(), piiReadConfigMap)).Should(gomega.Succeed())

		fmt.Println("Expecting configmap to be created")
		g.Eventually(func() error {
			return k8sClient.Get(context.Background(), piiReadConfigMapKey, piiReadConfigMap)
		}, timeout, interval).Should(gomega.Succeed())
		fmt.Println("Expecting policies to be compiled")
		g.Eventually(func() string {
			_ = k8sClient.Get(context.Background(), piiReadConfigMapKey, piiReadConfigMap)
			return piiReadConfigMap.Annotations["openpolicyagent.org/policy-status"]
		}, timeout, interval).Should(gomega.BeEquivalentTo("{\"status\":\"ok\"}"))

		defer func() {
			cm := &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: piiReadConfigMapKey.Namespace,
				Name: piiReadConfigMapKey.Name}}
			_ = k8sClient.Get(context.Background(), piiReadConfigMapKey, cm)
			_ = k8sClient.Delete(context.Background(), cm)
		}()
	}

	// Starting read with reduct
	fmt.Println("Starting read with reduct")

	application = &fapp.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/notebook/read-flow/fybrikapplication.yaml", application)).
		ToNot(gomega.HaveOccurred())
	application.ObjectMeta.Name += "-2"
	application.Spec.Data[0].DataSetID = catalogedAsset
	applicationKey = client.ObjectKeyFromObject(application)

	// Create FybrikApplication
	fmt.Println("Expecting application creation to succeed")
	g.Expect(k8sClient.Create(context.Background(), application)).Should(gomega.Succeed())

	fmt.Println("Expecting application to be created")
	g.Eventually(func() error {
		return k8sClient.Get(context.Background(), applicationKey, application)
	}, timeout, interval).Should(gomega.Succeed())
	fmt.Println("Expecting plotter to be constructed")
	g.Eventually(func() *fapp.ResourceReference {
		_ = k8sClient.Get(context.Background(), applicationKey, application)
		return application.Status.Generated
	}, timeout, interval).ShouldNot(gomega.BeNil())

	// The plotter has to be created
	plotter = &fapp.Plotter{}
	plotterObjectKey = client.ObjectKey{Namespace: application.Status.Generated.Namespace,
		Name: application.Status.Generated.Name}
	fmt.Println("Expecting plotter to be fetchable")
	g.Eventually(func() error {
		return k8sClient.Get(context.Background(), plotterObjectKey, plotter)
	}, timeout, interval).Should(gomega.Succeed())

	fmt.Println("Expecting application to be ready")
	g.Eventually(func() bool {
		err = k8sClient.Get(context.Background(), applicationKey, application)
		if err != nil {
			return false
		}
		return application.Status.Ready
	}, timeout, interval).Should(gomega.Equal(true))

	modulesNamespace = plotter.Spec.ModulesNamespace
	fmt.Printf("data access module namespace notebook test: %s\n", modulesNamespace)

	g.Expect(application.Status.AssetStates[catalogedAsset].Endpoint.Name).ToNot(gomega.BeEmpty())
	g.Expect(application.Status.AssetStates[catalogedAsset].Conditions[ReadyConditionIndex].Status).To(gomega.Equal(v1.ConditionTrue))

	// Forward port of arrow flight service to local port
	connection := application.Status.AssetStates[catalogedAsset].
		Endpoint.AdditionalProperties.Items["fybrik-arrow-flight"].(map[string]interface{})
	hostname := fmt.Sprintf("%v", connection["hostname"])
	port := fmt.Sprintf("%v", connection["port"])
	svcName := strings.Replace(hostname, "."+modulesNamespace, "", 1)

	defer func() {
		if !success {
			// Print information that might be useful for debugging the deployed module
			// before fybrikapplication is deleted:
			// - print the logs of the deployed module pod
			// - print the config-maps in the module namespace
			AFMservice := &v1.Service{}
			AFMserviceObjectKey := client.ObjectKey{Namespace: modulesNamespace,
				Name: svcName}
			g.Eventually(func() error {
				return k8sClient.Get(context.Background(), AFMserviceObjectKey, AFMservice)
			}, timeout, interval).Should(gomega.Succeed())

			pods, _ := fybrikUtils.GetPodsForSvc(clientset, AFMservice, modulesNamespace)
			for i := range pods.Items {
				fmt.Printf("pod name: %s\n", pods.Items[i].Name)
				fmt.Printf("pod logs:\n")
				fmt.Println(fybrikUtils.GetPodLogs(clientset, &pods.Items[i]))
			}
			cms, _ := fybrikUtils.GetConfigMapForNamespace(clientset, modulesNamespace)
			for i := range cms.Items {
				if strings.HasSuffix(cms.Items[i].Name, ".crt") {
					continue
				}
				fmt.Printf("config-map name: %s\n", cms.Items[i].Name)
				b, _ := json.Marshal(cms.Items[i].Data)
				fmt.Println(string(b))
			}
		}
		fybrikApplication := &fapp.FybrikApplication{ObjectMeta: metav1.ObjectMeta{Namespace: applicationKey.Namespace,
			Name: applicationKey.Name}}
		_ = k8sClient.Get(context.Background(), applicationKey, fybrikApplication)
		_ = k8sClient.Delete(context.Background(), fybrikApplication)
	}()

	fmt.Println("Starting kubectl port-forward for arrow-flight")
	portNum, err := strconv.Atoi(port)
	g.Expect(err).To(gomega.BeNil())
	listenPort, err := test.RunPortForward(modulesNamespace, svcName, portNum)
	g.Expect(err).To(gomega.BeNil())

	// Reading data via arrow flight
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(), grpc.WithTimeout(timeout))
	flightClient, err := flight.NewFlightClient(net.JoinHostPort("localhost", listenPort), nil, opts...)
	g.Expect(err).To(gomega.BeNil(), "Connect to arrow-flight service")
	defer flightClient.Close()

	request := ArrowRequest{
		Asset: catalogedAsset,
	}

	marshal, err := json.Marshal(request)
	g.Expect(err).To(gomega.BeNil())

	info, err := flightClient.GetFlightInfo(context.Background(), &flight.FlightDescriptor{
		Type: flight.FlightDescriptor_CMD,
		Cmd:  marshal,
	})
	g.Expect(err).To(gomega.BeNil())

	stream, err := flightClient.DoGet(context.Background(), info.Endpoint[0].Ticket)
	g.Expect(err).To(gomega.BeNil())

	reader, err := flight.NewRecordReader(stream)
	defer reader.Release()
	g.Expect(err).To(gomega.BeNil())
	var record arrow.Record

	for reader.Next() {
		record = reader.Record()

		g.Expect(record.ColumnName(0)).To(gomega.Equal("step"))
		g.Expect(record.ColumnName(1)).To(gomega.Equal("type"))
		g.Expect(record.ColumnName(3)).To(gomega.Equal("nameOrig"))
		column := record.Column(3) // Check out the third 4th column that should be nameOrig and redacted

		// Check that data of nameOrig column is the correct size and all records are redacted
		dt := column.Data().DataType()
		g.Expect(column.Data().Len()).To(gomega.Equal(100))
		g.Expect(dt.ID()).To(gomega.Equal(arrow.STRING))
		g.Expect(dt.Name()).To(gomega.Equal((&arrow.StringType{}).Name()))
		data := array.NewStringData(column.Data())
		for i := 0; i < data.Len(); i++ {
			g.Expect(data.Value(i)).To(gomega.Equal("XXXXX"))
		}
	}
	record.Release()
	success = true
}
