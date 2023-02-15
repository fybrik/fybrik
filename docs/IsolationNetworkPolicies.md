# Isolation 

_**Note:**_ below we start with Network Polices as an isolation method, later we will add other methods.

_TODO_: add links to other options.

## Introduction

In the current Fybrik implementation the entry point of a data plane is published as a part of the `FybrikApplication` 
status, it is completely unprotected, and as result everyone can connect to the data plane. If a Fybrik data plane 
consist of a module chain, the same vulnerability exists for the chained modules.  
In `Read` flow cases it leads to governance/security violations, `Write` and `Delete` flows can create severe accidents, 
such as unauthorized data changes or deletions. Given that we have to provide a mechanism, which will prevent 
unauthorized data plane access.

## Challenge

When we talk about Fybrik data plane isolation, we are talking about connectivity between a user workload, which is out of
Fybrik control, and Fybrik modules, which can be developed by third parties. Therefore, in this document (and in its
extensions) we provide our recommendations how to protect data plane access and possible solutions. Responsibility of the
data plane isolation is shared between Fybrik core, Fybrik users, module developers and probably IT teams.

- Fybrik users should provide Fybrik with the information about possible workload locations and configure their workloads.
For example, to set labels on the workload pods and inform Fybrik about the labels and namespace 
where the workload runs.

- Depending on the isolation method, Fybrik should propagate all the relevant information to the deployed modules, and/or
correctly configure a chosen protection mechanism, e.g. Network Policies..

- Module developers should allow protection of the module instances. For example, to allow TLS support 
 
- In some complicated cases, IT teams intervention can be required, e.g. to configure external access points, firewalls and so on.

## Requirements

- Only predefined users/workloads should be able to access the data plane.
- If a data plane consist of a module chain we should prevent unauthorized access to the intermediate modules too.
- Modules should not be able to "illegally" communicate with modules outside of the defined data plane(s) and should 
connect to only predefined data sets.
- Currently, Fybrik assumes co-location of client workloads with data plane entry point modules in the same cluster. 
Future implementations might support deployment of workloads on different clusters or run them out of Kubernetes as 
standalone applications.
  - We check the co-location of a Fybrik module and a relevant workload in the same cluster as a specific use case.
- We should allow workload-client isolation based on IP addresses. It can be useful if a workload is not 
co-located with a Fybrik data plane in the same cluster  

## Network Policies

Fybrik is a Kubernetes application, therefore, the simplest isolation method can be based on 
[Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) (NP).
NP allow to control traffic flow at the IP address or port level (OSI layer 3 or 4). NP can restrict incoming (ingress) 
and/or outgoing (egress) communications. 
The restrictions are based on combinations of Pods and Namespaces labels and IP blocks.

### Advantages:
- does not require any changes in module implementations, can be implemented in the Fybrik control plane only.
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

