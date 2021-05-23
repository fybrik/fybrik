package data_policies

verify_access_type {
		compare_str(AccessType(), AccessTypes[_])
}

verify_intent {
		compare_str(Intent(), Intents[_])
}

verify_role {
	compare_str(Role(), Roles[_])
}

verify_geography {
    compare_str(ProcessingGeo(), GeoDestinations[_])
}
