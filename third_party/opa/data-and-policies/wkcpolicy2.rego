package wkcpolicy
import data.data_policies as dp

deny[action] {
	description = "deny if role is not Data Scientist when purpose is Fraud Detection"
	dp.correct_input
    #user context and access type check
    #dp.check_access_type([dp.AccessTypes.READ])
	dp.check_access_type(["READ"])
	dp.check_purpose("Fraud Detection")
	dp.check_role_not("Data Scientist")
	dp.dataset_has_tag("residency = Turkey")
    action = dp.build_deny_access_action(dp.build_policy_from_description(description))
}
