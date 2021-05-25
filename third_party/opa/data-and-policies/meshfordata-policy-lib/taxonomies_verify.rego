package data_policies

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
