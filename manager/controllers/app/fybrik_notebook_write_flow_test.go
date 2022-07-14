// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/apache/arrow/go/v7/arrow"
	"github.com/apache/arrow/go/v7/arrow/array"
	"github.com/apache/arrow/go/v7/arrow/csv"
	"github.com/apache/arrow/go/v7/arrow/flight"
	"github.com/apache/arrow/go/v7/arrow/ipc"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"fybrik.io/fybrik/connectors/katalog/pkg/apis/katalog/v1alpha1"
	apiv1alpha1 "fybrik.io/fybrik/manager/apis/app/v1alpha1"
	"fybrik.io/fybrik/pkg/test"
)

func TestS3NotebookWriteFlow(t *testing.T) {
	if s, ok := os.LookupEnv("VALUES_FILE"); !ok || s != "charts/fybrik/notebook-test-writeflow.values.yaml" {
		t.Skip("Only executed for notebook tests")
	}
	gomega.RegisterFailHandler(Fail)

	g := gomega.NewGomegaWithT(t)
	defer GinkgoRecover()

	err := apiv1alpha1.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	k8sClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme.Scheme}) //nolint:govet
	g.Expect(err).To(gomega.BeNil())

	// This test checks the following senarios:
	// (a) how fybrik prevents writing new asset due to governance restrictions
	// (b) how to write data generated by the workload to an object store.
	// (c) how to read data from a dataset stored in an object store

	fmt.Printf("Starting deny write scenario")

	forbidWriteConfigMap := &v1.ConfigMap{}

	g.Expect(readObjectFromFile("../../testdata/notebook/write-flow/forbid-policy-cm.yaml", forbidWriteConfigMap)).ToNot(gomega.HaveOccurred())
	forbidWriteConfigMapKey := client.ObjectKeyFromObject(forbidWriteConfigMap)
	g.Expect(k8sClient.Create(context.Background(), forbidWriteConfigMap)).Should(gomega.Succeed())

	fmt.Printf("Expecting config-map to be created")
	g.Eventually(func() error {
		return k8sClient.Get(context.Background(), forbidWriteConfigMapKey, forbidWriteConfigMap)
	}, timeout, interval).Should(gomega.Succeed())
	fmt.Printf("Expecting config-map to be constructed")
	g.Eventually(func() string {
		_ = k8sClient.Get(context.Background(), forbidWriteConfigMapKey, forbidWriteConfigMap)
		return forbidWriteConfigMap.Annotations["openpolicyagent.org/policy-status"]
	}, timeout, interval).Should(gomega.BeEquivalentTo("{\"status\":\"ok\"}"))

	// Module installed by setup script directly from remote arrow-flight-module repository
	// Installing application
	writeApplication := &apiv1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/notebook/write-flow/fybrikapplication-write.yaml", writeApplication)).
		ToNot(gomega.HaveOccurred())
	writeApplicationKey := client.ObjectKeyFromObject(writeApplication)

	// Create FybrikApplication and FybrikModule
	fmt.Printf("Expecting write application creation to succeed")
	g.Expect(k8sClient.Create(context.Background(), writeApplication)).Should(gomega.Succeed())

	fmt.Printf("Expecting write application to be created")
	g.Eventually(func() error {
		return k8sClient.Get(context.Background(), writeApplicationKey, writeApplication)
	}, timeout, interval).Should(gomega.Succeed())

	fmt.Printf("Expecting write application to be ready")
	g.Eventually(func() bool {
		err = k8sClient.Get(context.Background(), writeApplicationKey, writeApplication)
		if err != nil {
			return false
		}
		return writeApplication.Status.Ready
	}, timeout, interval).Should(gomega.Equal(true))

	// Expect to get deny status due to governance restrictions
	g.Expect(writeApplication.Status.AssetStates["new-data"].Conditions[DenyConditionIndex].Status).To(gomega.Equal(v1.ConditionTrue))

	// cleanup
	g.Eventually(func() error {
		return k8sClient.Delete(context.Background(), forbidWriteConfigMap)
	}, timeout, interval).Should(gomega.Succeed())
	g.Eventually(func() error {
		return k8sClient.Delete(context.Background(), writeApplication)
	}, timeout, interval).Should(gomega.Succeed())

	fmt.Printf("Starting allow write scenario")

	writePolicyConfigMap := &v1.ConfigMap{}
	g.Expect(
		readObjectFromFile("../../testdata/notebook/write-flow/write-read-policy-cm.yaml", writePolicyConfigMap)).ToNot(gomega.HaveOccurred())
	writePolicyConfigMapKey := client.ObjectKeyFromObject(writePolicyConfigMap)
	g.Expect(k8sClient.Create(context.Background(), writePolicyConfigMap)).Should(gomega.Succeed())

	fmt.Printf("Expecting config-map to be created")
	g.Eventually(func() error {
		return k8sClient.Get(context.Background(), writePolicyConfigMapKey, writePolicyConfigMap)
	}, timeout, interval).Should(gomega.Succeed())
	fmt.Printf("Expecting config-map to be constructed")
	g.Eventually(func() string {
		_ = k8sClient.Get(context.Background(), writePolicyConfigMapKey, writePolicyConfigMap)
		return writePolicyConfigMap.Annotations["openpolicyagent.org/policy-status"]
	}, timeout, interval).Should(gomega.BeEquivalentTo("{\"status\":\"ok\"}"))

	// Module installed by setup script directly from remote arrow-flight-module repository
	// Installing application
	writeApplication = &apiv1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/notebook/write-flow/fybrikapplication-write.yaml", writeApplication)).
		ToNot(gomega.HaveOccurred())
	writeApplicationKey = client.ObjectKeyFromObject(writeApplication)
	// Ensure getting cleaned up after tests finish
	defer func() {
		application := &apiv1alpha1.FybrikApplication{ObjectMeta: metav1.ObjectMeta{Namespace: writeApplicationKey.Namespace,
			Name: writeApplicationKey.Name}}
		_ = k8sClient.Get(context.Background(), writeApplicationKey, application)
		_ = k8sClient.Delete(context.Background(), application)
	}()

	// Create FybrikApplication and FybrikModule
	fmt.Printf("Expecting write application creation to succeed")
	g.Expect(k8sClient.Create(context.Background(), writeApplication)).Should(gomega.Succeed())

	fmt.Printf("Expecting write application to be created")
	g.Eventually(func() error {
		return k8sClient.Get(context.Background(), writeApplicationKey, writeApplication)
	}, timeout, interval).Should(gomega.Succeed())

	fmt.Printf("Expecting plotter to be constructed")
	g.Eventually(func() *apiv1alpha1.ResourceReference {
		_ = k8sClient.Get(context.Background(), writeApplicationKey, writeApplication)
		return writeApplication.Status.Generated
	}, timeout, interval).ShouldNot(gomega.BeNil())

	// The plotter has to be created
	plotter := &apiv1alpha1.Plotter{}
	plotterObjectKey := client.ObjectKey{Namespace: writeApplication.Status.Generated.Namespace, Name: writeApplication.Status.Generated.Name}
	fmt.Printf("Expecting plotter to be fetchable")
	g.Eventually(func() error {
		return k8sClient.Get(context.Background(), plotterObjectKey, plotter)
	}, timeout, interval).Should(gomega.Succeed())

	fmt.Printf("Expecting write application to be ready")
	g.Eventually(func() bool {
		err = k8sClient.Get(context.Background(), writeApplicationKey, writeApplication)
		if err != nil {
			return false
		}
		return writeApplication.Status.Ready
	}, timeout, interval).Should(gomega.Equal(true))

	modulesNamespace := plotter.Spec.ModulesNamespace
	fmt.Printf("data access module namespace notebook test: %s\n", modulesNamespace)

	g.Expect(len(writeApplication.Status.AssetStates)).To(gomega.Equal(1))
	g.Expect(writeApplication.Status.AssetStates["new-data"].CatalogedAsset).
		ToNot(gomega.BeEmpty())
	newAssetID := writeApplication.Status.AssetStates["new-data"].
		CatalogedAsset
	newCatalogID := "fybrik-notebook-sample"

	g.Expect(len(writeApplication.Status.ProvisionedStorage)).To(gomega.Equal(1))
	// check provisioned storage
	g.Expect(writeApplication.Status.ProvisionedStorage["new-data"].DatasetRef).
		ToNot(gomega.BeEmpty(), "No storage provisioned")

	// Get the new connection details
	var newBucket, newObject string
	connectionMap := writeApplication.Status.ProvisionedStorage["new-data"].
		Details.Connection.AdditionalProperties.Items
	g.Expect(connectionMap).To(gomega.HaveKey("s3"))
	s3Conn := connectionMap["s3"].(map[string]interface{})
	g.Expect(s3Conn["endpoint"]).To(gomega.Equal("http://s3.fybrik-system.svc.cluster.local:9090"))
	newBucket = fmt.Sprint(s3Conn["bucket"])
	newObject = fmt.Sprint(s3Conn["object_key"])
	g.Expect(newBucket).NotTo(gomega.BeEmpty())
	g.Expect(newObject).NotTo(gomega.BeEmpty())

	err = v1alpha1.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	asset := &v1alpha1.Asset{}
	fmt.Printf("Expecting asset to be fetchable")
	assetObjectKey := client.ObjectKey{Namespace: newCatalogID, Name: newAssetID}
	g.Eventually(func() error {
		return k8sClient.Get(context.Background(), assetObjectKey, asset)
	}, timeout, interval).Should(gomega.Succeed())

	g.Expect(asset.Spec.Metadata.Geography).To(gomega.Equal("theshire"))
	g.Expect(writeApplication.Status.AssetStates["new-data"].Endpoint.Name).ToNot(gomega.BeEmpty())
	// Forward port of arrow flight service to local port
	connection := writeApplication.Status.AssetStates["new-data"].
		Endpoint.AdditionalProperties.Items["fybrik-arrow-flight"].(map[string]interface{})
	hostname := fmt.Sprintf("%v", connection["hostname"])
	port := fmt.Sprintf("%v", connection["port"])
	svcName := strings.Replace(hostname, "."+modulesNamespace, "", 1)

	fmt.Printf("Starting kubectl port-forward for arrow-flight")
	portNum, err := strconv.Atoi(port)
	g.Expect(err).To(gomega.BeNil())
	listenPort, err := test.RunPortForward(modulesNamespace, svcName, portNum)
	g.Expect(err).To(gomega.BeNil())

	// Writing data via arrow flight
	opts := make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())
	flightClient, err := flight.NewFlightClient(net.JoinHostPort("localhost", listenPort), nil, opts...)
	g.Expect(err).To(gomega.BeNil(), "Connect to arrow-flight service")
	defer flightClient.Close()

	// Prepare to write the data
	// construct the data schema
	stepField := arrow.Field{Name: "step", Type: arrow.PrimitiveTypes.Int64}
	typeField := arrow.Field{Name: "type", Type: arrow.BinaryTypes.String}
	amountField := arrow.Field{Name: "amount", Type: arrow.PrimitiveTypes.Float64}
	nameOrigField := arrow.Field{Name: "nameOrig", Type: arrow.BinaryTypes.String}
	oldbalanceOrgField := arrow.Field{Name: "oldbalanceOrg", Type: arrow.PrimitiveTypes.Float64}
	newbalanceOrigField := arrow.Field{Name: "newbalanceOrig", Type: arrow.PrimitiveTypes.Float64}
	nameDestField := arrow.Field{Name: "nameDest", Type: arrow.BinaryTypes.String}
	oldbalanceDestField := arrow.Field{Name: "oldbalanceDest", Type: arrow.PrimitiveTypes.Float64}
	newbalanceDestField := arrow.Field{Name: "newbalanceDest", Type: arrow.PrimitiveTypes.Float64}
	isFraudField := arrow.Field{Name: "isFraud", Type: arrow.PrimitiveTypes.Int64}
	isFlaggedFraudField := arrow.Field{Name: "isFlaggedFraud", Type: arrow.PrimitiveTypes.Int64}
	sc := arrow.NewSchema([]arrow.Field{stepField, typeField, amountField, nameOrigField,
		oldbalanceOrgField, newbalanceOrigField, nameDestField, oldbalanceDestField,
		newbalanceDestField, isFraudField, isFlaggedFraudField}, nil)

	filepath := "../../../samples/kubeflow/data.csv"
	raw, err := os.ReadFile(filepath)
	g.Expect(err).To(gomega.BeNil(), "read data file")

	csvReader := csv.NewReader(
		bytes.NewReader(raw), sc,
		csv.WithComment('#'), csv.WithComma(','),
		csv.WithChunk(10), csv.WithHeader(true),
	)
	defer csvReader.Release()

	request := ArrowRequest{
		Asset: "new-data",
	}

	marshal, err := json.Marshal(request)
	g.Expect(err).To(gomega.BeNil())

	// write the data to a new asset
	writeStream, err := flightClient.DoPut(context.Background())
	g.Expect(err).To(gomega.BeNil())

	wr := flight.NewRecordWriter(writeStream, ipc.WithSchema(sc))

	descr := &flight.FlightDescriptor{
		Type: flight.FlightDescriptor_CMD,
		Cmd:  marshal,
	}
	wr.SetFlightDescriptor(descr)

	var rec arrow.Record
	// write the records
	for csvReader.Next() {
		rec = csvReader.Record()
		err = wr.Write(rec)
		g.Expect(err).To(gomega.BeNil())
	}
	rec.Release()
	wr.Close()
	err = writeStream.CloseSend()
	g.Expect(err).To(gomega.BeNil())

	fmt.Printf("Starting read scenario")
	// Installing application to read new data
	readApplication := &apiv1alpha1.FybrikApplication{}
	g.Expect(readObjectFromFile("../../testdata/notebook/write-flow/fybrikapplication-read.yaml", readApplication)).
		ToNot(gomega.HaveOccurred())
	// Update the name of the dataset id
	readApplication.Spec.Data[0].DataSetID = newCatalogID + "/" + newAssetID
	readApplicationKey := client.ObjectKeyFromObject(readApplication)

	// Create FybrikApplication
	fmt.Printf("Expecting read application creation to succeed")
	g.Expect(k8sClient.Create(context.Background(), readApplication)).Should(gomega.Succeed())

	fmt.Printf("Expecting read application to be created")
	g.Eventually(func() error {
		return k8sClient.Get(context.Background(), readApplicationKey, readApplication)
	}, timeout, interval).Should(gomega.Succeed())
	// Ensure getting cleaned up after tests finish
	defer func() {
		application := &apiv1alpha1.FybrikApplication{ObjectMeta: metav1.ObjectMeta{Namespace: readApplicationKey.Namespace,
			Name: readApplicationKey.Name}}
		_ = k8sClient.Get(context.Background(), readApplicationKey, application)
		_ = k8sClient.Delete(context.Background(), application)
	}()
	fmt.Printf("Expecting plotter to be constructed")
	g.Eventually(func() *apiv1alpha1.ResourceReference {
		_ = k8sClient.Get(context.Background(), readApplicationKey, readApplication)
		return readApplication.Status.Generated
	}, timeout, interval).ShouldNot(gomega.BeNil())

	// The plotter has to be created
	plotter = &apiv1alpha1.Plotter{}
	plotterObjectKey = client.ObjectKey{Namespace: readApplication.Status.Generated.Namespace, Name: readApplication.Status.Generated.Name}
	fmt.Printf("Expecting plotter to be fetchable")
	g.Eventually(func() error {
		return k8sClient.Get(context.Background(), plotterObjectKey, plotter)
	}, timeout, interval).Should(gomega.Succeed())

	fmt.Printf("Expecting read application to be ready")
	g.Eventually(func() bool {
		err = k8sClient.Get(context.Background(), readApplicationKey, readApplication)
		if err != nil {
			return false
		}
		return readApplication.Status.Ready
	}, timeout, interval).Should(gomega.Equal(true))

	modulesNamespace = plotter.Spec.ModulesNamespace
	fmt.Printf("data access module namespace notebook test: %s\n", modulesNamespace)

	// Forward port of arrow flight service to local port
	connection = readApplication.Status.AssetStates["fybrik-notebook-sample/"+newAssetID].
		Endpoint.AdditionalProperties.Items["fybrik-arrow-flight"].(map[string]interface{})
	hostname = fmt.Sprintf("%v", connection["hostname"])
	port = fmt.Sprintf("%v", connection["port"])
	svcName = strings.Replace(hostname, "."+modulesNamespace, "", 1)

	fmt.Printf("Starting kubectl port-forward for arrow-flight")
	portNum, err = strconv.Atoi(port)
	g.Expect(err).To(gomega.BeNil())
	listenPort, err = test.RunPortForward(modulesNamespace, svcName, portNum)
	g.Expect(err).To(gomega.BeNil())

	// Reading data via arrow flight
	opts = make([]grpc.DialOption, 0)
	opts = append(opts, grpc.WithInsecure())
	flightClient, err = flight.NewFlightClient(net.JoinHostPort("localhost", listenPort), nil, opts...)
	g.Expect(err).To(gomega.BeNil(), "Connect to arrow-flight service")
	defer flightClient.Close()

	request = ArrowRequest{
		Asset: newCatalogID + "/" + newAssetID,
	}

	marshal, err = json.Marshal(request)
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
		column := record.Column(4) // Check out the third 1th column that should be oldbalanceOrg and redacted

		// Check that data of oldbalanceOrg column is the correct size and all records are redacted
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
}
