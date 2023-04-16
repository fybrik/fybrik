// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	managerUtils "fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/helm"
	"fybrik.io/fybrik/pkg/logging"
)

const myCluster = "MyCluster"

// This test checks Network Policies ingress rules creation
func TestCreateNPIngressRules(t *testing.T) {
	g := gomega.NewWithT(t)
	log := logging.LogInit(logging.CONTROLLER, "test-np-ingress")

	// Create a fake client to mock API calls.
	cl := fake.NewClientBuilder().Build()
	// Register operator types with the runtime scheme.
	s := managerUtils.NewScheme(g)

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
	rules, err := r.createNPIngressRules(false, nil, nil, myCluster, &log)
	g.Expect(err).To(gomega.BeNil(), "cannot create default NP IngressRules")
	g.Expect(rules).To(gomega.Equal(expectedIngressRules))

	// check mismatching error when endpoint is true, but application details is nil
	_, err = r.createNPIngressRules(true, nil, nil, myCluster, &log)
	g.Expect(err).Should(gomega.MatchError(NilApplicationDetailsError), "nil application error is not thrown")

	// check mismatching error when endpoint is true, but application details do not provide any information about possible
	// user workloads.
	app := fapp.ApplicationDetails{}
	_, err = r.createNPIngressRules(true, nil, &app, myCluster, &log)
	g.Expect(err).Should(gomega.MatchError(EmptyApplicationDetailsError), "empty application error is not thrown")

	compareRules := func(endPoint bool, ingresses []fapp.ModuleDeployment, app *fapp.ApplicationDetails,
		cluster string, expectedRules []netv1.NetworkPolicyIngressRule) {
		rules, err = r.createNPIngressRules(endPoint, ingresses, app, cluster, &log)
		_, file, line, ok := runtime.Caller(1)
		var msg string
		if !ok {
			msg = "Failed to get caller information"
		} else {
			msg = fmt.Sprintf("caller: %s:%d", file, line)
		}
		g.Expect(err).To(gomega.BeNil(), msg)
		g.Expect(len(rules)).To(gomega.Equal(1), msg)
		g.Expect(len(rules[0].Ports)).To(gomega.Equal(1), msg)
		g.Expect(len(rules[0].From)).To(gomega.Equal(len(expectedRules[0].From)), msg)
		g.Expect(rules[0].From).To(gomega.ConsistOf(expectedRules[0].From), msg)
	}

	// check a use-case, when endpoint is true, and only workload labels are defined.
	appName := "my-app"
	workloadSelector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesAppName: appName}}
	app.WorkloadSelector = workloadSelector
	from := []netv1.NetworkPolicyPeer{{PodSelector: &workloadSelector}}
	expectedIngressRules = []netv1.NetworkPolicyIngressRule{{From: from, Ports: ports}}
	compareRules(true, nil, &app, myCluster, expectedIngressRules)

	// check a use-case, when endpoint is true, and both workload labels and IPBlocks are defined.
	IPBlocks := []*netv1.IPBlock{{CIDR: "10.100.102.0/16"}, {CIDR: "14.144.256.27/32"}, {CIDR: "2001:0db8:85a3::/64"}}
	app.IPBlocks = IPBlocks
	for _, block := range IPBlocks {
		expectedIngressRules[0].From = append(expectedIngressRules[0].From, netv1.NetworkPolicyPeer{IPBlock: block})
	}
	compareRules(true, nil, &app, myCluster, expectedIngressRules)

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
	compareRules(true, nil, &app, myCluster, expectedIngressRules)

	// check a use-case, when endpoint is true, and workload labels, namespaces and IPBlocks are defined. In addition,
	// 2 ingresses are defined too.
	ingresses := []fapp.ModuleDeployment{{Cluster: myCluster, Release: "myapp-111"}, {Release: "myapp-123"}}
	for _, ingress := range ingresses {
		selector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesInstance: ingress.Release}}
		from = append(from, netv1.NetworkPolicyPeer{PodSelector: &selector})
	}
	expectedIngressRules = []netv1.NetworkPolicyIngressRule{{From: from, Ports: ports}}
	compareRules(true, ingresses, &app, myCluster, expectedIngressRules)

	// check a use-case, when endpoint is false, and workload labels, namespaces and IPBlocks are defined. In addition,
	// 2 ingresses are defined too.
	from = []netv1.NetworkPolicyPeer{}
	for _, ingress := range ingresses {
		selector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesInstance: ingress.Release}}
		from = append(from, netv1.NetworkPolicyPeer{PodSelector: &selector})
	}
	expectedIngressRules = []netv1.NetworkPolicyIngressRule{{From: from, Ports: ports}}
	ingresses = append(ingresses, fapp.ModuleDeployment{Cluster: "myCluster"})
	compareRules(false, ingresses, &app, myCluster, expectedIngressRules)
}

