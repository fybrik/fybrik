package data_policies

#In data part we provide set of general and industry specific taxonomies, also the user can add more taxonomies specific for his needs.
#Here is the place when for each category user chooses what taxonomies should be used

Purposes = data.DataPurposes

Roles = data.DataRoles | data.MedicalRoles

Sensitivity = data.DataSensitivity

AccessTypes = data.DataAccessTypes

GeoDestinations = data.DataGeoDestinations