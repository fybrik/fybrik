---
title: Istio Installation notes
linktitle: Istio Installation notes
weight: 30
---

## Turn on MutatingWebhook

Upon {{< name >}} deployment the Envoy sidecar is automatically injected to pods in the control-plane namespace (called 4d-system be default). To support automatic injection Istio's [MutatingAdmissionWebhook](https://istio.io/latest/docs/setup/additional-setup/sidecar-injection/#automatic-sidecar-injection) should be turned on in the cluster.

To check if MutatingAdmissionWebhook is turned on in the cluster run:

```bash
kube-apiserver -h | grep enable-admission-plugins
```

## Steps to secure the Ingress Traffic of a New Connector

Upon {{< name >}} deployment a new AuthorizationPolicy CRD is applied. This policiy
allows ingress traffic to any connector labeled ```m4d.ibm.com/componentType: connector``` only from workloads within the control-plane namespace.

Given that, a new label:```m4d.ibm.com/componentType: connector``` should be added to the components of a new connector (pods, service, deployment).
