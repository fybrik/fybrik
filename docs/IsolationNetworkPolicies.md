# Isolation based on Kubernetes Network Policies

## Introduction

As described in the [Fybrik Network topologies and Requirements](FybrikNetworkTopologiesAndRequirements.md)  we have to 
restrict network traffic of the data plane components to only allowed connections. One of the methods that can be used is 
Kubernetes Network Policies. This document explains the feature. 

## Network Policies

Fybrik is a Kubernetes application, therefore, the simplest isolation method can be based on 
[Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) (NP).
NP allow to control traffic flow at the IP address or port level (OSI layer 3 or 4). NP can restrict incoming (ingress) 
and/or outgoing (egress) communications. 
The restrictions are based on combinations of Pods, Namespaces labels and IP blocks.

### Advantages:
- does not require any changes in module implementations, can be implemented in the Fybrik control plane only.
  - A single possible module definition extension is to declare in the module manifest external services that a module 
    needs to connect, see below. 
- can be combined with other isolation methods (future work).

### Disadvantages:
- a single Kubernetes cluster oriented, cross clusters restrictions can be based on IP blocks only, which significantly 
complicates the configuration. Therefore, later we will provide other isolation methods as an extension to NP.
- NP are implemented by the network plugin. Creating a NetworkPolicy resource without a controller - a Container Network 
Interface (CNI) that implements it will have no effect. For example, `Kind` deployments which by default use "kindnetd" 
do not support NP. In order to support NP, installation of another CNI, e.g. [Calico](https://github.com/projectcalico/calico)
is required. Below there are instruction how to install kind with calico networking.  [instructions](https://alexbrand.dev/post/creating-a-kind-cluster-with-calico-networking/). 
On the other hand, all K8s clusters in production deployments use CNIs which support NP.

<details> 
<summary>Install kind with calico networking</summary>
    
The instructions are based on Alexander Brand's [blog](https://alexbrand.dev/post/creating-a-kind-cluster-with-calico-networking/)

Kind has a default Container Networking Interface (CNI) plugin called `kindnet`, which is a minimal implementation of a CNI plugin.
To use Calico as the CNI plugin in Kind clusters, we need to do the following:
1. Disable the installation of `kindnet`
To do so, create a `kind-calico.yaml` file that contains the following:
```yaml
kind: Cluster
apiVersion: kind.sigs.k8s.io/v1alpha4
networking:
  disableDefaultCNI: true # disable kindnet
  podSubnet: 192.168.0.0/16 # set to Calico's default subnet
```
_Note:_ you can use the file from ./manager/testdata

3. Create your Kind cluster, passing the configuration file using the --config flag:
```bash
kind create cluster --config ./manager/testdata/kind-calico.yaml
```

3. Verify Kind Cluster
Once the cluster is up, list the pods in the kube-system namespace to verify that `kindnet` is not running:
```bash
export KUBECONFIG="$(kind get kubeconfig-path --name="kind")"
kubectl get pods -n kube-system
```
`kindnet` should be missing from the list of pods:
```bash 
NAME                                         READY   STATUS    RESTARTS   AGE
coredns-5c98db65d4-dgfs9                     0/1     Pending   0          77s
coredns-5c98db65d4-gg4fh                     0/1     Pending   0          77s
etcd-kind-control-plane                      1/1     Running   0          16s
kube-apiserver-kind-control-plane            1/1     Running   0          24s
kube-controller-manager-kind-control-plane   1/1     Running   0          41s
kube-proxy-qsxp4                             1/1     Running   0          77s
kube-scheduler-kind-control-plane            1/1     Running   0          10s
```
_Note:_ The coredns pods are in the pending state. This is expected. They will remain in the pending state until a CNI 
plugin is installed.

4. Install Calico
Use the following command to install Calico:
```bash
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/calico.yaml
```
5. Verify Calico Is Up
To verify that calico-node is running, list the pods in the kube-system namespace:
```bash
kubectl -n kube-system get pods | grep calico-node
```
You should see the calico-node pod running and ready (1/1 containers ready):
```bash
calico-node-v5k5z                            1/1     Running   0          11s
```
You should also see the CoreDNS pods running if you get a full listing of pods in the kube-system namespace.

</details>

## Implementation

The `from`/`to` element of `ingress`/`egress`  
[NetworkPolicySpec](https://kubernetes.io/docs/reference/kubernetes-api/policy-resources/network-policy-v1/#NetworkPolicySpec) 
allows to restrict incoming/outgoing network traffic based on `podSelectors` (labels based), and/or 
`namespaceSelectors` (labels based) and/or `ipBlocks`. The last parameter can be used to restrict cross cluster connections 

We suggest, before deploying a data plane module, creating an instance of a network policy that will restrict the 
module connectivity.

### Backward compatibility and Network Policies isolation

Addition of NP isolation can change the existing Fybrik behavior. In order to support backward compatibility, we suggest
adding NP as an optional feature. A new entry `global.isNPEnabled` will be added into the Fybrik chart `values.yaml`.
NP will be evoked only if this value is set to `true`.

### FybrikApplication extensions

`FybrikApplicationSpec` has the “[selector](https://fybrik.io/v1.2/reference/crds/#fybrikapplicationspecselector)” 
element, which is a combination of `clusterName` and `workloadSelector`. Unfortunately, based on this information we can 
implement only one label-based `podSelector`, which will allow incoming connections from pods in the namespace where the NP 
instance is deployed and have the defined label. In order to restrict incoming connections to Fybrik modules, the NP
instances should be deployed in the same namespace where the modules are running. However, Fybrik separates user's 
objects, e.g. workloads and Fybrik objects, e.g. modules. Therefore, `FybrikApplicationSpec.selector` should be extended.

#### Suggestions

Extend the `FybrikApplicationSpec.selector` element with an optional array of namespaces names, and an optional blocks
of IPs. These extensions will not break CRDs backward compatability.

The `Selector` will have the following structure: 

```go
type Selector struct {
	// Cluster name 
	// +optional 
	ClusterName string `json:"clusterName"`
	
	// WorkloadSelector enables to connect the resource to the application 
	// Applications labels should match the labels in the selector. 
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
- if `Namespaces` is empty, a [NetworkPolicyPeer](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#networkpolicypeer-v1-networking-k8s-io)
  with `podSelector` equals to `WorkloadSelector` will be created. Which means that only pods with the defined labels from 
  the module namespace will be able to connect, or be connected.
- If `Namespaces` is not empty, for each entry in the array will be created a `NetworkPolicyPeer`
   with `podSelector` equals to `WorkloadSelector` and a `namespaceSelector` contains a single `kubernetes.io/metadata.name`
   label equals to the entry from `Namespaces`. [Here](https://kubernetes.io/docs/concepts/services-networking/network-policies/#targeting-a-namespace-by-its-name) 
   you can read more about targeting a Namespace by its name. If `Namespaces` is not empty but `WorkloadSelector` is empty, 
   then any pod from the selected namespaces will be able to connect.

- If `IPBlocks` is not empty, a separate NetworkPolicyPeer will be created for each IPBlocks entry.

### FybrikModule extensions

Some of Fybrik modules need external services for their functionality. For example, a currency exchange module checks 
the currency rates form an external service.
We suggest adding an `externalServices` array into the `FybrikModule.Spec`. Each entry will be a DNS name or IP address of 
the external service. During `egress` NetworkPolicyPeer creations, each service name will be translated to a single IPBlock entry.

### PlotterController extensions

In order to create ingress policies for modules, the blueprint that creates NPs should know potential clients, therefore
PlotterController who has the entire picture, should provide this information. 

NetworkPolicies select pods according to labels. In order to separate policies for 2 different modules, each module should
have a unique label, in additional to the common, fybrik application defined labels.

### BlueprintController extensions

BlueprintController installs and uninstalls Fybrik modules. We suggest extending this functionality by creating an instance
of NP for each installing module, and delete it when the module is uninstalled.

For the data plane entry module, the ingress element will be created based on information from a 
relevant `FybrikApplication`. For other modules this information will be provided by the `PlotterController`
Fybrik will not support different network restrictions to different module ports. (NP allow it)

The egress element will be a combination of destinations, such as, the next module (`2a`), data source (`5`), Fybrik (`3`) 
or module required (`4`) services. The connection types are taken from the [network topologies](FybrikNetworkTopologiesAndRequirements.md)


