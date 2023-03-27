// Copyright 2023 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"

	"emperror.dev/errors"
	"github.com/rs/zerolog"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	fapp "fybrik.io/fybrik/manager/apis/app/v1beta1"
	managerUtils "fybrik.io/fybrik/manager/controllers/utils"
	"fybrik.io/fybrik/pkg/environment"
)

// Defines the default Network Polices peer if no input was provided
// returning nil, means denying all ingress connections
// returning empty NetworkPolicyPeer array, means allowing all ingress connections
func getDefaultNetworkPolicyFrom() []netv1.NetworkPolicyPeer {
	return nil
}

func (r *BlueprintReconciler) createNetworkPolicies(ctx context.Context,
	blueprint *fapp.Blueprint, releaseName string, log *zerolog.Logger) error {
	np := netv1.NetworkPolicy{}
	np.Name = releaseName
	np.Namespace = blueprint.Spec.ModulesNamespace
	labelsMap := blueprint.Labels
	np.Labels = managerUtils.CopyFybrikLabels(labelsMap)
	podsSelector := meta.LabelSelector{}
	podsSelector.MatchLabels = map[string]string{managerUtils.KubernetesInstance: releaseName}
	np.Spec.PodSelector = podsSelector

	np.Spec.PolicyTypes = []netv1.PolicyType{ /* netv1.PolicyTypeEgress, */ netv1.PolicyTypeIngress}

	tcp := corev1.ProtocolTCP
	npPorts := []netv1.NetworkPolicyPort{{Protocol: &tcp}}
	var from []netv1.NetworkPolicyPeer
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

	np.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{From: from, Ports: npPorts}}

	res, err := ctrlutil.CreateOrUpdate(ctx, r.Client, &np, func() error { return nil })
	if err != nil {
		return errors.WithMessage(err, releaseName+": failed to create NetworkPolicy")
	}
	log.Trace().Msgf("Network Policy %s/%s was createdOrUpdated result: %s", np.Namespace, releaseName, res)
	return nil
}

func (r *BlueprintReconciler) cleanupNetworkPolicies(ctx context.Context, blueprint *fapp.Blueprint) error {
	l := client.MatchingLabels{}
	l[managerUtils.ApplicationNameLabel] = blueprint.Labels[managerUtils.ApplicationNameLabel]
	l[managerUtils.ApplicationNamespaceLabel] = blueprint.Labels[managerUtils.ApplicationNamespaceLabel]
	l[managerUtils.BlueprintNameLabel] = blueprint.Name
	l[managerUtils.BlueprintNamespaceLabel] = blueprint.Namespace

	if err := r.Client.DeleteAllOf(ctx, &netv1.NetworkPolicy{}, client.InNamespace(environment.GetDefaultModulesNamespace()),
		l); err != nil {
		r.Log.Error().Err(err).Msg("Error while deleting Network Policies")
		return err
	}
	return nil
}
