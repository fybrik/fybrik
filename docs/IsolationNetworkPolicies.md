# Isolation based on Kubernetes Network Policies

## Introduction

As described in the [Fybrik Network topologies and Requirements](FybrikNetworkTopologiesAndRequirements.md)  we have to 
isolate network traffic of the data plane components to only allowed connections. One of the methods that can be used is 
Kubernetes Network Policies. This document explains the feature. 

## Network Policies

Fybrik is a Kubernetes application, therefore, the simplest isolation method can be based on 
[Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) (NP).
NP allow to control traffic flow at the IP address or port level (OSI layer 3 or 4). NP can restrict incoming (ingress) 
and/or outgoing (egress) communications. 
The restrictions are based on combinations of Pods, Namespaces labels and IP blocks.

### Advantages:
- does not require any changes in module implementations, can be implemented in the Fybrik control plane only.
  - A possible required change is to declare in the module manifest external services that a module needs to connect, see below. 
- can be combined with other isolation methods (future work).

### Disadvantages:
- a single Kubernetes cluster oriented, cross clusters restrictions can be based on IP blocks only, which significantly 
complicates the configuration. Therefore, later we will provide other isolation methods as an extension to NP.
- NP are implemented by the network plugin. Creating a NetworkPolicy resource without a controller - a Container Network 
Interface (CNI) that implements it will have no effect. Kind deployments which by default use "kindnetd" do not
support NP. Support of NP requires installation of another CNI, e.g. [Calico](https://github.com/projectcalico/calico). Here is the
[instructions](https://alexbrand.dev/post/creating-a-kind-cluster-with-calico-networking/). On the other hand, all K8s 
clusters in production deployments use CNIs which support NP.

## Implementation

The `from`/`to` element of `ingress`/`egress`  
[NetworkPolicySpec](https://kubernetes.io/docs/reference/kubernetes-api/policy-resources/network-policy-v1/#NetworkPolicySpec) 
allows to restrict incoming/outgoing network traffic based on `podSelectors` (labels based), and/or 
`namespaceSelectors` (labels based) and/or `ipBlocks`. The last parameter can be used for restriction cross cluster 
boundaries network connections 

We suggest in addition to deploy a data plane module, to create an instance of a network policy that will restrict the 
module connectivity.

### Backward compatibility and Network Policies isolation

Addition of NP isolation can change the existing Fybrik behavior. In order to support backward compatibility, we suggest
to add NP as an optional feature. A new entry `coordinator.isNPenable` will be added into the Fybrik chart `values.yaml`.
NP will be evoked only if this value is set to `true`.

### FybrikApplication extensions

Current `FybrikApplicationSpec` has the “[selector](https://fybrik.io/v1.2/reference/crds/#fybrikapplicationspecselector)” 
element, which is a combination of `clusterName` and `workloadSelector`. Unfortunately, based on this information we can 
implement only one label-based `podSelector`, which will allow incoming connections from pods in the namespace where the NP 
instance is deployed and have the defined label. In order to restrict incoming connections to Fybrik modules, the NP
instances should be deployed in the same namespace where the modules are running. However, Fybrik separates user's 
objects, e.g. workloads and Fybrik objects, e.g. modules. Therefore, the existing mechanism should be extended.

#### Suggestions

Extend the `FybrikApplicationSpec.selector` element with an optional array of namespaces names, and an optional blocks
of IPs. These extensions will not break CRD backward compatability.

The `Selector` will have the following structure: 

```go
type Selector struct {
    // Cluster name
    // +optional
    ClusterName string `json:"clusterName"`

    // WorkloadSelector enables to connect the resource to the application
    // Application labels should match the labels in the selector.
    // +required
    WorkloadSelector metav1.LabelSelector `json:"workloadSelector"`
	
	// Namespaces where user application might run
	// +optional
	Namespaces []string `json:"namespaces"`

	// IPBlocks define policy on particular IPBlocks.
	// the structure of the IPBlock is defined at https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#ipblock-v1-networking-k8s-io
	// +optional
	IPBlocks []*IPBlock `json:"ipBlocks,omitempty"`
}
```
The mapping the `Selector` elements into the `from` NP ingress definitions will be done according the following algorithm:
- `Namespaces` is not empty,for  each entry in the array will be created a 
   [NetworkPolicyPeer](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#networkpolicypeer-v1-networking-k8s-io)
   with `podSelector` equals to `WorkloadSelector` and a `namespaceSelector` contains a single `kubernetes.io/metadata.name`
   label equals to the entry from `Namespaces`. You can read more about 
   [Targeting a Namespace by its name](https://kubernetes.io/docs/concepts/services-networking/network-policies/#targeting-a-namespace-by-its-name) 

- If `IPBlocks` is not empty, a separate NetworkPolicyPeer will be created for each IPBlocks entry.

### FybrikModule extensions

Some of Fybrik modules need external services for their functionality. For example, a currency exchange module checks 
the currency rates form an external service.
We suggest to add `externalServices` array into the `FybrikModule.Spec`. Each entry will be a DNS name or IP address of 
the external service. During `egress` NetworkPolicyPeer creations, each service name will be translated to a single IPBlock entry.

### PlotterController extensions

In order to create ingress policies for modules, the blueprint that creates NPs should know potential clients, therefore
PlotterController who has the entire picture, should provide this information. 

NetworkPolicies select pods according to labels. In order to separate policies for 2 different modules, each module should
have a unique label, in additional to the common, fybrik application defined labels.

### BlueprintController extensions

BlueprintController install and uninstall Fybrik modules. We suggest to extend this functionality to create an instance of NP for 
each installing module, and delete it when the module is uninstalled.

The ingress entry for the module which provides an entry to a data plane, will be created based on information from a 
relevant `FybrikApplication`. For other modules this information will be provided by the `PlotterController`
Fybrik will not support different network restrictions to different module ports. (NP allow it)

The egress entry will be a combination of the next module, or data source; Fybrik services, such as Vault and module 
required services.



