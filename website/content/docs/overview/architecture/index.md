---
title: "Architecture"
date: 2020-04-30T23:10:46+03:00
draft: false
weight: 10
---

{{< warning >}}
The information below is outdated
{{</ warning >}}

{{< name >}} takes a **modular** approach to provide an open platform for controlling and securing the use of data across an organization. The figure below showcases the current architecture of the {{< name >}} platform, running on top of OpenShift Container Platform. The storage systems shown in the lower half of the figure are merely an example.


![Architecture](architecture.png)

{{< bullet n=1 >}} A {{< name >}} operator configures the resources available to {{< name >}}: compute and storage resources used to run governance actions and optimize data access performace, and `M4DModule` resources that describe modules injectable into the data path.

{{< bullet n=2 >}} A data streward configures policies in an external policy manager over assets defined in an external data catalog. The policy manager and data catalog are connected to {{< name >}} `PolicyCompiler` and `DataCatalog` services via `PolicyManagerAdapter` and `DataCatalogAdapter`, respectively.

{{< bullet n=3 >}} A developer submits a `M4DApplication` resource (e.g., via CI/CD pipeline) holding metadata about the application. such metadata includes the data assets required by the application and the processing purpose.

{{< bullet n=4 >}} The _pilot_, a core component of {{< name >}}, processes the submission and retrieves a set of all governance actions required to be enforced according to policies. This is also performed upon any policy change.

{{< bullet n=5 >}} The pilot retrives information about the resources and modules that it can use. Specifically, it searches for modules that can enforce the required governance actions without impacting the application in terms of performance and used client SDKs.

{{< bullet n=6 >}} The pilot generates a `Blueprint` describing the entire data path, including components injected into the data ingress and egress paths of the application.

{{< bullet n=7 >}} The _orchestrator_, a core component of {{< name >}}, processes the blueprint and deploys all runtime components accordingly. In the current release the deployment is to the same OpenShift Container Platform that the {{< name >}} control plane runs in.

{{< bullet n=8 >}} Processing jobs for preparing the environment for the application are running. The figure shows an example of creating an implicit copy job to copy data from PostgresSQL to Ceph.

{{< bullet n=9 >}} The application reads from and writes data to whitelisted targets. Requests are handled by `M4DModule` instances. The application can not interact with non-whitelisted targets.
