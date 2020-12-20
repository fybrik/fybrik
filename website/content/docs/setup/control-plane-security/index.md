---
title: Control Plane Security
linktitle: Control Plane Security
weight: 30
---

Kubernetes  [`NetworkPolicies`](https://kubernetes.io/docs/concepts/services-networking/network-policies/) and optionally [Istio](https://istio.io/) are used to protect components of the control plane. Specifically, traffic to connectors that run as part of the control plane must be secured. Follow this page to enable control plane security.

# Ingress traffic policy

The installation of  {{< name >}} applies a Kubernetes [`NetworkPolicy`](https://kubernetes.io/docs/concepts/services-networking/network-policies/) resource to the `m4d-system` namespace. This resource ensures that ingress traffic to connectors is only allowed from workloads that run in the `m4d-system` namespace and thus disallow access to connectors from other namespaces or external parties.

The `NetworkPolicy` is always created. However, your Kubernetes cluster must have a [Network Plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/) with `NetworkPolicy` support. Otherwise, `NetworkPolicy` resources will have no affect. While most Kubernetes distributions include a network plugin that enfoces network policies, some like [Kind](https://kind.sigs.k8s.io/) do not and require you to install a separate network plugin instead.

# Mutual TLS

If Istio is installed in the cluster then you can use [automatic mutual TLS](https://istio.io/latest/docs/tasks/security/authentication/authn-policy/#auto-mutual-tls) to encrypt the traffic to the connectors.
Before you begin ensure that Istio 1.6 or above is installed.

Before installing {{< name >}} run the following:
```bash
make -C manager deploy_control_plane_security
```

If you already installed  {{< name >}} before running the above command then you must restart existing pods:
```bash
kubectl delete pod --all -n m4d-system
```

