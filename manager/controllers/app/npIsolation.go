// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fybrik.io/fybrik/pkg/logging"
	"github.com/rs/zerolog/log"
	"net"
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
)

const (
	KubeSystemNamespace   = "kube-system"
	OpenShiftDNSNamespace = " openshift-dns"
	KubeDNSValue          = "kube-dns"
	DNSPortName           = "dns"
	DNSTCPPortName        = "dns-tcp"
)

var dnsEngressRules = createDNSEngressRules()

func createDNSEngressRules() netv1.NetworkPolicyEgressRule {
	// allow DNS access
	udp := corev1.ProtocolUDP
	tcp := corev1.ProtocolTCP
	dnsPort := intstr.FromString(DNSPortName)
	dnsTCPPort := intstr.FromString(DNSTCPPortName)
	policyPorts := []netv1.NetworkPolicyPort{{Protocol: &udp, Port: &dnsPort}, {Protocol: &tcp, Port: &dnsTCPPort}}
	if environment.IsOpenShiftDeployment() {
		nsSelector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesNamespaceName: OpenShiftDNSNamespace}}
		podSelector := meta.LabelSelector{MatchExpressions: []meta.LabelSelectorRequirement{{Key: managerUtils.OpenShiftDNS, Operator: meta.LabelSelectorOpExists}}}
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
// returning empty NetworkPolicyPeer array, means allowing all ingress connections
func getDefaultNetworkPolicyFrom() []netv1.NetworkPolicyPeer {
	return nil
}

func (r *BlueprintReconciler) createNetworkPolicies(ctx context.Context,
	releaseName string, network fapp.ModuleNetwork, blueprint *fapp.Blueprint, logger *zerolog.Logger) error {
	log := logger.With().Str(managerUtils.KubernetesInstance, releaseName).Logger()
	log.Trace().Str(logging.ACTION, logging.CREATE).Msg("Creating Network Policies for  " + releaseName)

	np, err := r.createNetworkPoliciesDefinition(ctx, releaseName, network, blueprint, &log)
	if err != nil {
		return err
	}
	res, err := ctrlutil.CreateOrUpdate(ctx, r.Client, np, func() error { return nil })
	if err != nil {
		return errors.WithMessage(err, releaseName+": failed to create NetworkPolicy")
	}
	log.Trace().Str(logging.ACTION, logging.CREATE).Msgf("Network Policies for %s/%s were createdOrUpdated result: %s", np.Namespace, releaseName, res)
	return nil
}

func (r *BlueprintReconciler) createNetworkPoliciesDefinition(ctx context.Context, releaseName string, network fapp.ModuleNetwork,
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

	ingress, err := r.createNPIngressRules(releaseName, network, blueprint, log)
	if err != nil {
		return nil, err
	}
	np.Spec.Ingress = ingress

	egress, err := r.createNPEgressRules(ctx, releaseName, network, blueprint, log)
	if err != nil {
		return nil, err
	}
	np.Spec.Egress = egress
	return &np, nil
}

