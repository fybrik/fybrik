package data_policies

#general enforcment action structure
enforcement_action_struct = {
    "action_name" : "<name of action>",
    "desription" : "<free text description of the action>",
    "arguments" : "<arguments set can be different for each action>",
    "used_policy" : "<used_policy_struct>"
}

build_action_from_policies(used_policy) = action {
    action = {
        "used_policy" : used_policy
    }
}

build_action_from_name(action_name, used_policy) = action {
    action = {
        "action_name" : action_name,
        "desription" : action_name,
        "arguments" :[],
        "used_policy" : used_policy
    }
}

build_action(action_name, description, arguments, used_policy) = action {
    action = {
        "action_name" : action_name,
        "description" : description,
        "arguments" :arguments,
        "used_policy" : used_policy
    }
}

################################### Enforcement Actions #######################################

#deny access
deny_access_struct = {
    "action_name" : "deny access",
    "description" : "Access to this data asset is denied",
    "arguments" : {},
    "used_policy" : "<used_policy_struct>"
}

deny_write_struct = {
    "action_name" : "deny",
    "description" : "Writing of this data asset is denied",
    "arguments" : {},
    "used_policy" : "<used_policy_struct>"
}

build_deny_access_action(used_policies) = action {
    action = build_action(deny_access_struct.action_name, deny_access_struct.description, deny_access_struct.arguments, used_policies)
}

build_deny_write_action(used_policies) = action {
    action = build_action(deny_write_struct.action_name, deny_write_struct.description, deny_write_struct.arguments, used_policies)
}

#remove column
remove_column_struct = {
    "action_name" : "remove column",
    "description" : "Single column is removed",
    "arguments" : { 
        "column_name": "<column name>"
    },
    "used_policy" : "<used_policy_struct>"
}

build_remove_column_action(column_name, used_policies) = action {
    args := { 
       "column_name" : column_name
    }
    action = build_action(remove_column_struct.action_name, remove_column_struct.description, args, used_policies)
}

#encrypt colmn
encrypt_column_struct = {
    "action_name" : "encrypt column",
    "description" : "Single column is encrypted with its own key",
    "arguments" : { 
        "column_name": "<column name>"
    },
    "used_policy" : "<used_policy_struct>"
}

build_encrypt_column_action(column_name, used_policies) = action {
    args := { 
       "column_name" : column_name
    }
    action = build_action(encrypt_column_struct.action_name, encrypt_column_struct.description, args, used_policies)
}

#mask_redact_column
redact_column_struct = {
    "action_name" : "redact column",
    "description" : "Single column is obfuscated with XXX instead of values",
    "arguments" : { 
        "column_name": "<column name>"
    },
    "used_policy" : "<used_policy_struct>"
}

build_redact_column_action(column_name, used_policies) = action {
    args := { 
       "column_name" : column_name
    }
    action = build_action(redact_column_struct.action_name, redact_column_struct.description, args, used_policies)
}

#periodic_blackout
periodic_blackout_struct = {
    "action_name" : "periodic blackout",
    "description" : "Access to dataset is denied based on date of the access",
    "arguments" : { 
        #only one of the arguments should be filled in
        "monthly_days_end": "<number of days before the end of month when data is denied>",
        "yearly_days_end": "<number of days before the end of year when data is denied>",
    },
    "used_policy" : "<used_policy_struct>"
}

build_monthly_periodic_blackout_action(days_before_month_end, used_policies) = action {
    args := { 
       "monthly_days_end" : days_before_month_end
    }
    action = build_action(periodic_blackout_struct.action_name, periodic_blackout_struct.description, args, used_policies)
}

build_yearly_periodic_blackout_action(days_before_year_end, used_policies) = action {
    args := { 
       "yearly_days_end" : days_before_year_end
    }
    action = build_action(periodic_blackout_struct.action_name, periodic_blackout_struct.description, args, used_policies)
}