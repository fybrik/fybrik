# Compatability Requirements and Implementation Plan

## What is Compatability
Before talking about how to provide compatability, it is worthwhile to clarify what we mean by compatibility. 
Fybrik required to provide forwards and backwards compatibility in its behaviour. Compatibility is hard, especially 
handling issues around rollback-safety. 
Let's try to define compatability in this way:
- Existing clients need not be aware of a new change in order for them to continue to function as they did previously, 
even when the change is in use.
- Any Fybrik operations that succeeded before the new change must succeed after the change.
- Any Fybrik operations that does not use the new change must behave the same as it did before the change.
- Any Fybrik operations that use the new change must not cause problems (e.g. crash or degrade behavior) when issued 
against Fybrik objects, e.g. Modules, Plotters, Blueprints and Applications created before the functionality was added.
The objects should be smoothly promoted to the new version (if required).
- It must be possible to round-trip Fybrik version changes (convert to different API versions and back) with no loss of 
information (Should we support downgrade ?)

## Fybrik components or features that can influence on compatability 

*Note:* the least is not closed, and can be extended

- Custom Resource Definitions (CRDs)
- Data Catalog Connector Interface
- Policy Manager Interface
- Certificates and Secret Management
- IT Configuration Policies
- IT attributes
- Kubernetes Resources
- ....

## Fybrik Version Policy
Fybrik uses [semantic versioning model](https://semver.org), which is the module adopted by wide range of projects, 
e.g. Kubernetes, goLang, Helm and many others.
Fybrik version starts from the `v` character and 

Fybrik adapts GoLang [module version numbering](https://go.dev/doc/modules/version-numbers), which is very common 
policy used by many other open-source projects. The release names should be according to the semantic versioning model



## Kubernetes Resources
