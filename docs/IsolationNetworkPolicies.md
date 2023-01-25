# Isolation 

_**Note:**_ below we start with Network Polices as an isolation method, later we will add another methods.

## Introduction

In the current Fybrik implementation the entry point of a data plane is published as a part of the FybrikApplication 
status, it is completely non-protected, and as result everyone can connect to the data plane. 
In “Read” flow cases it leads to governance/security violations, “Write” and “Delete” flows can create severe accidents, 
such as unauthorized data changes or deletions. Given that we need to prevent unauthorized data plane access. 

_Note:_ the same situation exists between chained modules.  

## Challenge

When we talk about accessing Fybrik data plane, we are talking about connectivity between a user workload, which is out of 
Fybrik scope, and Fybrik module, which can be developed by a third party. Therefore, in this document (and in its 
extensions) we provide our recommendations how to protect data plane access and possible solutions.Responsibility of the
data plane isolation is shared between Fybrik core, Fybrik users, module developers and probably IT teams.

- Fybrik users should provide all relevant information to Fybrik and correctly configure their workloads. For example, 
to set labels on the workload and inform Fybrik about the labels.

- Depends on the isolation method, Fybrik should propagate all relevant information to the deployed modules, and/or 
correctly configure Kubernetes Network Policies.
- Module developers should allow to protect the module instances. For example, to allow TLS support

In some complicated cases, IT teams interfere can be required, e.g. to configure external access points.

## Requirements

- Only predefined users/workloads should be able to access the data plane.
- We do not assume that client workloads will be collocated on the same Kubernetes cluster with the data module, 
furthermore they can run out of Kubernetes at all. 
  - However, we do check the collocation of a Fybrik module and a relevant workload in the same cluster as a specific use case. 
- We should allow workload client connections from different IP addresses.
- Isolation of modules in the data plane. I.e. Modules should not be able to interface "illegally" with modules outside
of the defined data plane(s) and connect to only predefined data sets.


## Network Policies

Fybrik is a Kubernetes application, therefore, the simplest isolation method can be based on 
[Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) (NP).
NP allow to control traffic flow at the IP address or port level (OSI layer 3 or 4). If a user workload or a Fybrik 
previous chained module collocated with the next module in the same Kubernetes cluster, NPs can help to restrict the Fybrik 
module ingress access. However, when a workload or a previous chained module runs out of the cluster, the NP based 
isolation is going to be more complicated. The collocated deployment can be based on Kubernetes labels or namespaces, but 
cross-clusters connections can be based on requester IP addresses, which is not always easy to discover. 
Therefore, later we will provide other isolation methods as an extension to NP.

### Advantages:
- does not require any changes in modules implementations, can be implemented in the Fybrik control plane only.
- can be combined with other isolation methods

### Disadvantages:
- complicated in cross clusters deployments.
- NP are implemented by the network plugin. Creating a NetworkPolicy resource without a controller - a Container Network 
Interface (CNI) that implements it will have no effect. Default Kind deployment which by default uses "kindnetd” doesn’t 
support NP. It requires installation another CNI, e.g. [Calico](https://github.com/projectcalico/calico). Here is the 
[instructions](https://alexbrand.dev/post/creating-a-kind-cluster-with-calico-networking/). All K8s clusters in 
production deployments use CNIs which support NP.
- requires information about workload deployment, and/or setting labels to the workload

## Implementation

The “from” element of ingress 
[NetworkPolicySpec](https://kubernetes.io/docs/reference/kubernetes-api/policy-resources/network-policy-v1/#NetworkPolicySpec) 
allows to restrict incoming network traffic based on “podSelectors” (labels based), and/or 
“namespaceSelectors” (labels based) and/or “ipBlocks”. The last one permits incoming network traffic from predefined IP 
address blocks.

Current “FybrikApplicationSpec” has the “[selector](https://fybrik.io/v1.2/reference/crds/#fybrikapplicationspecselector)” 
element, which is a combination of “clusterName” and “workloadSelector”. Unfortunately, based on this information we can 
implement only one label-based “podSelector”. Furthermore, for the backward compatibility of the current implementation 
assumes that a Read scenario should have at least one label, which is not always true. 

Similarly, the "egress.to" element restricts teh outgoing connections. In this way we can restrict destinations where a  
module can connect.

### Suggetions:

- extend the “FybrikApplicationSpec.selector” element with an array of 
[NetworkPolicyPeers](https://github.com/kubernetes/api/blob/59fcd23597fd090dba6b7e903eb0a8c9e8efb0a6/networking/v1/types.go#L183)
this “NetworkPoliciesPeers” array will be an input element for the NP `from`element. This will restrict possible incoming
connections to the relevant Fybrik module.
- a blueprint will provide the ingress input data for chain connected modules.

- deprecate the “workloadSelector” element, but continue to support it for backward compatibility. Process it as an 
additional “podSelector” entry in the “NetworkPoliciesPeers” array

- terminate the backward compatibility assumption that all Read flow should have at least a single workload label. 
It was deprecated several Fybrik versions ago.

- change the Flow field to be required instead of optional.

- if the NetworkPolicyPeers array and “workloadSelector” are empty, no Network Policies will be created, which is 
similar to current implementation.

- allow to user by modifying the "FybrikApplication.spec.selector" in the instance of its fybrikapplication to modify 
the NP without modification/redeployment of the data plane. *TODO:* we have to check if it is possible.    

- extend blueprint’s module install/uninstall process by installation of a relevant instance of network policies.

  - add the relevant ports to the NetworkPolicyIngressRules.
  - in future Fybrik releases, if the same module serves several workloads or other modules, we will create an 
independent NP instance per source. That will help us to manage the NP instances.

_Note:_ Fybrik will not support different network restrictions to different module ports.




