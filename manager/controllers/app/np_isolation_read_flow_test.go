// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	managerUtils "fybrik.io/fybrik/manager/controllers/utils"
)

func uploadToS3(endpoint string, g gomega.Gomega) {
	region := "theshire"
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
		fmt.Printf("error %v\n", err)
		log.Println("Object already exists in S3!")
	}
}

func TestNetworkPolicyReadFlow(t *testing.T) {
	fmt.Println("network policy test")
	valuesYaml, ok := os.LookupEnv("VALUES_FILE")
	if !ok || !(strings.Contains(valuesYaml, readFlow)) {
		t.Skip("Only executed for notebook tests")
	}
	isolation, ok := os.LookupEnv("RUN_ISOLATION")
	if !ok || isolation != "1" {
		t.Skip("Only executed when isolation in enabled")
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
	endpoint := "http://localhost:9090"
	// S3 duplicate is assumed to be exposed on localhost at port 9393
	endpointDup := "http://localhost:9393"
	// upload to the s3 store
	uploadToS3(endpoint, g)
	// upload to the duplicate s3 store
	uploadToS3(endpointDup, g)

	err := fapp.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	k8sClient, err := client.New(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme.Scheme}) //nolint:govet
	g.Expect(err).To(gomega.BeNil())

	// Create Kubernetes objects for test
	// - namespace (in setup before)
	// - asset (in setup before)
	// - asset secret (in setup before)
	// - fybrik modules (in setup before)
	// - rego policy

	// Deploy policy from a configmap
	piiReadConfigMap := &v1.ConfigMap{}
	// Create a redact PII policy
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

	// Installing application
	application := &fapp.FybrikApplication{}
	plotter := &fapp.Plotter{}
	var applicationKey client.ObjectKey
	var plotterObjectKey client.ObjectKey
	var modulesNamespace string

	g.Expect(readObjectFromFile("../../testdata/notebook/read-flow/fybrikapplication-isolation.yaml", application)).
		ToNot(gomega.HaveOccurred())
	application.ObjectMeta.Name += "-1"
	application.Spec.Data[0].DataSetID = catalogedAsset
	applicationKey = client.ObjectKeyFromObject(application)

	// Create FybrikApplication
	fmt.Println("Expecting application creation to succeed")
	g.Expect(k8sClient.Create(context.Background(), application)).Should(gomega.Succeed())

	// Ensure getting cleaned up after tests finish
	// delete application
	defer func() {
		fybrikApplication := &fapp.FybrikApplication{ObjectMeta: metav1.ObjectMeta{Namespace: applicationKey.Namespace,
			Name: applicationKey.Name}}
		_ = k8sClient.Get(context.Background(), applicationKey, fybrikApplication)
		_ = k8sClient.Delete(context.Background(), fybrikApplication)
	}()

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

	g.Expect(application.Status.AssetStates[catalogedAsset].Endpoint.Name).ToNot(gomega.BeEmpty())
	g.Expect(application.Status.AssetStates[catalogedAsset].Conditions[ReadyConditionIndex].Status).To(gomega.Equal(v1.ConditionTrue))

	// Get the connection endpoint
	connection := application.Status.AssetStates[catalogedAsset].
		Endpoint.AdditionalProperties.Items["fybrik-arrow-flight"].(map[string]interface{})
	hostname := fmt.Sprintf("%v", connection["hostname"])
	port := fmt.Sprintf("%v", connection["port"])

	// using my-shell pod to read
	gvk := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	}

	restClient, err := apiutil.RESTClientForGVK(gvk, false, ctrl.GetConfigOrDie(), serializer.NewCodecFactory(scheme.Scheme))
	g.Expect(err).To(gomega.BeNil())

	readCommand := "python3 /root/client.py --host " + hostname + " --port " + port + " --asset " + catalogedAsset
	var stdout, stderr bytes.Buffer
	fmt.Println("Expecting successful read from the namespace of the module")
	// Add the application label
	podObj := &v1.Pod{}
	podObjKey := client.ObjectKey{Namespace: modulesNamespace, Name: "my-shell"}
	err = k8sClient.Get(context.Background(), podObjKey, podObj)
	g.Expect(err).To(gomega.BeNil())
	podObj.ObjectMeta.Labels["app"] = "my-app"
	err = k8sClient.Update(context.Background(), podObj)
	time.Sleep(20 * time.Second)
	err = managerUtils.ExecPod(restClient, ctrl.GetConfigOrDie(), "my-shell", modulesNamespace, readCommand, nil, &stdout, &stderr)
	g.Expect(err).To(gomega.BeNil())
	stdout.Reset()
	stderr.Reset()

	// Changing the label
	podObj = &v1.Pod{}
	podObjKey = client.ObjectKey{Namespace: modulesNamespace, Name: "my-shell"}
	err = k8sClient.Get(context.Background(), podObjKey, podObj)
	podObj.ObjectMeta.Labels["app"] = "my-app1"
	err = k8sClient.Update(context.Background(), podObj)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(err).To(gomega.BeNil())
	fmt.Println("Expecting Reading command to fail now")
	err = managerUtils.ExecPod(restClient, ctrl.GetConfigOrDie(), "my-shell", modulesNamespace, readCommand, nil, &stdout, &stderr)
	g.Expect(err).ToNot(gomega.BeNil())
	stdout.Reset()
	stderr.Reset()

	// Try to read from other namespace
	fmt.Println("Expecting Reading from default namespace to fail")
	err = managerUtils.ExecPod(restClient, ctrl.GetConfigOrDie(), "my-shell", "default", readCommand, nil, &stdout, &stderr)
	g.Expect(err).ToNot(gomega.BeNil())
	stdout.Reset()
	stderr.Reset()

	// change labels and try to read
	podObj = &v1.Pod{}
	podObjKey = client.ObjectKey{Namespace: "default", Name: "my-shell"}
	err = k8sClient.Get(context.Background(), podObjKey, podObj)
	g.Expect(err).To(gomega.BeNil())
	podObj.ObjectMeta.Labels["app"] = "my-app"
	err = k8sClient.Update(context.Background(), podObj)
	g.Expect(err).To(gomega.BeNil())
	fmt.Println("Expecting Reading from default namespace with labels to fail")
	err = managerUtils.ExecPod(restClient, ctrl.GetConfigOrDie(), "my-shell", "default", readCommand, nil, &stdout, &stderr)
	g.Expect(err).ToNot(gomega.BeNil())
	stdout.Reset()
	stderr.Reset()

	// Check connection to the second module
	err = k8sClient.Get(context.Background(), plotterObjectKey, plotter)
	g.Expect(err).To(gomega.BeNil())
	steps := plotter.Spec.Flows[0].SubFlows[0].Steps
	g.Expect(len(plotter.Spec.Flows[0].SubFlows[0].Steps)).To(gomega.Equal(1))
	g.Expect(len(plotter.Spec.Flows[0].SubFlows[0].Steps[0])).To(gomega.Equal(2))
	var hostnameToCheck string
	var portToCheck string
	for _, step := range steps[0] {
		connectionInterface := step.Parameters.API.Connection.AdditionalProperties.Items["fybrik-arrow-flight"]
		connectionMap, ok := connectionInterface.(map[string]interface{})
		g.Expect(ok).To(gomega.Equal(true))
		hostnameTmp := fmt.Sprintf("%v", connectionMap["hostname"])
		if hostnameTmp == hostname {
			continue
		} else {
			hostnameToCheck = hostnameTmp
			portToCheck = fmt.Sprintf("%v", connection["port"])
			g.Expect(ok).To(gomega.Equal(true))
		}
	}
	readCommand = "python3 /root/client.py --host " + hostnameToCheck + " --port " + portToCheck + " --asset " + catalogedAsset
	fmt.Println("Expecting reading from the second module to fail")
	// Add the application label
	podObjKey = client.ObjectKey{Namespace: modulesNamespace, Name: "my-shell"}
	err = k8sClient.Get(context.Background(), podObjKey, podObj)
	g.Expect(err).To(gomega.BeNil())
	podObj.ObjectMeta.Labels["app"] = "my-app"
	err = k8sClient.Update(context.Background(), podObj)
	err = managerUtils.ExecPod(restClient, ctrl.GetConfigOrDie(), "my-shell", modulesNamespace, readCommand, nil, &stdout, &stderr)

	g.Expect(err).ToNot(gomega.BeNil())
	stdout.Reset()
	stderr.Reset()

	// Check egress connection
	moduleConfigMapList := &v1.ConfigMapList{}
	opts := &client.ListOptions{
		Namespace: "fybrik-blueprints",
	}
	err = k8sClient.List(context.Background(), moduleConfigMapList, opts)
	for i, configMap := range moduleConfigMapList.Items {
		confYaml, ok := configMap.Data["conf.yaml"]
		if !ok {
			continue
		}
		var yamlData map[string]interface{}
		err = yaml.Unmarshal([]byte(confYaml), &yamlData)
		g.Expect(err).To(gomega.BeNil())
		// Check if this configmap has an s3 connection
		val, ok := yamlData["data"].([]interface{})[0].(map[interface{}]interface{})["connection"].(map[interface{}]interface{})["s3"]
		if !ok {
			continue
		}
		// Change the endpoint to the second s3 storage
		val.(map[interface{}]interface{})["endpoint_url"] = "http://s3-dup.fybrik-system:9090"
		newYaml, err := yaml.Marshal(yamlData)
		g.Expect(err).To(gomega.BeNil())
		configMap.Data["conf.yaml"] = string(newYaml)
		err = k8sClient.Update(context.Background(), &moduleConfigMapList.Items[i])
		g.Expect(err).To(gomega.BeNil())
		time.Sleep(100 * time.Second)
		fmt.Println("Expecting Reading command to fail because the module not allowed to connect to the second s3 storage")
		readCommand = "python3 /root/client.py --host " + hostname + " --port " + port + " --asset " + catalogedAsset
		err = managerUtils.ExecPod(restClient, ctrl.GetConfigOrDie(), "my-shell", modulesNamespace, readCommand, nil, &stdout, &stderr)
		g.Expect(err).ToNot(gomega.BeNil())
		stdout.Reset()
		stderr.Reset()
		break
	}

	fmt.Println("isolation read flow test succeeded")
}
