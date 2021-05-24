package dataapi.authz

import data.data_policies as dp

#Example of data policies that use "data_policies" package to create easily data policies that deny access or transform the data accordingly

transform[action] {
	description = "location data should be removed before copy"

    #user context and access type check
    dp.check_access_type(["COPY"])

    column_names := dp.column_with_tag("location")
    action = dp.build_remove_column_action(column_names[_], dp.build_policy_from_description(description))
}

transform[action] {
	description = "sensitive columns in health data should be removed"

    #user context and access type check
    dp.check_access_type(["COPY", "READ"])

    dp.dataset_has_tag("HealthData")

    column_names := dp.column_with_tag("SPI")
    action = dp.build_remove_column_action(column_names[_], dp.build_policy_from_description(description))
}

transform[action] {
	description = "encrypt sensitive personal and health data on COPY out of united states for health data assets"

    #user context and access type check
    dp.check_access_type("COPY")

    dp.dataset_has_tag("HealthData")
    not dp.check_destination([dp.GeoDestinations.US])

    column_names := dp.column_with_any_tag(["SPI", "SMI"])
    #action = dp.build_encrypt_column_action(column_names[_], dp.build_policy_from_description(description))
    action = dp.build_redact_column_action(column_names[_], dp.build_policy_from_description(description))
}

#for transactions dataset
transform[action] {
	#description = "test for transactions dataset that encrypts some columns by name"
    description = "test for transactions dataset that redacts some columns by name"

    #user context and access type check
    dp.check_access_type(["READ"])

    dp.dataset_has_tag("finance")

    column_names := dp.column_with_any_name({"nameOrig", "nameDest", "nameDest::6", "nameOrig::3"})
    #action = dp.build_encrypt_column_action(column_names[_], dp.build_policy_from_description(description))
    action = dp.build_redact_column_action(column_names[_], dp.build_policy_from_description(description))

}

#for transactions dataset
deny[action] {
	description = "test for transactions dataset with deny"

    #user context and access type check
    dp.check_access_type(["COPY"])

    dp.dataset_has_tag("Finance")

    action = dp.build_deny_access_action(dp.build_policy_from_description(description))
}