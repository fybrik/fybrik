---
title: Istio post-installation steps
linktitle: Istio post-installation steps
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
allows ingress traffic of any connector labeld ```m4d.ibm.com/componentType: connector``` only from the pilot.

Given that, the following steps should apply to secure a new connector ingress traffic:

1. Add a new label:```m4d.ibm.com/componentType: connector``` to the components of the connectors (pods, service, deployment).
1. Create a service account for the connector identity. This is only needed if the new connector is in the ingress traffic of other connectors. In that case, the AuthorizationPolicy CRDs of the other connectors should be updated with the new connector identity. Please see [Istio principles](https://Istio.io/latest/docs/concepts/security/#principals) for more information.
1. Add AuthorizationPolicy CRD for the new connector if needed. This is only needed if the new connector has ingress traffic from another connector. Please see [AuthorizationPolicy CRD example](https://{{< github_base >}}/{{< github_repo >}}/blob/master/connectors/helpers/base/Istio/egr-connector-authorization.yaml) for more details.
