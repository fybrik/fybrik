package dataapi.authz
import data.data_policies as dp

deny[action] {
    description = "Default Action is Deny"
    action = dp.build_deny_access_action(dp.build_policy_from_description(description))
}