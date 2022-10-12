# Fybrik Versions, Compatibility Requirements and Implementation Plan

## Disclaimer
This document was created according to the Fybrik release v1.1.0, future releases may require updates code pointers

## What is Compatibility
Before talking about how to provide compatability, it is worthwhile to clarify what we mean by compatibility. 
Fybrik is required to provide forwards and backwards compatibility in its behaviour. Compatibility is hard, especially 
handling issues around rollback-safety. 
Let's try to define compatability in this way:
- Existing clients need not be aware of a new change in order to continue to function as they did previously, 
even when the change is in use.
- Any Fybrik operation that succeeded before the new change must succeed after the change.
- Any Fybrik operation that does not use the new change must behave the same as it did before the change.
- Any Fybrik operation that use the new change must not cause problems (e.g. crash or degrade behavior) when issued 
against Fybrik objects, e.g. FybrikModules, Plotters, Blueprints and FybrikApplications created before the functionality was added.
The objects should be smoothly promoted to the new version (if required).
- It must be possible to round-trip Fybrik version changes (convert to different API versions and back) with no loss of 
information (Should we support downgrade ?)

*Note:* Fybrik allows to its users to [redefine most of the components](https://fybrik.io/v1.1/concepts/taxonomy/#taxonomy-contributors) 
by using a custom taxonomy. If any of the components were redefined, it will be the user responsibility to update its 
custom taxonomy files and use them during **each** 
fybrik [upgrade](https://fybrik.io/v1.1/tasks/custom-taxonomy/#deploy-fybrik-with-custom-taxonomy). 

## Fybrik Version Policy
Fybrik uses [semantic versioning model](https://semver.org), which is the module adopted by a wide range of projects,
e.g. Kubernetes, goLang, Helm and many others.

Fybrik versions are of the form "vX.Y.Z" where X is the major version number, Y is the minor version number
and Z is the patch release number.
- Major Releases - released less frequently, indicate that **backward-incompatible public API changes**. This release
  carries no guarantee that it will be backward compatible with preceding major versions.
- Minor Releases - indicate **backward-compatible public API changes**. This release guarantees backward compatibility
  and stability. The backward compatibility policy is defined in a separate document [ADD link] but in general, Fybrik
  supports 2 previous minor releases.
- Patch Releases - should not change main functionality, mostly for security or bug fixes, and therefore guarantee
  backward compatibility.

*Note:* Kubernetes uses similar semantic version model, but its versions do not have the "v" prefix. In order to use
the same release tag and goLang module tag, we have chosen goLang releases modules pattern with the "v" prefix. The same
version pattern will be used by all others, except CRDa, versioned Fybrik components. CRDs use its proprietary defined 
by Kubernetes versioning.   

### Backward compatibility policy
Fybrik will support up to 2 backward minor releases. For example, the version v1.5.x should be able to work with 
releases v1.4.x and v1.3.x, on another hand it might work with releases v1.2.x, but it is not guaranteed 

## Fybrik components or features that can influence compatability 

*Note:* the list is not closed, and can be extended

*Note:* some components, listed below, are external integration points, changes in them impacts not only on the backward 
compatibility, but on external development process. We emphasized these components with the label `integration point`.  

- Fybrik control plane [Kubernetes Custom Resources](#kubernetes-custom-resource-versions-and-upgrade-plan) 
`FybrikApplication` is a Fybrik entry point and `FybrikModule` is a definition for externally developed modules, 
therefore these CRDs are integration points. 
- [Connectors](#connectors): are Open API services that the Fybrik control plane uses to connect to external systems. 
These connector services are deployed alongside the Fybrik control plane, and as we can see from their name, all of them 
are external integration points
  - Data Catalog connector
  - Policy Manager connector
  - Credential Management connector
- [Default Taxonomy](#default-taxonomy). Terms and values defined by taxonomy are part of Connectors APIs and CRDs 
validations, therefore we can see taxonomy as part of external connectivity and an integration point.  
- [Default Policies](#default-policies)
  - [Data Access Policies](#default-data-access-policies)
  - [IT Configuration Policies](#default-it-config-policies)
- [Fybrik Logs format](#fybrik-logs-format). If Fybrik logs are used by external analytic tools, they will be part of 
external integration point too.

## Kubernetes Custom Resource Versions and Upgrade Plan

Current Fybrik [Custom Resource Definitions](https://github.com/fybrik/fybrik/tree/master/manager/apis/app/v1beta1) 

### Versions
Fybrik for its Kubernetes API objects (CRDs) uses API versioning like one used by
[Kubernetes](https://kubernetes.io/docs/reference/using-api/#api-versioning).
- Alpha
    - The version names contain alpha (for example, v1alpha1).
    - The software is not stable may contain bugs.
    - The support for a feature may be dropped at any time without notice.
    - The API may change in incompatible ways in a later software release without notice.
- Beta
    - The version names contain beta (for example, v2beta3).
    - The software is well tested. Enabling a feature is considered safe.
    - The support for a feature will not be dropped, though the details may change.
    - The schema and/or semantics of objects may change in incompatible ways in a subsequent beta or stable release.
      When this happens, migration instructions are provided. Schema changes may require deleting, editing, and re-creating
      API objects. The editing process may not be straightforward. The migration may require downtime for applications that
      rely on the feature.
- Stable
    - The version name is vX where X is an integer.
    - The stable versions of features appear in released software for many subsequent versions.
### Relations between CRD changes and Fybrik releases
Based on the above definitions, the relation between CRD versions and Fybrik versions is unidirectional. Which means,
CRD API changes requires at least minor Fybrik release, on the other hand not each Fybrik release requires a new CRD
version.

### Upgrade Plan
The CustomResourceDefinition API provides a workflow for introducing and upgrading to new versions of a 
CustomResourceDefinition. For more information about 
[serving multiple versions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#serving-multiple-versions-of-a-crd) of your CustomResourceDefinition 
and migrating your objects from one version to another see 
[Versions in CustomResourceDefinitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/)

The first step of adding a new version is to "pick a conversion strategy. Since custom resource objects need to be able 
to be served at both versions, that means they will sometimes be served at a different version than their storage 
version. In order for this to be possible, the custom resource objects must sometimes be converted between the version 
they are stored at and the version they are served at." If the conversion does not involve CRD schema changes, 
the default **None** conversion strategy may be used and only the apiVersion field will be modified when serving 
different versions.
If the conversion does change schema and requires custom logic, Kubernetes suggests using webhooks to implement this 
custom conversation, and setting the conversation strategy as **Webhook**. However, due to some Fybrik customers' 
limitations, we cannot use Webhooks in Fybrik deployments. Therefore, below we investigate
possible CRD changes without changing its schema. Adn in the future will check possible in-house solutions. 


#### CRD changes with None conversation strategy
When we use the None conversation strategy, we are able to change a required field in version v1 to an optional field in 
v2. In addition, we are able to add a new  optional field in v2. These changes are inline with compatibility examples 
from [\[5\]](#references)

*Note:* If we go from previous API version to a new one without supporting the previous version, let say `v1alpha1` and 
a cluster API server has stored objects of the previous version, we will get an error similar to the following:  
```
CustomResourceDefinition.apiextensions.k8s.io "plotters.app.fybrik.io" is invalid: status.storedVersions[0]: Invalid value: "v1alpha1": must appear in spec.versions
```
## Connectors
The Fybrik website [describes](https://fybrik.io/v1.1/concepts/connectors/) what are connectors and their types.
There are three different connectors types:
- [Data Catalog](https://github.com/fybrik/fybrik/blob/master/connectors/api/datacatalog.spec.yaml) connector
- [Policy Manager](https://github.com/fybrik/fybrik/blob/master/connectors/api/policymanager.spec.yaml) connector
- [Credential Management](https://github.com/fybrik/fybrik/blob/master/pkg/vault/vault_interface.go) connector 
(TODO add OpenAPI definition, see the issue [#1761](https://github.com/fybrik/fybrik/issues/1761) )

Fybrik manager communicates with connectors over REST APIs. 

REST does not provide any specific versioning guidelines, however the more commonly used approaches fall into three 
categories:
- URI Versioning
- Versioning using Custom Request Header
- Versioning using the “Accept” header
For examples and explanations, see [REST API Versioning](https://restfulapi.net/versioning/)

For connectors REST API, we suggest using URI versioning as the most commonly used and easiest to check.
Which means, all REST requests to for example, version 3 will have URI prefix `/v3`. Currently, Fybrik connectors do not
have the versioning prefix. We can assume that requests without the prefix relate to the v1 version. 

*Notes:* 
- The current implementation of controllers REST API does not support any versioning. The new incompatible versions will 
start from the "/v2" prefix, and we will assume that missing version information is the current (old) version.
- We do not expect significant API changes. Most of the possible changes are covered by the Taxonomy
- OpenApi [specification](https://swagger.io/specification/) includes `version` entry in the `Info`, however, the object 
> "provides metadata about the API. The metadata MAY be used by the clients if needed, and MAY be presented in editing or 
documentation generation tools for convenience." 

We will update the OpenAPI version values, but they cannot be used for API communications.

## Default Taxonomy
Fybrik widely uses [Taxonomy](https://fybrik.io/v1.1/concepts/taxonomy/) to defines the terms and related values that 
need to be commonly understood and supported across the components in the system. Taxonomy provides a mechanism for all 
Fybrik components to interact using a common dialect. Taxonomy defines a set of immutable structural JSON schemas, or 
"taxonomies" for resources deployed in Fybrik. However, since the taxonomy is meant to be configurable, a taxonomy.json 
file is referenced from these schemas for any definition that is customizable. The taxonomy.json file is generated from 
a base taxonomy and zero or more taxonomy layers:
- The base taxonomy is maintained by the project and includes all the structural definitions that are subject to 
customization (e.g.: tags, actions).
- The taxonomy layers are maintained by users and external systems that add customizations over the base taxonomy 
(e.g., defining specific tags, actions), therefore the layer are beyond Fybrik responsibility. 

The structural JSON schemas are autogenerated files in the 
[taxonomy](https://github.com/fybrik/fybrik/tree/master/charts/fybrik/files/taxonomy) direction.
- The `fybrik_application.json` and `fybrik_module.json` files are generated from corresponded CRDs, so their 
compatability is covered by the CRDs compatibility.
- The `datacatalog.json`, `infraatributes.json` and `policymanager.json` files are generated form files in the 
[pkg/model](https://github.com/fybrik/fybrik/tree/master/pkg/model), therefore any changes in the directory should be  
validated for backward compatibility.
- The basic [taxonomy file](https://github.com/fybrik/fybrik/blob/master/charts/fybrik/files/taxonomy/taxonomy.json), is
autogenerated from the both above sources. However, users are able to overwrite it, as describe by [Using a Custom 
Taxonomy for Resource Validation](https://fybrik.io/v1.1/tasks/custom-taxonomy/). 

Unfortunately all these files do not have version definition, and it can be complicated to separate files from different 
releases. In order to be able to compare the files, I suggest to add an entry `version` and modify the 
[json-schema-generator](https://github.com/fybrik/json-schema-generator) project to generate the entry. It's an optional 
entry, so we don't have to change the go structures, but will be able to validate the content of the 
`fybrik-taxonomy-config` configuration map. Due to possibility of user modifications of the basic taxonomy file we propose 
in addition to the version, add the data of the file generation, and maybe its CRC. 

*Note:* Should we create a taxonomy json file for the Credential Management connector?
*Note:* Should we provide a [default metric taxonomy](https://fybrik.io/v1.1/tasks/infrastructure/#add-a-new-attribute-definition-to-the-taxonomy)

## Default Policies
Fybrik is installed with a set of different default policies. Most of the change in the policies do not prevent Fybrik 
operation, but can change its default behaviour and therefore influence on its users.

### Default Data Access Policies
Fybrik can be deployed with different policy managers, for example it can be Open Policy Agent ([OPA](https://www.openpolicyagent.org/)), 
or some other agent. However, usually any deployment contains some default setting, which can be all allowed, or all denied, 
or other default policies settings. How it was set, depends on the used policy agent. When fybrik deployed with OPA, 
the [default OPA policy rules](https://github.com/fybrik/fybrik/blob/master/charts/fybrik/files/opa-server/policy-lib/default_policy.rego) 
deployed as well and define Fybrik behavior if no other rules are not provided by users.

### Default IT Config Policies 
[IT config policies](https://fybrik.io/v1.1/concepts/config-policies/) are the mechanism via which the organization may 
influence the construction of the data plane, taking into account infrastructure capabilities and costs.
Out of the box policies come with the fybrik deployment. They define the deployment of basic capabilities, 
such as read, write, copy and delete.
The default IT policies and the default infrastructure definitions located at 
[./charts/fybrik/adminconfig](https://github.com/fybrik/fybrik/tree/master/charts/fybrik/files/adminconfig)

*Note:* in order to update the IT config policies and define infrastructure, Fybrik 
[recommends](https://fybrik.io/v1.1/concepts/config-policies/#how-to-provide-custom-policies) to overwrite the files in 
the `./charts/fybrik/adminconfig` or direct changes in the `fybrik-adminconfig` configmap. However, fybrik updates will
overwrite these changes. We have to provide a different mechanism, which will automatically merge default and user defined 
settings.  

## Fybrik logs format
Fybrik, during its running produces log files. The files are output of Fybrik execution, so we don't have to define a 
separate version mechanism for the files format. They will be inline with the Fybrik versions. However, changing log 
[component types](https://github.com/fybrik/fybrik/blob/b19f114988fc318819ddb6fa42059bf7e473ae63/pkg/logging/logging.go#L38)
or log [entry parameters](https://github.com/fybrik/fybrik/blob/b19f114988fc318819ddb6fa42059bf7e473ae63/pkg/logging/logging.go#L60) 
can break log analytics tools.   
### Backward compatability strategy
- Include in the Fybrik logs its version, it will help to parse historical logs.
- Avoid changes in the log format
- Use a graceful period (2 minor versions) for deprecated log entries
- For renamed log entries, use both the old and the new definitions during the graceful period. 

## References
1. Kubernetes internal, "[Changing the API](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md#changing-the-api)"
requirements<a name="1"></a>
2. Kubernetes "[Versions in CustomResourceDefinitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/)"<a name="2"></a>
3. Kubernetes "[Serving multiple versions of a CRD](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#serving-multiple-versions-of-a-crd)"<a name="3"></a>
4. Kubebuilder: "[Multi-Version API](https://book.kubebuilder.io/multiversion-tutorial/tutorial.html)"<a name="4"></a>
5. "[Backward vs. Forward Compatibility](https://stevenheidel.medium.com/backward-vs-forward-compatibility-9c03c3db15c9)"<a name="5"></a>
6. [REST API Versioning](https://restfulapi.net/versioning/)<a name="6></a>
7. [OpenAPI Specification](https://spec.openapis.org/oas/v3.1.0)<a name="7"></a>
