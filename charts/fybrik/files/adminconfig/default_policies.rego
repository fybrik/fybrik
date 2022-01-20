package adminconfig

# out-of-the-box policies

# read capability deployment
config[{"read": decision}] {
    read_request := input.request.usage.read
    policy := {"ID": "read-default", "description":"Read capability is requested for read workloads", "version": "0.1"}
    decision := {"policy": policy, "deploy": read_request}
}

# write capability deployment
config[{"write": decision}] {
    write_request := input.request.usage.write 
    policy := {"ID": "write-default", "description":"Write capability is requested for workloads that write data", "version": "0.1"}
    decision := {"policy": policy, "deploy": write_request}
}

# copy requested by the user
config[{"copy": decision}] {
    input.request.usage.copy == true
    policy := {"ID": "copy-request", "description":"Copy capability is requested by the user", "version": "0.1"}
    decision := {"policy": policy, "deploy": true}
}
