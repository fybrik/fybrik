package app

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/onsi/gomega"
	netv1 "k8s.io/api/networking/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	"fybrik.io/fybrik/manager/controllers/utils"
	managerUtils "fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/helm"
	"fybrik.io/fybrik/pkg/logging"
)

const my_cluster = "MyCluster"

// This test checks that a short release name is not truncated
func TestCreateNPIngressRules(t *testing.T) {
	g := gomega.NewWithT(t)
	log := logging.LogInit(logging.CONTROLLER, "test-np-igress")

	// Create a fake client to mock API calls.
	cl := fake.NewClientBuilder().Build()
	// Register operator types with the runtime scheme.
	s := utils.NewScheme(g)

	r := &BlueprintReconciler{
		Client: cl,
		Name:   "TestCreateNPIngressRules",
		Log:    log,
		Scheme: s,
		Helmer: helm.NewEmptyFake(),
	}

	ports := []netv1.NetworkPolicyPort{{Protocol: &tcp}}

	expectedIngressRules := getDefaultNetworkPolicyIngressRules()

	// check default Ingress rules creation
	rules, err := r.createNPIngressRules(false, nil, nil, my_cluster, &log)
	g.Expect(err).To(gomega.BeNil(), "cannot create default NP IngressRules")
	g.Expect(rules).To(gomega.Equal(expectedIngressRules))

	// check mismatching error when endpoint is true, but application details is nil
	rules, err = r.createNPIngressRules(true, nil, nil, my_cluster, &log)
	g.Expect(err).Should(gomega.MatchError(NilApplicationDetailsError), "nil application error is not thrown")

	// check mismatching error when endpoint is true, but application details do not provide any information about possible
	// user workloads.
	app := fapp.ApplicationDetails{}
	rules, err = r.createNPIngressRules(true, nil, &app, my_cluster, &log)
	g.Expect(err).Should(gomega.MatchError(EmptyApplicationDetailsError), "empty application error is not thrown")

	compareRules := func(endPoint bool, ingresses []fapp.ModuleDeployment, app *fapp.ApplicationDetails,
		cluster string, expectedRules []netv1.NetworkPolicyIngressRule) {
		rules, err = r.createNPIngressRules(endPoint, ingresses, app, cluster, &log)
		_, file, line, ok := runtime.Caller(1)
		var msg string
		if !ok {
			msg = "Failed to get caller information"
		}
		msg = fmt.Sprintf("caller: %s:%d", file, line)
		g.Expect(err).To(gomega.BeNil(), msg)
		g.Expect(len(rules)).To(gomega.Equal(1), msg)
		g.Expect(len(rules[0].Ports)).To(gomega.Equal(1), msg)
		g.Expect(len(rules[0].From)).To(gomega.Equal(len(expectedIngressRules[0].From)), msg)
		g.Expect(rules[0].From).To(gomega.ConsistOf(expectedIngressRules[0].From), msg)
	}

	// check a use-case, when endpoint is true, and only workload labels are defined.
	appName := "my-app"
	workloadSelector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesAppName: appName}}
	app.WorkloadSelector = workloadSelector
	from := []netv1.NetworkPolicyPeer{{PodSelector: &workloadSelector}}
	expectedIngressRules = []netv1.NetworkPolicyIngressRule{{From: from, Ports: ports}}
	compareRules(true, nil, &app, my_cluster, expectedIngressRules)

	// check a use-case, when endpoint is true, and both workload labels and IPBlocks are defined.
	IPBlocks := []*netv1.IPBlock{{CIDR: "10.100.102.0/16"}, {CIDR: "14.144.256.27/32"}, {CIDR: "2001:0db8:85a3::/64"}}
	app.IPBlocks = IPBlocks
	from = []netv1.NetworkPolicyPeer{}
	for _, block := range IPBlocks {
		expectedIngressRules[0].From = append(expectedIngressRules[0].From, netv1.NetworkPolicyPeer{IPBlock: block})
	}
	compareRules(true, nil, &app, my_cluster, expectedIngressRules)

	// check a use-case, when endpoint is true, and workload labels, namespaces and IPBlocks are defined.
	namespaces := []string{"fybrik1", "fybrik2"}
	app.Namespaces = namespaces
	from = []netv1.NetworkPolicyPeer{}
	for _, ns := range namespaces {
		nsSelector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesNamespaceName: ns}}
		from = append(from, netv1.NetworkPolicyPeer{PodSelector: &workloadSelector, NamespaceSelector: &nsSelector})
	}
	for _, block := range IPBlocks {
		from = append(from, netv1.NetworkPolicyPeer{IPBlock: block})
	}
	expectedIngressRules = []netv1.NetworkPolicyIngressRule{{From: from, Ports: ports}}
	compareRules(true, nil, &app, my_cluster, expectedIngressRules)

	// check a use-case, when endpoint is true, and workload labels, namespaces and IPBlocks are defined. In addition,
	// 2 ingresses are defined too.
	ingresses := []fapp.ModuleDeployment{{Cluster: my_cluster, Release: "myapp-111"}, {Release: "myapp-123"}}
	for _, ingress := range ingresses {
		selector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesInstance: ingress.Release}}
		from = append(from, netv1.NetworkPolicyPeer{PodSelector: &selector})
	}
	expectedIngressRules = []netv1.NetworkPolicyIngressRule{{From: from, Ports: ports}}
	compareRules(true, ingresses, &app, my_cluster, expectedIngressRules)

	// check a use-case, when endpoint is false, and workload labels, namespaces and IPBlocks are defined. In addition,
	// 2 ingresses are defined too.
	from = []netv1.NetworkPolicyPeer{}
	for _, ingress := range ingresses {
		selector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesInstance: ingress.Release}}
		from = append(from, netv1.NetworkPolicyPeer{PodSelector: &selector})
	}
	expectedIngressRules = []netv1.NetworkPolicyIngressRule{{From: from, Ports: ports}}
	ingresses = append(ingresses, fapp.ModuleDeployment{Cluster: "my_cluster"})
	compareRules(false, ingresses, &app, my_cluster, expectedIngressRules)
}
