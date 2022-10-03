# Taxonomy

Fybrik interacts with multiple external components, such as the data catalog, data governance policy manager, and modules.  In order for fybrik to orchestrate the data plane of a given workload it is essential that all the components involved use common terms.  For example, if a data governance policy refers to a particular transform it is crucial that the module implementing that transform refer to it in the same way, or for fybrik to be able to map between the disparate terms.

A taxonomy defines the terms and related values that need to be commonly understood and supported across the components in the system.

some components that use taxonomy:

* [FybrikApplication](../architecture/#fybrikapplication) yaml - information provided about the workload and the datasets.
* Fybrik manager ([FybrikApplication controller](../architecture/#fybrikapplication)) - validates that the data is used in accord with the data governance policies and the IT config policies.
* [Data catalog](../connectors/#data-catalog) - provides metadata about the asset.
* [Data Governance Policy Manager](../connectors/#policy-manager) - defines the governance policies to follow.
* [Config Policy Manager](./config-policies.md) - defines the IT policies to follow.
* [FybrikModules](./modules.md) - describes capabilities that can be included in a data plane.

Default taxonomies are provided by fybrik in a JSON file format, and are and are meant as a starting point on which to [expand](../../tasks/custom-taxonomy).

## Issues Addressed by Taxonomy

The taxonomy addresses the following:

* Redundancy: No need for the same structures and values to be hardcoded in multiple places, such as in the fybrik manager and in the plugins.
* Validation: Validates structures and values passed between components.
* Dynamic Updates: New terms and new values can be added, removed and updated dynamically. For example, one can add new enforcement actions, new connection types, new purposes, etc without needing to redeploy fybrik.

## Taxonomy Contributors

Different actors and components define the contents of different parts of the taxonomy.  The following table describes the taxonomy and which component most logically owns each part of it. 

| Taxonomy         | Contributing Component       | Actor              | Example Values                |
|------------------|------------------------------|--------------------|-------------------------------|
| catalog          | Data Catalog                 | Data Steward       | data stores, formats, metadata|
| application      | Policy Manager, Data Catalog | Governance Officer | roles, intents                |
| module           | Modules                      | Module Developer   | capabilities, transforms      |

If, for example, a Data Governance Officer writes a policy that limits the use of sensitive data for marketing, then the possible valid intents such as marketing would be defined by him in the Data Policy Manager.  These values must be added to fybrik's taxonomy, either manually or via an automated feed, so that fybrik can validate the intent provided in a FybrikApplication yaml when a user's workload requests data.

As new capabilities, transforms, data types, and protocols are made available via FybrikModules, fybrik's module taxonomy must be updated.  Once updated these capabilities are available for use by other components, such as the Data Catalog and Data Governance Policy manager should they choose to leverage them.

## Validation Points

Fybrik validates the structures and values it receives from all external components.  

For interface components (FybrikApplication and FybrikModule), validation occurs when the resource is created, updated or deleted.  How validation errors are received depends on whether fybrik is deployed with webhooks or not.

1. If webhooks are deployed, errors are received from the kubernetes command (ex: `kubectl apply` ) and no resource is created.  
2. If webhooks are *not* deployed, validation is done in the resource's controller.  If there is an error, the resource is created but its status will contain the error.  (Note: These resources will need to manually be removed by the person creating them.)


## Summary

The taxonomy mechanism enables independent components to work together, without the need for version updates and redeployment as capabilities change.
