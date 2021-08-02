# Connectors

The project currently has two extension mechanisms, namely connectors and modules. 
This page describes what connectors are and what connectors are installed using the default Fybrik installation.

## What are connectors?

Connectors are GRPC services that the Fybrik control plane uses to connect to external systems. Specifically, the control plane needs connectors to data catalog and policy manager. These connector GRPC services are deployed alongside the control plane.

## Can I write my own connectors?

Yes. Fybrik provides some default connectors described in this page but anyone can develop their own connectors.

A connector needs to implement one or more of the GRPC interfaces described in the [API documentation](../reference/connectors.md), depending on the connector type. Note that a single Kubernetes service can implement all GRPC interfaces if the system it connects to supports the required functionality, but it can also be different services.

In addition, to benefit from the [control plane security](../tasks/control-plane-security.md) feature ensure that the `Pods` of your connector:
1. Have a `fybrik.io/componentType: connector` label 
1. Have a `sidecar.istio.io/inject: "true"` annotation


## Connector types

### Data catalog

Fybrik assumes the use of an enterprise data catalog. For example, to reference a required data asset in a `FybrikApplication` resource, you provide a link to the asset in the catalog.

The catalog provides metadata about the asset such as security tags. It also provides connection information to describe how to connect to the data source to consume the data. Fybrik uses the metadata provided by the catalog both to enable seamless connectivity to the data and as input to making policy decisions. The data user is not concerned with any of it and just selects the data that it needs regardless of where the data resides.

Fybrik is not a data catalog. Instead, it links to existing data catalogs using connectors.
The default installation of Fybrik installs [Katalog](../reference/katalog.md), a built-in data catalog using Kubernetes CRDs used for evaluation. A connector to [ODPi Egeria](https://www.odpi.org/projects/egeria) is also available.

### Policy manager

Enforcing data governance policies requires a Policy Decision Point (PDP) that dictates what enforcement actions need to take place.
Fybrik supports a wide and extendible set of enforcement actions to perform on data read, write or copy. These include transformation of data, verification of the data, and various restrictions on the external activity of an application that can acceess the data.

A PDP returns a list of enforcement actions given a set of policies and specific context about the application and the data it uses. 
Fybrik includes a PDP that is powered by [Open Policy Agent](https://www.openpolicyagent.org/) (OPA). However, the PDP can also use external policy managers via connectors, to cover some or even all policy types. 

Policies are therefore defined externally in the policy manager of choice. Fybrik provides a package to help writing data policies in OPA. Otherwise, data stewards are expected to keep using the policy manager that they already use, as long as there is a connector to it.

