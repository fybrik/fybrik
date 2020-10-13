package data_policies

transform[action] {
	description = "location data should be removed before copy"

	correct_input
    
    #user context and access type check
    check_access_type([AccessTypes.COPY])
    
    column_names := column_with_tag("location")
    action = build_remove_column_action(column_names[_], build_policy_from_description(description))
}  

transform[action] {
	description = "sensitive columns in health data should be removed"

	correct_input
    
    #user context and access type check
    check_access_type([AccessTypes.COPY, AccessTypes.READ])
    
    dataset_has_tag("HealthData")
    
    column_names := column_with_tag("SPI")
    action = build_remove_column_action(column_names[_], build_policy_from_description(description))
} 

transform[action] {
	description = "encrypt sensitive personal and health data on COPY out of united states for health data assets"
    
	correct_input
    
    #user context and access type check
    check_access_type(AccessTypes.COPY)
    
    dataset_has_tag("HealthData")
    not check_destination([GeoDestinations.US])
    
    column_names := column_with_any_tag(["SPI", "SMI"])
    action = build_encrypt_column_action(column_names[_], build_policy_from_description(description))
}

#internal checks
transform[action] {
	description = "test1 description"
    
	correct_input
    
    #user context and access type check
    check_access_type([AccessTypes.READ])
    
    dataset_has_tag("Finance")
    
    column_names := column_with_any_name({"nameOrig", "nameDest"})
    action = build_encrypt_column_action(column_names[_], build_policy_from_description(description))
}