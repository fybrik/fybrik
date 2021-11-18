package adminconfig

# out-of-the-box policies

# read capability deployment
config[{"read": decision}] {
    read_request := input.request.usage.read
    policy := {"ID": "read-default", "description":"Read capability is requested for read workloads"}
    decision := {"policy": policy, "deploy": read_request}
}

# write capability deployment
config[{"write": decision}] {
    write_request := input.request.usage.write 
    policy := {"ID": "write-default", "description":"Write capability is requested for workloads that write data"}
    decision := {"policy": policy, "deploy": write_request}
}

# allow implicit copies by default
config[{"copy": decision}] {
    input.request.usage.read == true
    policy := {"ID": "copy-default", "description":"Implicit copies are allowed in read scenarios"}
    decision := {"policy": policy}
}
