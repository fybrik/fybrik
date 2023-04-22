// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/rs/zerolog"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	managerUtils "fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/environment"
	"fybrik.io/fybrik/pkg/logging"
)

const (
	KubeSystemNamespace   = "kube-system"
	OpenShiftDNSNamespace = "openshift-dns"
	KubeDNSValue          = "kube-dns"
	DNSPortName           = "dns"
	DNSTCPPortName        = "dns-tcp"

	NilApplicationDetailsError   = "Misconfiguration, endpoint with nil application details"
	EmptyApplicationDetailsError = "Misconfiguration, endpoint with empty application details"
	CannotParseURLError          = "cannot parse %s as URL"
)

var dnsEgressRules = createDNSEgressRules()

var tcp = corev1.ProtocolTCP
var udp = corev1.ProtocolUDP

// allow DNS access
func createDNSEgressRules() netv1.NetworkPolicyEgressRule {
	dnsPort := intstr.FromString(DNSPortName)
	dnsTCPPort := intstr.FromString(DNSTCPPortName)
	policyPorts := []netv1.NetworkPolicyPort{{Protocol: &udp, Port: &dnsPort}, {Protocol: &tcp, Port: &dnsTCPPort}}
	if environment.IsOpenShiftDeployment() {
		nsSelector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesNamespaceName: OpenShiftDNSNamespace}}
		podSelector := meta.LabelSelector{
			MatchExpressions: []meta.LabelSelectorRequirement{{Key: managerUtils.OpenShiftDNS, Operator: meta.LabelSelectorOpExists}}}
		npPeer := netv1.NetworkPolicyPeer{PodSelector: &podSelector, NamespaceSelector: &nsSelector}
		return netv1.NetworkPolicyEgressRule{To: []netv1.NetworkPolicyPeer{npPeer}, Ports: policyPorts}
	}
	nsSelector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesNamespaceName: KubeSystemNamespace}}
	podSelector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesAppName: KubeDNSValue}}
	podSelectorOld := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesAppNameOld: KubeDNSValue}}
	npPeer := netv1.NetworkPolicyPeer{PodSelector: &podSelector, NamespaceSelector: &nsSelector}
	npPeerOld := netv1.NetworkPolicyPeer{PodSelector: &podSelectorOld, NamespaceSelector: &nsSelector}
	return netv1.NetworkPolicyEgressRule{To: []netv1.NetworkPolicyPeer{npPeer, npPeerOld}, Ports: policyPorts}
}

// Defines the default Network Polices peer if no input was provided
// returning nil, means denying all ingress connections
// returning empty NetworkPolicyIngressRule array, means allowing all ingress connections
func getDefaultNetworkPolicyIngressRules() []netv1.NetworkPolicyIngressRule {
	return nil
}

func (r *BlueprintReconciler) createNetworkPolicies(ctx context.Context,
	releaseName string, network *fapp.ModuleNetwork, blueprint *fapp.Blueprint, logger *zerolog.Logger) error {
	log := logger.With().Str(managerUtils.KubernetesInstance, releaseName).Logger()
	log.Trace().Str(logging.ACTION, logging.CREATE).Msg("Creating Network Policies for  " + releaseName)

	np, err := r.createNetworkPoliciesDefinition(ctx, releaseName, network, blueprint, &log)
	if err != nil {
		return err
	}
	res, err := ctrlutil.CreateOrUpdate(ctx, r.Client, np, func() error { return nil })
	if err != nil {
		return errors.WithMessagef(err, "failed to create NetworkPolicy: %v", np)
	}
	log.Trace().Str(logging.ACTION, logging.CREATE).Msgf("Network Policies for %s/%s were createdOrUpdated result: %s",
		np.Namespace, releaseName, res)
	return nil
}