// This test checks Network Policies egress rules creation
func TestCreateNPEgressRules(t *testing.T) {
	g := gomega.NewWithT(t)
	log := logging.LogInit(logging.CONTROLLER, "test-np-egress")

	modulesNamespace := "fybrik-modules"

	service := corev1.Service{}
	service.Name = "vault"
	service.Namespace = "fybrik-system"
	service.Spec.Ports = []corev1.ServicePort{{Name: "http", Protocol: tcp, TargetPort: intstr.FromInt(8080)},
		{Name: "https-internal", Protocol: tcp, TargetPort: intstr.FromInt(8081)}}
	service.Spec.Selector = map[string]string{managerUtils.KubernetesAppName: "myApp"}
	// Create a fake client to mock API calls.
	cl := fake.NewClientBuilder().WithRuntimeObjects(&service).Build()
	// Register operator types with the runtime scheme.
	s := managerUtils.NewScheme(g)

	r := &BlueprintReconciler{
		Client: cl,
		Name:   "TestCreateNPIngressRules",
		Log:    log,
		Scheme: s,
		Helmer: helm.NewEmptyFake(),
	}
	{
		// check default DNS rules
		expectedRules := []netv1.NetworkPolicyEgressRule{dnsEngressRules}
		rules := r.createNPEgressRules(context.Background(), nil, nil, myCluster, modulesNamespace, &log)
		g.Expect(expectedRules).To(CompareNPEgressRules(rules))
	}
	{
		// check ModuleNetwork.Engress settings
		release1 := "my-release-111"
		release2 := "my-release-222"
		release3 := "my-release-333"
		egresses := []fapp.ModuleDeployment{
			{Cluster: myCluster, Release: release1},
			{Cluster: myCluster + "-test", Release: release2}, // another cluster should be skipped for now
			{Release: release3, URLs: []string{release3 + "-1123:8080", "123" + release3 + ":8090"}},
		}
		podSelector1 := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesInstance: release1}}
		to := netv1.NetworkPolicyPeer{PodSelector: &podSelector1}
		expectedRules := []netv1.NetworkPolicyEgressRule{dnsEngressRules, {To: []netv1.NetworkPolicyPeer{to}}}
		podSelector2 := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesInstance: release3}}
		to = netv1.NetworkPolicyPeer{PodSelector: &podSelector2}
		port := intstr.FromInt(8080)
		npPorts := []netv1.NetworkPolicyPort{{Protocol: &tcp, Port: &port}}
		expectedRules = append(expectedRules, netv1.NetworkPolicyEgressRule{To: []netv1.NetworkPolicyPeer{to}, Ports: npPorts})
		rules := r.createNPEgressRules(context.Background(), egresses, nil, myCluster, modulesNamespace, &log)
		g.Expect(expectedRules).To(CompareNPEgressRules(rules))
	}
	{
		// check internal service
		podSelector := meta.LabelSelector{MatchLabels: service.Spec.Selector}
		namespaceSelector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesNamespaceName: service.Namespace}}
		to := netv1.NetworkPolicyPeer{PodSelector: &podSelector, NamespaceSelector: &namespaceSelector}
		npPorts := []netv1.NetworkPolicyPort{}
		for i := range service.Spec.Ports {
			npPorts = append(npPorts, netv1.NetworkPolicyPort{Protocol: &service.Spec.Ports[i].Protocol,
				Port: &service.Spec.Ports[i].TargetPort})
		}
		expectedRules := []netv1.NetworkPolicyEgressRule{dnsEngressRules, {To: []netv1.NetworkPolicyPeer{to}, Ports: npPorts}}
		rules := r.createNPEgressRules(context.Background(), nil,
			[]string{service.Name + "." + service.Namespace}, myCluster, modulesNamespace, &log)
		g.Expect(expectedRules).To(CompareNPEgressRules(rules))
	}
	{
		// check IP block
		ipBlocks := []string{"192.168.0.0/24", "2001:db8::/64"}
		expectedRules := []netv1.NetworkPolicyEgressRule{dnsEngressRules}
		for _, block := range ipBlocks {
			if _, _, err := net.ParseCIDR(block); err == nil {
				ipBlock := netv1.IPBlock{CIDR: block}
				expectedRules = append(expectedRules, netv1.NetworkPolicyEgressRule{To: []netv1.NetworkPolicyPeer{{IPBlock: &ipBlock}}})
			}
		}
		rules := r.createNPEgressRules(context.Background(), nil, ipBlocks, myCluster, modulesNamespace, &log)
		g.Expect(expectedRules).To(CompareNPEgressRules(rules))
	}
	{
		// check external IPs
		urls := []string{"192.168.0.23", "192.168.1.25:80" /*, TODO: parseURl doesn't correctly parse IPv6. "2001:db8::68" */}
		expectedRules := []netv1.NetworkPolicyEgressRule{dnsEngressRules}
		for _, urlStr := range urls {
			ipPort := strings.Split(urlStr, ":")
			if ip := net.ParseIP(ipPort[0]); ip != nil {
				ipBlock := ipToIPBlock(ip)
				to := []netv1.NetworkPolicyPeer{{IPBlock: &ipBlock}}
				port := netv1.NetworkPolicyPort{Protocol: &tcp}
				if len(ipPort) > 1 {
					portInt, err := strconv.Atoi(ipPort[1])
					if err != nil {
						t.Errorf("cannot transfer %s to port", ipPort[1])
					}
					p := intstr.FromInt(portInt)
					port.Port = &p
				}
				expectedRules = append(expectedRules, netv1.NetworkPolicyEgressRule{To: to, Ports: []netv1.NetworkPolicyPort{port}})
			}
		}
		rules := r.createNPEgressRules(context.Background(), nil, urls, myCluster, modulesNamespace, &log)
		g.Expect(expectedRules).To(CompareNPEgressRules(rules))
	}
	{
		// check external hosts
		// TODO: can we use it in the unit tests
		urlHost := "www.google.com"
		urlPort := 80
		urlString := fmt.Sprintf("http://%s:%d", urlHost, urlPort)
		ips, err := net.LookupIP(urlHost)
		if err != nil {
			t.Errorf("Cannot lookupIPs of %s", urlHost)
		}
		expectedRules := []netv1.NetworkPolicyEgressRule{dnsEngressRules}
		to := []netv1.NetworkPolicyPeer{}
		for _, ip := range ips {
			ipBlock := ipToIPBlock(ip)
			to = append(to, netv1.NetworkPolicyPeer{IPBlock: &ipBlock})
		}
		p := intstr.FromInt(urlPort)
		port := netv1.NetworkPolicyPort{Protocol: &tcp, Port: &p}
		expectedRules = append(expectedRules, netv1.NetworkPolicyEgressRule{To: to, Ports: []netv1.NetworkPolicyPort{port}})

		rules := r.createNPEgressRules(context.Background(), nil, []string{urlString}, myCluster, modulesNamespace, &log)
		g.Expect(expectedRules).To(CompareNPEgressRules(rules))
	}
}

