package data_policies

#general functions that make data-policies composing easier

verify_access_type {
		compare_str(AccessType(), AccessTypes[_])
}

verify_purpose {
		compare_str(Purpose(), Purposes[_])
}

verify_role {
	compare_str(Role(), Roles[_])
}

verify_geography {
    compare_str(ProcessingGeo(), GeoDestinations[_])
}

dataset_has_tag(tag) {
    compare_str(tag,  DatasetTags()[_])
}

dataset_has_tag_not(tag) {
    compare_str_not(tag,  DatasetTags()[_])
}

column_has_tag(tag) {
	compare_str(tag, input.details.metadata.components_metadata[_].tags[_])
}

check_purpose(purpose) {
    compare_str(purpose, Purpose())
}

check_role(role) {
    compare_str(role, Role())
}

check_role_not(role) {
    compare_str_not(role, Role())
}

check_access_type(access_types) {
    compare_str(AccessType(), access_types[_])
}

check_destination(destinations) {
    compare_str(DestinationGeo(), destinations[_])
}

check_processingGeo_not(processingGeo) {
	compare_str_not(processingGeo, ProcessingGeo())
}

clean_string(str) = result {
    str2 := lower(str)
    str3 = replace(str2, " ", "")
    str4 := replace(str3, "-", "")
    str5 := replace(str4, "_", "")

    result=str5
}

compare_str(str1, str2) {
    clean_string(str1) == clean_string(str2)
}

compare_str_not(str1, str2) {
    clean_string(str1) != clean_string(str2)
}