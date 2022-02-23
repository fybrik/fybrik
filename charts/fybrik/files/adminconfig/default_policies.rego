package adminconfig

# out-of-the-box policies

# read capability deployment
config[{"capability": "read", "decision": decision}] {
    input.request.usage == "read"
    policy := {"ID": "read-default-enabled", "description":"Read capability is requested for read workloads", "version": "0.1"}
    decision := {"policy": policy, "deploy": "True"}
}

# write capability deployment
config[{"capability": "write", "decision": decision}] {
    input.request.usage == "write"
    policy := {"ID": "write-default-enabled", "description":"Write capability is requested for workloads that write data", "version": "0.1"}
    decision := {"policy": policy, "deploy": "True"}
}

# copy requested by the user
config[{"capability": "copy", "decision": decision}] {
    input.request.usage == "copy"
    policy := {"ID": "copy-request", "description":"Copy (ingest) capability is requested by the user", "version": "0.1"}
    decision := {"policy": policy, "deploy": "True"}
}
