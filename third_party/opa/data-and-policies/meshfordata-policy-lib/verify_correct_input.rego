package data_policies

correct_input {
	count(incorrect_input) == 0
}

incorrect_input[used_policy] {
   not verify_access_type
   used_policy := build_action_from_policies(build_policy_from_description("unknown access type"))
} {
    not verify_intent
    used_policy := build_action_from_policies(build_policy_from_description("unknown intent"))
} {
    not verify_role
    used_policy := build_action_from_policies(build_policy_from_description("unknown role"))
} {
	check_access_type(["COPY"])
    not verify_geography
    used_policy := build_action_from_policies(build_policy_from_description("unknown geography to copy the data"))
}