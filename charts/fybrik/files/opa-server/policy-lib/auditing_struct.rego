package data_policies

#general structure ofused policy for auditing
used_policy_struct = {
    "policy_id" : "<unique id>",
    "description" : "<free text description of the policy reason>",
    "policy_type" : "<classification of policy itslef>",
    "hierarchy" : "<relation to other policies>"
}

build_policy_from_id(id) = policy {
    policy = { "policy_id" : id }
}

build_policy_from_description(desc) = policy {
    policy = { "description" : desc }
}

build_policy(id, desc, type, hierarchy) = policy {
    policy = {
        "policy_id" : id,
        "description" : desc,
        "policy_type" : type,
        "hierarchy" : hierarchy
    }
}