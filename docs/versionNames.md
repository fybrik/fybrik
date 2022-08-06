
## Fybrik Version Policy
Fybrik uses [semantic versioning model](https://semver.org), which is the module adopted by a wide range of projects,
e.g. Kubernetes, goLang, Helm and many others.

Fybrik versions are of the form "vX.Y.Z" where X is the major version number, Y is the minor version number 
and Z is the patch release number. 
- Major Releases - released less frequently, indicate that **backward-incompatible public API changes**. This release 
carries no guarantee that it will be backward compatible with preceding major versions.
- Minor Releases - indicate **backward-compatible public API changes**. This release guarantees backward compatibility 
and stability.
- Patch Releases - should not change main functionality, mostly for security or bug fixes, and therefore guarantee 
backward compatibility.

*Note:* Kubernetes uses similar semantic version model, but its versions do not have the "v" prefix. In order to use 
the same release tag and goLang module tag, we have chosen goLang releases modules pattern with the "v" prefix.

## Kubernetes Custom Resources versions
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

##  Relations between CRD changes and Fybrik releases
Based on the above definitions, the relation between CRD versions and Fybrik versions is unidirectional. Which means, 
CRD API changes requires at least minor Fybrik release, on the other hand not each Fybrik release requires a new CRD 
version. 