func CompareNPEgressRules(rules []netv1.NetworkPolicyEgressRule) types.GomegaMatcher {
	return &egressRulesMatcher{
		Rules: rules,
	}
}

type egressRulesMatcher struct {
	Rules []netv1.NetworkPolicyEgressRule
}

func (matcher *egressRulesMatcher) Match(actual interface{}) (bool, error) {
	actualRules, ok := actual.([]netv1.NetworkPolicyEgressRule)
	if !ok {
		return false, fmt.Errorf("compareNPEgressRules matcher expects an array/slice of netv1.NetworkPolicyEgressRule. Got:%T",
			actual)
	}
	if len(actualRules) != len(matcher.Rules) {
		return false, nil
	}
	for _, actualRule := range actualRules {
		toMatcher := gomega.ConsistOf(actualRule.To)
		portsMatcher := gomega.ConsistOf(actualRule.Ports)
		exist := false
		for _, expectRule := range matcher.Rules {
			ok, err := portsMatcher.Match(expectRule.Ports)
			if err != nil {
				return false, err
			}
			if !ok {
				continue
			}
			ok, err = toMatcher.Match(expectRule.To)
			if err != nil {
				return false, err
			}
			if ok {
				exist = true
				break
			}
		}
		if !exist {
			return false, nil
		}
	}
	return true, nil
}

func (matcher *egressRulesMatcher) FailureMessage(actual interface{}) string {
	return format.Message(actual, "to compare NP EgressRules of", matcher.Rules)
}

func (matcher *egressRulesMatcher) NegatedFailureMessage(actual interface{}) string {
	return ""
}
