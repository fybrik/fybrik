---
title: Istio post-installation steps
linktitle: Istio post-installation steps
weight: 30
---

## Istio post-installation steps

Envoy sidecar is automatically injected to pods in the control-plane namespace (called 4d-system be default). To ensure the cluster supports automatic injection Istio's [MutatingWebhook](https://istio.io/latest/docs/setup/additional-setup/sidecar-injection/#automatic-sidecar-injection) should be turned on in the cluster.

To check it MutatingWebhook is turned on in the cluster run:

```bash
kubeadm config print init-defaults
```

To turn it on in Kind run:

```bash
kubeadm init --config the-mesh-for-data/istio/admission_patch.yaml
```

## Steps to secure the Ingress Traffic of a New Connector

1. Inject the Envoy sidecar to the connector.
1. Create a service account for the connector identity. This is only needed if the new connector is in the ingress traffic of other connectors. In that case the AuthorizationPolicy CRDs of the other connectors should be updated with the new connector identity. Please see [Istio principles](https://Istio.io/latest/docs/concepts/security/#principals) for more information.
1. Deploy a new AuthorizationPolicy CRD to control the connector ingress traffic. Please see [AuthorizationPolicy CRD example](https://{{< github_base >}}/{{< github_repo >}}/blob/master/connectors/helpers/base/Istio/egr-connector-authorization.yaml) for more details. 