func (r *BlueprintReconciler) createNPIngressRules(releaseName string, network fapp.ModuleNetwork,
	blueprint *fapp.Blueprint, log *zerolog.Logger) ([]netv1.NetworkPolicyIngressRule, error) {
	tcp := corev1.ProtocolTCP
	npPorts := []netv1.NetworkPolicyPort{{Protocol: &tcp}}
	var from []netv1.NetworkPolicyPeer
	ingressRules := []netv1.NetworkPolicyIngressRule{}

	if network.Endpoint {
		workLoadSelector := blueprint.Spec.Application.WorkloadSelector
		namespaces := blueprint.Spec.Application.Namespaces
		ipBlocks := blueprint.Spec.Application.IPBlocks
		// 1. check that something is defined
		if len(ipBlocks) == 0 && len(namespaces) == 0 && len(workLoadSelector.MatchExpressions) == 0 &&
			len(workLoadSelector.MatchLabels) == 0 {
			from = getDefaultNetworkPolicyFrom()
		} else {
			from = []netv1.NetworkPolicyPeer{}
			if len(namespaces) == 0 {
				npPeer := netv1.NetworkPolicyPeer{PodSelector: &workLoadSelector}
				from = append(from, npPeer)
			}
			for _, ns := range namespaces {
				nsSelector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesNamespaceName: ns}}
				npPeer := netv1.NetworkPolicyPeer{NamespaceSelector: &nsSelector, PodSelector: &workLoadSelector}
				from = append(from, npPeer)
			}
			for _, ip := range ipBlocks {
				npPeer := netv1.NetworkPolicyPeer{IPBlock: ip}
				from = append(from, npPeer)
			}
		}
		ingressRules = append(ingressRules, netv1.NetworkPolicyIngressRule{From: from, Ports: npPorts})
	}
	if len(network.Ingress) > 0 {
		from = []netv1.NetworkPolicyPeer{}
		for _, ingress := range network.Ingress {
			if ingress.Cluster == "" || ingress.Cluster == blueprint.Spec.Cluster {
				if ingress.Release != "" {
					for _, url := range ingress.URLs {
						if strings.HasPrefix(url, releaseName) {
							continue
						} else {
							// TODO: can we have the URL for ingress, which is not form a module?
						}
					}
					selector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesInstance: ingress.Release}}
					npPeer := netv1.NetworkPolicyPeer{PodSelector: &selector}
					from = append(from, npPeer)
				}
			} else {
				// TODO: multi-cluster support
			}
		}
		if len(from) > 0 {
			ingressRules = append(ingressRules, netv1.NetworkPolicyIngressRule{From: from, Ports: npPorts})
		}
	}
	return ingressRules, nil
}