func (r *BlueprintReconciler) createNetworkPoliciesDefinition(ctx context.Context, releaseName string, network *fapp.ModuleNetwork,
	blueprint *fapp.Blueprint, log *zerolog.Logger) (*netv1.NetworkPolicy, error) {
	np := netv1.NetworkPolicy{}
	np.Name = releaseName
	np.Namespace = blueprint.Spec.ModulesNamespace
	labelsMap := blueprint.Labels
	np.Labels = managerUtils.CopyFybrikLabels(labelsMap)
	podsSelector := meta.LabelSelector{}
	podsSelector.MatchLabels = map[string]string{managerUtils.KubernetesInstance: releaseName}
	np.Spec.PodSelector = podsSelector

	np.Spec.PolicyTypes = []netv1.PolicyType{netv1.PolicyTypeEgress, netv1.PolicyTypeIngress}

	ingress, err := r.createNPIngressRules(network.Endpoint, network.Ingress, blueprint.Spec.Application, blueprint.Spec.Cluster, log)
	if err != nil {
		return nil, err
	}
	np.Spec.Ingress = ingress
	egress := r.createNPEgressRules(ctx, network.Egress, network.URLs, blueprint.Spec.Cluster, blueprint.Spec.ModulesNamespace, log)

	np.Spec.Egress = egress
	return &np, nil
}

func (r *BlueprintReconciler) createNPIngressRules(endpoint bool, ingresses []fapp.ModuleDeployment,
	application *fapp.ApplicationDetails, cluster string, log *zerolog.Logger) ([]netv1.NetworkPolicyIngressRule, error) {
	log.Trace().Str(logging.ACTION, logging.CREATE).Msgf("Ingress rules creation from Endpoint:%v appDetails %v and Ingress: %v",
		endpoint, application, ingresses)

	var from []netv1.NetworkPolicyPeer
	// access from user workloads
	// TODO: we don't check cluster here, because meantime we don't support workloads from other clusters.
	if endpoint {
		if application == nil {
			err := errors.New(NilApplicationDetailsError)
			log.Err(err)
			return nil, err
		}
		workLoadSelector := application.WorkloadSelector
		namespaces := application.Namespaces
		ipBlocks := application.IPBlocks
		if len(ipBlocks) == 0 && len(namespaces) == 0 && workLoadSelector.Size() == 0 {
			err := errors.New(EmptyApplicationDetailsError)
			log.Err(err)
			return nil, err
		}
		for _, ip := range ipBlocks {
			npPeer := netv1.NetworkPolicyPeer{IPBlock: ip}
			from = append(from, npPeer)
		}
		if len(namespaces) == 0 {
			npPeer := netv1.NetworkPolicyPeer{PodSelector: &workLoadSelector}
			from = append(from, npPeer)
		}
		for _, ns := range namespaces {
			nsSelector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesNamespaceName: ns}}
			npPeer := netv1.NetworkPolicyPeer{NamespaceSelector: &nsSelector, PodSelector: &workLoadSelector}
			from = append(from, npPeer)
		}
	}

	for _, ingress := range ingresses {
		if ingress.Cluster == "" || ingress.Cluster == cluster {
			// local cluster
			if ingress.Release != "" {
				selector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesInstance: ingress.Release}}
				npPeer := netv1.NetworkPolicyPeer{PodSelector: &selector}
				from = append(from, npPeer)
			} else {
				log.Warn().Str(logging.ACTION, logging.CREATE).Msgf("Ingress has empty release %v", ingress)
			}
		} else {
			// TODO: multi-cluster support
			log.Debug().Str(logging.ACTION, logging.CREATE).Msgf("Cross-cluster ingress connectivity, ingress.Cluster %s, blueprint cluster %s",
				ingress.Cluster, cluster)
		}
	}
	if len(from) == 0 {
		return getDefaultNetworkPolicyIngressRules(), nil
	}
	return []netv1.NetworkPolicyIngressRule{{From: from, Ports: []netv1.NetworkPolicyPort{{Protocol: &tcp}}}}, nil
}

