---
title: Istio security policies deployment
linktitle: Istio security policies deployment
weight: 30
---

## Steps to secure the Ingress Traffic of a New Connector

1. Inject the Envoy sidecar to the connector.
1. Create a service account for the connector identity. This is only needed if the new connector is in the ingress traffic of other connectors. In that case the AuthorizationPolicy CRDs of the other connectors should be updated with the new connector identity. Please see [Istio principles](https://Istio.io/latest/docs/concepts/security/#principals) for more information.
1. Deploy a new AuthorizationPolicy CRD to control the connector ingress traffic. Please see [AuthorizationPolicy CRD example](https://{{< github_base >}}/{{< github_repo >}}/blob/master/connectors/helpers/base/Istio/egr-connector-authorization.yaml) for more details. 
