---
title: Control Plane Security
linktitle: Control Plane Security
weight: 30
---

Upon {{< name >}} deployment [Kubernetes NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) resource is applied to ensure that the connector ingress is limited to traffic only from workloads within the control plane namespace (called 4d-system be default).
In addition, Istio authentication policy can be applied upon installation of {{< name >}} to ensure mutual TLS between the pilot and connectors. 
Istio resources are not part of the default {{< name >}} deployment but it can be applied by running the
following command (assuming Istio version 1.6.1 and above is installed on the cluster):

```bash
make -C manager deploy_control_plane_security

```

In order for the Istio resouces to take effect the pods in the control plane should be restarted after running the make command.

For more details on the manager deploy\_control\_plane\_security configuration please see the manager README [config section](https://{{< github_base >}}/{{< github_repo >}}/tree/master/manager#config).


## Kubernetes cluster installation notes

[Network Plugins](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/) are prerequisites for the Kubernetes NetworkPolicy resource to take affect. Please see Kubernetes [prerequisites](https://kubernetes.io/docs/concepts/services-networking/network-policies/#prerequisites) for more details.

## Istio installation notes

If control plane security configuration is used the Envoy sidecar is automatically injected to pods in the control plane namespace. To support automatic injection Istio's [MutatingAdmissionWebhook](https://istio.io/latest/docs/setup/additional-setup/sidecar-injection/#automatic-sidecar-injection) should be turned on in the cluster.

## Steps to secure the ingress traffic of a new connector

Upon {{< name >}} deployment a new NetworkPolicy CRD is applied. This policiy
allows ingress traffic to any connector labeled ```m4d.ibm.com/componentType: connector``` only from workloads within the control plane namespace.

Given that, a new label:```m4d.ibm.com/componentType: connector``` should be added to the new connector pods.
