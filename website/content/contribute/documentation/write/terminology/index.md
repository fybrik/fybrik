---
title: Terminology Standards
description: Explains the terminology standards used in the documentation.
weight: 6
summary: Get familiar with the project terminology
---

{{< warning >}}
This page is not yet updated to include {{< name >}} specific terminology
{{< /warning >}}

To provide clarity to our users, use the standard terms in this section
consistently within the documentation.

## Service

Avoid using the term **service**. Research shows that different folks understand
different things under that term. The following table shows acceptable
alternatives that provide greater specificity and clarity to readers:

|Do                                          | Don't
|--------------------------------------------|-----------------------------------------
| Workload A sends a request to Workload B.  | Service A sends a request to Service B.
| New workload instances start when ...      | New service instances start when ...
| The application consists of two workloads. | The service consists of two services.

Our glossary establishes the agreed-upon terminology, and provides definitions to
avoid confusion.

## Envoy

We prefer to use "Envoy” as it’s a more concrete term than "proxy" and
resonates if used consistently throughout the docs.

Synonyms:

- "Envoy sidecar” - ok
- "Envoy proxy” - ok
- "The Istio proxy” -- best to avoid unless you’re talking about advanced
  scenarios where another proxy might be used.
- "Sidecar”  -- mostly restricted to conceptual docs
- "Proxy" -- only if context is obvious

## Miscellaneous

|Do              | Don't
|----------------|------
| addon          | `add-on`
| Bookinfo       | `BookInfo`, `bookinfo`
| certificate    | `cert`
| colocate       | `co-locate`
| configuration  | `config`
| delete         | `kill`
| Kubernetes     | `kubernetes`, `k8s`
| load balancing | `load-balancing`
| Mixer          | `mixer`
| multicluster   | `multi-cluster`
| mutual TLS     | `mtls`
| service mesh   | `Service Mesh`
| sidecar        | `side-car`, `Sidecar`
