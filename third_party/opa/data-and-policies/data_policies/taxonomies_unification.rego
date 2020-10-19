package data_policies

#In data part we provide set of general and industry specific taxonomies, also the user can add more taxonomies specific for his needs.
#Here is the place when for each category user chooses what taxonomies should be used

Purposes = { x | x = data.DataPurposes[_] }

Roles = { x | x = data.DataRoles[_] } | { x | x = data.MedicalRoles[_] }

Sensitivity = { x | x = data.DataSensitivity[_] }

AccessTypes = { x | x = data.DataAccessTypes[_] }

GeoDestinations = { x | x = data.DataGeoDestinations[_] }