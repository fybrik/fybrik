package adminconfig

# configure where transformations take place
config[{"capability": "transform", "decision": decision}] {
    policy := {"ID": "transform-geo", "description":"Governance based transformations must take place in the geography where the data is stored", "version": "0.1"}
    clusters := { "metadata.region" : [ input.request.dataset.geography ] }
    decision := {"policy": policy, "restrictions": {"clusters": clusters}}
}

# configure the scope of the read capability
config[{"capability": "read", "decision": decision}] {
    input.request.usage[_] == "read"
    policy := {"ID": "read-scope", "description":"Deploy read at the workload scope", "version": "0.1"}
    decision := {"policy": policy, "restrictions": {"modules": {"capabilities.scope" : ["workload"]}}}
}

# configure where the read capability will be deployed
config[{"capability": "read", "decision": decision}] {
    input.request.usage[_] == "read"
    policy := {"ID": "read-location", "description":"Deploy read in the workload cluster", "version": "0.1"}
    clusters := { "name" : [ input.workload.cluster.name ] }
    decision := {"policy": policy, "restrictions": {"clusters": clusters}}
}

# allow implicit copies by default
config[{"capability": "copy", "decision": decision}] {
    input.request.usage[_] == "read"
    policy := {"ID": "copy-default", "description":"Implicit copies are allowed in read scenarios", "version": "0.1"}
    decision := {"policy": policy}
}

# configure when implicit copies should be made
config[{"capability": "copy", "decision": decision}] {
    input.request.usage[_] == "read"
    input.request.dataset.geography != input.workload.cluster.metadata.region
    count(input.actions) > 0
    clusters := { "metadata.region" : [ input.request.dataset.geography ] }
    policy := {"ID": "copy-remote", "description":"Implicit copies should be used if the data is in a different region than the compute, and transformations are required", "version": "0.1"}
    decision := {"policy": policy, "deploy": "True", "restrictions": {"clusters": clusters}}
}