func (r *BlueprintReconciler) createNPEgressRules(ctx context.Context, egresses []fapp.ModuleDeployment, urls []string,
	cluster string, modulesNamespace string, log *zerolog.Logger) []netv1.NetworkPolicyEgressRule {
	log.Trace().Str(logging.ACTION, logging.CREATE).
		Msgf("Egress rules creation from network.Egress: %v, network.URLs %v",
			egresses, urls)
	egressRules := []netv1.NetworkPolicyEgressRule{}

	modulesEgressRules := r.createNextModulesEgressRules(egresses, cluster, log)
	egressRules = append(egressRules, modulesEgressRules...)

	// The URLs can be a CIDR block, or urls to the cluster internal or external services.
	for _, urlString := range urls {
		log.Trace().Msgf("Processing external URL %s", urlString)

		// Check if it is a CIDR (Classless Inter-Domain Routing)
		_, _, err := net.ParseCIDR(urlString)
		if err == nil {
			ipBlock := netv1.IPBlock{CIDR: urlString}
			to := []netv1.NetworkPolicyPeer{{IPBlock: &ipBlock}}
			egressRules = append(egressRules, netv1.NetworkPolicyEgressRule{To: to})
			continue
		}
		// parse URL
		servURL, err := managerUtils.ParseRawURL(urlString)
		if err != nil {
			log.Err(err).Msgf(CannotParseURLError, urlString)
			continue
		}
		hostName := servURL.Hostname()
		if hostName == "" {
			log.Warn().Msgf("URL without host name: %s", servURL)
			continue
		}
		// Check if the hostName is actually an IP address.
		// NOTE: IP address to a local service will not work
		ip := net.ParseIP(hostName)
		if ip != nil {
			ipBlock := ipToIPBlock(ip)
			to := []netv1.NetworkPolicyPeer{{IPBlock: &ipBlock}}
			policyPort := policyPortFromURL(servURL, log)
			egressRules = append(egressRules, netv1.NetworkPolicyEgressRule{To: to, Ports: []netv1.NetworkPolicyPort{policyPort}})
			continue
		}
		if environment.CreateNP4ServiceDestination() {
			// Check if it is a local service
			hostStrings := strings.Split(hostName, ".")
			service := corev1.Service{}
			key := types.NamespacedName{Name: hostStrings[0]}
			if len(hostStrings) > 1 {
				key.Namespace = hostStrings[1]
			} else {
				key.Namespace = modulesNamespace
			}
			// we assume that the service exists
			if err = r.Get(ctx, key, &service); err != nil {
				log.Info().Msgf("Get service returned error. %v", service)
			} else {
				// TODO: check NodePort and LoadBalancer
				podSelector := meta.LabelSelector{MatchLabels: service.Spec.Selector}
				npPeer := netv1.NetworkPolicyPeer{PodSelector: &podSelector}
				if len(hostStrings) > 1 {
					nsSelector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesNamespaceName: hostStrings[1]}}
					npPeer.NamespaceSelector = &nsSelector
				}
				to := []netv1.NetworkPolicyPeer{npPeer}
				npPorts := []netv1.NetworkPolicyPort{}
				for _, port := range service.Spec.Ports {
					protocol := port.Protocol
					targetPort := port.TargetPort
					npPorts = append(npPorts, netv1.NetworkPolicyPort{Protocol: &protocol, Port: &targetPort})
				}
				egressRules = append(egressRules, netv1.NetworkPolicyEgressRule{To: to, Ports: npPorts})
				if !environment.CreateNP4Service() {
					continue
				}
			}
		}
		// 3. deal with external service names
		ips, err := net.LookupIP(hostName)
		if err != nil {
			log.Err(err).Msgf("cannot get IP addresses for %s", hostName)
			continue
		}
		to := []netv1.NetworkPolicyPeer{}
		for _, ip := range ips {
			ipBlock := ipToIPBlock(ip)
			to = append(to, netv1.NetworkPolicyPeer{IPBlock: &ipBlock})
		}
		policyPort := policyPortFromURL(servURL, log)
		egressRules = append(egressRules, netv1.NetworkPolicyEgressRule{To: to, Ports: []netv1.NetworkPolicyPort{policyPort}})
	}
	egressRules = append(egressRules, dnsEgressRules)
	return egressRules
}

