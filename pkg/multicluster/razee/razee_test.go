package razee

import (
	"fmt"
	"github.com/ibm/the-mesh-for-data/pkg/multicluster"
	"os"

	//"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"

	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	razeeTestURL string = "http://localhost:3333/graphql"
	loginTestUser string = "razee-dev@example.com"
	razeeTestPassword string = "password123"
	selfLink string = "/api/v1/namespaces/razeedeploy/configmaps/watch-keeper-cluster-metadata"
)

var (
	orgId = os.Getenv("RazeeOrgId")
	clusterId =os.Getenv("RazeeClusterId")
)

var mySemantic = conversion.EqualitiesOrDie(
	func(a, b resource.Quantity) bool {
		return a.Cmp(b) == 0
	},
	func(a, b metav1.MicroTime) bool {
		return a.UTC() == b.UTC()
	},
	// We ignore creation time for now
	func(a, b metav1.Time) bool {
		return true
	},
	func(a, b labels.Selector) bool {
		return a.String() == b.String()
	},
	func(a, b fields.Selector) bool {
		return a.String() == b.String()
	},
)

func TestGetResourceByKeys(t *testing.T) {
	razeeClient := NewRazeeLocalClient(razeeTestURL, loginTestUser, razeeTestPassword)
	if jsonData, err := razeeClient.getResourceByKeys(orgId, clusterId, selfLink); err != nil {
		t.Error(err)
	} else {
		t.Log("json of configmap got from razee-client: ", jsonData)
		actualConfigMap := v1.ConfigMap{}
		scheme := runtime.NewScheme()
		if err := multicluster.Decode(jsonData, scheme, &actualConfigMap); err != nil {
			println("failed decoding")
			t.Error(err)
		}

		expectedConfigMap := v1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "watch-keeper-cluster-metadata",
				Namespace: "razeedeploy",
				SelfLink: "/api/v1/namespaces/razeedeploy/configmaps/watch-keeper-cluster-metadata",
				UID: "4d94a73e-1686-412e-b0d1-84cddfd7d6e5",
				ResourceVersion: "1089",
				Labels: map[string]string {
					"razee/cluster-metadata": "true",
					"razee/watch-resource": "lite",
				},
				Annotations: map[string]string{
					"selfLink": "/api/v1/namespaces/razeedeploy/configmaps/watch-keeper-cluster-metadata",
				},
			},
		}
		fmt.Printf("expectedConfigMap and actualConfigMap are equal (without compating time): %t\n",mySemantic.DeepEqual(expectedConfigMap, actualConfigMap))
		if !mySemantic.DeepEqual(expectedConfigMap, actualConfigMap) {
			t.Fail()
		}
	}
}