func (r *BlueprintReconciler) createNPEgressRules(ctx context.Context, releaseName string, network fapp.ModuleNetwork,
	blueprint *fapp.Blueprint, log *zerolog.Logger) ([]netv1.NetworkPolicyEgressRule, error) {
	tcp := corev1.ProtocolTCP
	npPorts := []netv1.NetworkPolicyPort{{Protocol: &tcp}}
	var to []netv1.NetworkPolicyPeer
	egressRules := []netv1.NetworkPolicyEgressRule{}
	if len(network.Egress) > 0 {
		to = []netv1.NetworkPolicyPeer{}
		for _, egress := range network.Egress {
			if egress.Cluster == "" || egress.Cluster == blueprint.Spec.Cluster {
				if egress.Release != "" {
					selector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesInstance: egress.Release}}
					npPeer := netv1.NetworkPolicyPeer{PodSelector: &selector}
					to = append(to, npPeer)
				}
			} else {
				// TODO: multi-cluster support
			}
		}
		if len(to) > 0 {
			// TODO: add ports
			egressRules = append(egressRules, netv1.NetworkPolicyEgressRule{To: to, Ports: npPorts})
		}
	}
	for _, urlString := range network.URLs {
		log.Trace().Msgf("Processing external URL %s", urlString)

		// 1. check if it is a CIDR (Classless Inter-Domain Routing)
		// in this case be in a form 192.168.1.0/24 and optionally port separated by colon
		stringsArray := strings.Split(urlString, ":")
		_, _, err := net.ParseCIDR(stringsArray[0])
		if err == nil {
			ipBlock := netv1.IPBlock{CIDR: stringsArray[0]}
			to = []netv1.NetworkPolicyPeer{{IPBlock: &ipBlock}}
			if len(stringsArray) > 1 {
				if port, err := strconv.Atoi(stringsArray[1]); err == nil {
					pr := intstr.FromInt(port)
					egressRules = append(egressRules, netv1.NetworkPolicyEgressRule{To: to, Ports: []netv1.NetworkPolicyPort{{Port: &pr}}})
					continue
				}
				log.Debug().Msgf("Cannot parse port %s", stringsArray[1])
			}
			egressRules = append(egressRules, netv1.NetworkPolicyEgressRule{To: to})
			continue
		}
		// 2 parse URL
		url, err := managerUtils.ParseRawURL(urlString)
		if err != nil {
			// TODO
			log.Err(err).Msgf("cannot parse %s as URL", urlString)
			continue
		}
		hostName := url.Hostname()
		if hostName == "" {
			log.Warn().Msgf("URL without host name: %s", url)
			continue
		}
		// 2.1. check if the hostName is actually an IP address.
		// NOTE: IP address to a local service will not work
		ip := net.ParseIP(hostName)
		if ip != nil {
			// TODO: deal with IP
		}
		// 2.2. Check if it is a local service
		hostStrings := strings.Split(hostName, ".")
		service := corev1.Service{}
		key := types.NamespacedName{Name: hostStrings[0]}
		if len(hostStrings) > 1 {
			key.Namespace = hostStrings[1]
		} else {
			// TODO: add namespace
		}
		if err := r.Get(ctx, key, &service); err == nil {
			// TODO: check NodePort and LoadBalancer
			podSelector := meta.LabelSelector{MatchLabels: service.Spec.Selector}
			npPeer := netv1.NetworkPolicyPeer{PodSelector: &podSelector}
			if len(hostStrings) > 1 {
				nsSelector := meta.LabelSelector{MatchLabels: map[string]string{managerUtils.KubernetesNamespaceName: hostStrings[1]}}
				npPeer.NamespaceSelector = &nsSelector
			}
			to = []netv1.NetworkPolicyPeer{npPeer}
			npPorts := []netv1.NetworkPolicyPort{}
			for _, port := range service.Spec.Ports {
				protocol := port.Protocol
				targetPort := port.TargetPort
				npPorts = append(npPorts, netv1.NetworkPolicyPort{Protocol: &protocol, Port: &targetPort})
			}
			egressRules = append(egressRules, netv1.NetworkPolicyEgressRule{To: to, Ports: npPorts})
			// we add the service IP too
			// continue
		}
		// 3. deal with external service names
		ips, err := net.LookupIP(hostName)
		if err != nil {
			// TODO
			log.Err(err).Msgf("cannot get IP addresses for %s", hostName)
			continue
		}
		to = []netv1.NetworkPolicyPeer{}
		for _, ip := range ips {
			var ipBlock netv1.IPBlock
			if ipv4 := ip.To4(); ipv4 != nil {
				ipBlock = netv1.IPBlock{CIDR: ip.String() + "/32"}
			} else {
				ipBlock = netv1.IPBlock{CIDR: ip.String() + "/64"}
			}
			to = append(to, netv1.NetworkPolicyPeer{IPBlock: &ipBlock})
		}
		policyPort := netv1.NetworkPolicyPort{Protocol: &tcp}
		portString := url.Port()
		if portString == "" {
			// TODO: add default ports

		} else {
			portInt, err := strconv.Atoi(portString)
			if err != nil {
				log.Err(err).Msgf("cannot convert port %s", portInt)
			} else {
				port := intstr.FromInt(portInt)
				policyPort.Port = &port
			}
		}
		egressRules = append(egressRules, netv1.NetworkPolicyEgressRule{To: to, Ports: []netv1.NetworkPolicyPort{policyPort}})
	}
	egressRules = append(egressRules, dnsEngressRules)
	return egressRules, nil
}

func (r *BlueprintReconciler) cleanupNetworkPolicies(ctx context.Context, blueprint *fapp.Blueprint) error {
	l := client.MatchingLabels{}
	l[managerUtils.ApplicationNameLabel] = blueprint.Labels[managerUtils.ApplicationNameLabel]
	l[managerUtils.ApplicationNamespaceLabel] = blueprint.Labels[managerUtils.ApplicationNamespaceLabel]
	l[managerUtils.BlueprintNameLabel] = blueprint.Name
	l[managerUtils.BlueprintNamespaceLabel] = blueprint.Namespace
	r.Log.Trace().Str(logging.ACTION, logging.DELETE).Msgf("Delete Network Policies with labels %v", l)
	if err := r.Client.DeleteAllOf(ctx, &netv1.NetworkPolicy{}, client.InNamespace(environment.GetDefaultModulesNamespace()),
		l); err != nil {
		log.Error().Err(err).Msg("Error while deleting Network Policies")
		return err
	}
	r.Log.Trace().Str(logging.ACTION, logging.DELETE).Msg("Network Polices were deleted")
	return nil
}
