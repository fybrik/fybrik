# Taxonomy

Fybrik interacts with multiple external components, such as the data catalog, data governance policy manager, and modules.  In order for fybrik to orchestrate the data plane of a given workload it is essential that all the components involved use common terms.  For example, if a data governance policy refers to a particular transform it is crucial that the module implementing that transform refer to it in the same way, or for fybrik to be able to map between the disparate terms.

A taxonomy defines the terms and related values that need to be commonly understood and supported across the components in the system:

FybrikApplication yaml - information provided about the workload and the datasets
Fybrik manager
Data catalog
Data Governance Policy Manager
Config Policy Manager
FybrikModules

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

As new capabilities, transforms, data types, and protocols are made available via FybrikModules fybrik's module taxonomy must be updated.  Once updated these capabilities are available for use by other components, such as the Data Catalog and Data Governance Policy manager should they choose to leverage them.

Default taxonomies are provided by fybrik, and are meant as a starting point on which to expand.

## Validation Points

Fybrik validates the structures and values it receives from all external components.

_Add diagram_

## Summary

The taxonomy mechanism enables independent components to work together, without the need for version updates and redeployment as capabilities change.