func (r *BlueprintReconciler) createNextModulesEgressRules(egresses []fapp.ModuleDeployment, cluster string,
	log *zerolog.Logger) []netv1.NetworkPolicyEgressRule {
	var egressRules []netv1.NetworkPolicyEgressRule
	for _, egress := range egresses {
		if egress.Cluster == "" || egress.Cluster == cluster {
			if egress.Release != "" {
				selector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesInstance: egress.Release}}
				to := netv1.NetworkPolicyPeer{PodSelector: &selector}
				// TODO: get port from local service, combine with the multi-cluster implementation.
				/*
					var npPorts []netv1.NetworkPolicyPort
					for _, urlString := range egress.URLs {
						u, err := managerUtils.ParseRawURL(urlString)
						if err != nil {
							log.Err(err).Msgf(CannotParseURLError, urlString)
							continue
						}
						if strings.HasPrefix(u.Hostname(), egress.Release) {
							policyPort := policyPortFromURL(u, log)
							npPorts = append(npPorts, policyPort)
							continue
						}
						log.Warn().Msgf("Egress URL %s is not part of release %s", urlString, egress.Release

					} */
				egressRules = append(egressRules, netv1.NetworkPolicyEgressRule{To: []netv1.NetworkPolicyPeer{to}})
			}
		} else {
			// TODO: multi-cluster support
			log.Debug().Str(logging.ACTION, logging.CREATE).Msgf("Cross-cluster egress connectivity, egress.Cluster %s, blueprint cluster %s",
				egress.Cluster, cluster)
		}
	}

	return egressRules
}

func (r *BlueprintReconciler) cleanupNetworkPolicies(ctx context.Context, blueprint *fapp.Blueprint) error {
	l := client.MatchingLabels{}
	l[managerUtils.ApplicationNameLabel] = blueprint.Labels[managerUtils.ApplicationNameLabel]
	l[managerUtils.ApplicationNamespaceLabel] = blueprint.Labels[managerUtils.ApplicationNamespaceLabel]
	l[managerUtils.BlueprintNameLabel] = blueprint.Name
	l[managerUtils.BlueprintNamespaceLabel] = blueprint.Namespace
	r.Log.Trace().Str(logging.ACTION, logging.DELETE).Msgf("Delete Network Policies with labels %v", l)
	if err := r.Client.DeleteAllOf(ctx, &netv1.NetworkPolicy{},
		client.InNamespace(environment.GetDefaultModulesNamespace()), l); err != nil {
		r.Log.Error().Err(err).Msg("Error while deleting Network Policies")
		return err
	}
	r.Log.Trace().Str(logging.ACTION, logging.DELETE).Msg("Network Polices were deleted")
	return nil
}

func policyPortFromURL(ur *url.URL, log *zerolog.Logger) netv1.NetworkPolicyPort {
	portString := ur.Port()
	if portString == "" {
		// TODO: add default ports
		return netv1.NetworkPolicyPort{Protocol: &tcp}
	}
	return policyPortFromString(portString, log)
}

func policyPortFromString(portString string, log *zerolog.Logger) netv1.NetworkPolicyPort {
	portInt, err := strconv.Atoi(portString)
	if err != nil {
		log.Err(err).Msgf("cannot convert port %s to integer", portString)
		return netv1.NetworkPolicyPort{Protocol: &tcp}
	}
	port := intstr.FromInt(portInt)
	return netv1.NetworkPolicyPort{Protocol: &tcp, Port: &port}
}

func ipToIPBlock(ip net.IP) netv1.IPBlock {
	if ipv4 := ip.To4(); ipv4 != nil {
		return netv1.IPBlock{CIDR: ip.String() + "/32"}
	}
	return netv1.IPBlock{CIDR: ip.String() + "/64"}
}
