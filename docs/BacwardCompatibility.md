# Compatibility Requirements and Implementation Plan

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

For the Fybrik versioning convention see **TODO**

### Backward compatibility policy
Fybrik will support up to 2 backward minor releases, e.g. the version v1.5.x should be able to work with releases v1.4.x
and v1.3.x

## Fybrik components or features that can influence on compatability 

*Note:* the list is not closed, and can be extended

- Custom Resource Definitions (CRDs)
- Data Catalog Connector Interface
- Policy Manager Interface
- IT Configuration Policies
- IT attributes
- ....


## Custom Resource Definitions Upgrade Plan
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

## Data Catalog and Policy Manager Interfaces
Both data catalog and policy manager connectors use REST APIs. We do not expect significant API changes. Most of the 
possible changes are covered by the Taxonomy. However, in order to cover possible changes, we need to versioning them.   
Unfortunately, currently these interfaces are not versioned.

We can extend them with a standard REST versioning, 

## References
1. Kubernetes internal, "[Changing the API](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md#changing-the-api)"
requirements<a name="1"></a>
2. Kubernetes "[Versions in CustomResourceDefinitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/)"<a name="2"></a>
3. Kubernetes "[Serving multiple versions of a CRD](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#serving-multiple-versions-of-a-crd)"<a name="3"></a>
4. Kubebuilder: "[Multi-Version API](https://book.kubebuilder.io/multiversion-tutorial/tutorial.html)"<a name="4"></a>
5. "[Backward vs. Forward Compatibility](https://stevenheidel.medium.com/backward-vs-forward-compatibility-9c03c3db15c9)"<a name="5"></a>
