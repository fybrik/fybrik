package data_policies

#In data part we provide set of general and industry specific taxonomies, also the user can add more taxonomies specific for his needs.
#Here is the place when for each category user chooses what taxonomies should be used

Intents = { x | x = data["m4d-system"]["meshfordata-external-data"]["taxonomies.json"].Intents[_] }

Roles = { x | x = data["m4d-system"]["meshfordata-external-data"]["taxonomies.json"].Roles[_] } | { x | x = data["m4d-system"]["meshfordata-external-data"]["medical_taxonomies.json"].MedicalRoles[_] }

Sensitivity = { x | x = data["m4d-system"]["meshfordata-external-data"]["taxonomies.json"].DataSensitivity[_] }

AccessTypes = { x | x = data["m4d-system"]["meshfordata-external-data"]["taxonomies.json"].DataAccessTypes[_] }

GeoDestinations = { x | x = data["m4d-system"]["meshfordata-external-data"]["taxonomies.json"].DataGeoDestinations[_] }