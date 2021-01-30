package data_policies

#In data part we provide set of general and industry specific taxonomies, also the user can add more taxonomies specific for his needs.
#Here is the place when for each category user chooses what taxonomies should be used

Purposes = { x | x = data.opa["opa-json"]["taxonomies.json"].DataPurposes[_] }

Roles = { x | x = data.opa["opa-json"]["taxonomies.json"].DataRoles[_] } | { x | x = data.opa["opa-json"]["medical_taxonomies.json"].MedicalRoles[_] }

Sensitivity = { x | x = data.opa["opa-json"]["taxonomies.json"].DataSensitivity[_] }

AccessTypes = { x | x = data.opa["opa-json"]["taxonomies.json"].DataAccessTypes[_] }

GeoDestinations = { x | x = data.opa["opa-json"]["taxonomies.json"].DataGeoDestinations[_] }