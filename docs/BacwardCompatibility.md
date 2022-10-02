# Fybrik Versions, Compatibility Requirements and Implementation Plan

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
the same release tag and goLang module tag, we have chosen goLang releases modules pattern with the "v" prefix.

### Backward compatibility policy
Fybrik will support up to 2 backward minor releases, e.g. the version v1.5.x should be able to work with releases v1.4.x
and v1.3.x

## Fybrik components or features that can influence on compatability 

*Note:* the list is not closed, and can be extended

- [Kubernetes Custom Resources](#kubernetes-custom-resource-versions-and-upgrade-plan)
- [Data Catalog Connector Interface](#data-catalog-and-policy-manager-connector-interfaces)
- [Policy Manager Interface](#data-catalog-and-policy-manager-connector-interfaces)
- IT Configuration Policies
- IT attributes
- ....


## Kubernetes Custom Resource Versions and Upgrade Plan

### Versions
Fybrik for its Kubernetes API objects (CRDs) uses API versioning similar to one used by
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
they are stored at and the version they are served at. If the conversion involves schema changes and requires custom 
logic, a conversion **Webhook** should be used. If there are no schema changes, the default **None** conversion strategy may 
be used and only the apiVersion field will be modified when serving different versions."

Due to some Fybrik customers' limitations, we cannot use Webhooks in Fybrik deployments. Therefore, below we investigate 
possible CRD changes without changing its schema.

We were able to change a required field in version v1 to an optional field in v2. In addition, we were able to add a new 
optional field in v2. This is inline with compatibility examples from [\[5\]](#5)

*Note:* If we go from previous API version to a new one without supporting the previous version, let say `v1alpha1` and 
a cluster API server has stored objects of the previous version, we will get an error similar to the following:  
```
CustomResourceDefinition.apiextensions.k8s.io "plotters.app.fybrik.io" is invalid: status.storedVersions[0]: Invalid value: "v1alpha1": must appear in spec.versions
```

## Data Catalog and Policy Manager Connector Interfaces
Fybrik manager communicates with connectors (data catalog and policy manager) over REST API. 

REST doesn’t provide any specific versioning guidelines, however the more commonly used approaches fall into three 
categories:
- URI Versioning
- Versioning using Custom Request Header
- Versioning using the “Accept” header
  For examples and explanations, see [REST API Versioning](https://restfulapi.net/versioning/)
Traditionally, REST APIs do not the support semantic version model, each new version is a single number. 

For connectors REST API, we suggest using URI versioning as the most commonly used and easiest to check.
Which means, all REST requests to for example, version 3 will have URI prefix `/v3`

*Notes:* 
- The current implementation of controllers REST API does not support any versioning. When we will define a new API 
version we will add the URI prefix as the REST API versioning mechanisms and assume that missing version information is the
current (old) version.
- We do not expect significant API changes. Most of the possible changes are covered by the Taxonomy

## Other Components
Work in progress, TBD

## References
1. Kubernetes internal, "[Changing the API](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md#changing-the-api)"
requirements<a name="1"></a>
2. Kubernetes "[Versions in CustomResourceDefinitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/)"<a name="2"></a>
3. Kubernetes "[Serving multiple versions of a CRD](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#serving-multiple-versions-of-a-crd)"<a name="3"></a>
4. Kubebuilder: "[Multi-Version API](https://book.kubebuilder.io/multiversion-tutorial/tutorial.html)"<a name="4"></a>
5. "[Backward vs. Forward Compatibility](https://stevenheidel.medium.com/backward-vs-forward-compatibility-9c03c3db15c9)"<a name="5"></a>
6. [REST API Versioning](https://restfulapi.net/versioning/)
