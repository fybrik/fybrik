package adminconfig

# out-of-the-box policies

# read capability deployment
config[{"read": decision}] {
    input.request.usage.read == true
    read_request := input.request.usage.read
    policy := {"ID": "read-default-enabled", "description":"Read capability is requested for read workloads", "version": "0.1"}
    decision := {"policy": policy, "deploy": "True"}
}

# read capability deployment
config[{"read": decision}] {
    input.request.usage.read == false
    read_request := input.request.usage.read
    policy := {"ID": "read-default-disabled", "description":"Read capability is requested for read workloads", "version": "0.1"}
    decision := {"policy": policy, "deploy": "False"}
}

# write capability deployment
config[{"write": decision}] {
    input.request.usage.write == true
    policy := {"ID": "write-default-enabled", "description":"Write capability is requested for workloads that write data", "version": "0.1"}
    decision := {"policy": policy, "deploy": "True"}
}

# write capability deployment
config[{"write": decision}] {
    input.request.usage.write == false
    policy := {"ID": "write-default-disabled", "description":"Write capability is requested for workloads that write data", "version": "0.1"}
    decision := {"policy": policy, "deploy": "False"}
}

# copy requested by the user
config[{"copy": decision}] {
    input.request.usage.copy == true
    policy := {"ID": "copy-request", "description":"Copy capability is requested by the user", "version": "0.1"}
    decision := {"policy": policy, "deploy": "True"}
}
