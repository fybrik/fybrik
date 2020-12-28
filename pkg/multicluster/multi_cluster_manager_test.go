package multicluster

import (
	"github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestDecodeJsonToRuntimeObject(t *testing.T) {
	var json = `
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
	 "name": "d1",
	 "namespace": "default"
  },
  "spec": {
    "replicas": 2,
    "template": {
	  "spec": {
	    "containers": [
		  {
		    "name": "container",
            "image": "image"
		  }
	    ]
	  }
    }
  }
}
`
	g := gomega.NewGomegaWithT(t)
	actualDeployment := apps.Deployment{}
	scheme := runtime.NewScheme()
	// utilruntime.Must(apps.AddToScheme(scheme))
	if err := Decode(json, scheme, &actualDeployment); err != nil {
		println("failed decoding")
		t.Error(err)
	}
	var replicaNumber int32 = 2

	expectedDeployment := apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "d1",
			Namespace: "default",
		},
		Spec: apps.DeploymentSpec{
			Replicas: &replicaNumber,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "container",
							Image: "image",
						},
					},
				},
			},
		},
	}
	g.Expect(expectedDeployment).To(gomega.Equal(actualDeployment))
}