The `from` element of `ingress` 
[NetworkPolicySpec](https://kubernetes.io/docs/reference/kubernetes-api/policy-resources/network-policy-v1/#NetworkPolicySpec) 
allows to restrict incoming network traffic based on `podSelectors` (labels based), and/or 
`namespaceSelectors` (labels based) and/or `ipBlocks`. The last one permits incoming network traffic from predefined IP 
address blocks. 
The `to` element of `egress` restricts outgoing connections and has similar structure. 

Current `FybrikApplicationSpec` has the “[selector](https://fybrik.io/v1.2/reference/crds/#fybrikapplicationspecselector)” 
element, which is a combination of `clusterName` and `workloadSelector`. Unfortunately, based on this information we can 
implement only one label-based `podSelector`, which will allow incoming connections from pods in the namespace where the NP 
instance is deployed and have the defined label. In order to restrict incoming connections to Fybrik modules, the NP
instances should be deployed in the same namespace where the modules are running. However, Fybrik separates user's 
objects, e.g. workloads and Fybrik objects, e.g. modules. Therefore, the existing mechanism should be extended.

### Suggestions:

- Extend the `FybrikApplicationSpec.selector` element with an array of `WorkloadLocations`, this array will be an input 
element for the Network Policy `from`element of the data plane entry module. 
- The `WorkloadLocation` wil have the following structure: 

```go
type WorkloadLocation struct {
    // This is a workload's pod label selector which selects user workload pods. This field follows standard label
    // selector semantics; if present but empty, it selects all pods.
    //
    // If NamespaceSelector is also set, then the WorkloadLocation as a whole selects
    // the workload pods matching WorkloadPodSelector in the namespaces selected by NamespaceSelector.
    // Otherwise, it selects the pods matching WorkloadPodSelector in the Fybrik data plane entry point namespace.
    // +optional
	WorkloadPodSelector *metav1.LabelSelector `json:"workloadPodSelector,omitempty" protobuf:"bytes,1,opt,name=workloadPodSelector"`

    // Selects namespaces using cluster-scoped labels. This field follows standard label
    // selector semantics; if present but empty, it selects all namespaces.
    //
    // If WorkloadPodSelector is also set, then the WorkloadLocation as a whole selects
    // the workload pods matching WorkloadLocation in the namespaces selected by NamespaceSelector.
    // Otherwise, it selects all Pods in the namespaces selected by NamespaceSelector.
	// IF you want to select a specific namespace based on its name, it can be done by selecting an immutable label 
	// `kubernetes.io/metadata.name` which is automatically set by the Kubernetes control plane. The value of the label 
	// is the namespace name.
    // +optional
    NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty" protobuf:"bytes,2,opt,name=namespaceSelector"`

    // IPBlock defines policy on a particular IPBlock. If this field is set then neither of the other fields can be.
	// the structure of the IPBlock is defined at https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#ipblock-v1-networking-k8s-io
    // +optional
    IPBlock *IPBlock `json:"ipBlock,omitempty" protobuf:"bytes,3,rep,name=ipBlock"`
} 
```
- A Plotter and relevant blueprints will provide the ingress/egress data for chain connected modules.

- Deprecate the `workloadSelector` element, but continue to support it for backward compatibility. Process it as an 
additional `workloadPodSelector` entry in the `WorkloadLocation` array, in a combination with namespace where a 
`FybrikApplication` is deployed. For example, if a `FybrikApplication` is deployed in the `fybrik-notebook-sample` 
namespace and a `workloadSelector` defines that a workload should be labeled as `app: my-notebook`, this will be equivalent to  
```yaml
workloadLocation:
  workloadPodSelector:
    matchLabels:
      app: my-notebook
  namespaceSelector:
    matchLabels:
      kubernetes.io/metadata.name: fybrik-notebook-sample
```
In other words, the following two `FybrikApplications` are equals
```yaml
apiVersion: app.fybrik.io/v1beta1
kind: FybrikApplication
metadata:
  name: my-notebook
  namespace: fybrik-notebook-sample
  labels:
    app: my-notebook
spec:
  selector:
    workloadSelector:
      matchLabels:
        app: my-notebook
...
```
```yaml
apiVersion: app.fybrik.io/v1beta1
kind: FybrikApplication
metadata:
  name: my-notebook
  namespace: fybrik-notebook-sample
  labels:
    app: my-notebook
spec:
  selector:
    workloadLocations:
      - workloadPodSelector:
          matchLabels:
            app: my-notebook
        namespaceSelector:
          matchLabels:
            kubernetes.io/metadata.name: fybrik-notebook-sample 
...
```
- If a user wants to run his workloads in different location, he can define several `WorkloadLocation`, they all will be
processed as logical OR, similar to NP.
- Terminate the backward compatibility assumption that all `Read` flows should have at least a single workload label. 
It was deprecated several Fybrik versions ago.

_Note_:  should be validated with our customers.

- If a `WorkloadLocation` array and a `workloadSelector` are empty, no Network Policies will be created, which is 
similar to current implementation. Probably other isolation methods are used.

- By modifying the `FybrikApplication.spec.selector` allow users to modify the NP without modification/redeployment of 
the data plane.

_Note_: we have to check if it is possible.    

- Extend blueprint’s module install/uninstall process by installation/removing of relevant NP instances.

- Fybrik will not support different network restrictions to different module ports. (NP allow it)

- If a user expects to run workload outside Fybrik clusters (it requires additional networking configuration) and the user
knows the IP address with which the workload will access the Fybrik data plane, the user might define the isolation 
restriction based on the IP address or a range of IP addresses.
```yaml
apiVersion: app.fybrik.io/v1beta1
kind: FybrikApplication
metadata:
  name: my-notebook
  namespace: fybrik-notebook-sample
  labels:
    app: my-notebook
spec:
  selector:
    workloadLocations:
      - ipBlock:
          cidr: 167.45.35.23/32
...
```


