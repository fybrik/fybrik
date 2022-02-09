package adminconfig

# configure where transformations take place
config[{"capability": "transform", "decision": decision}] {
    policy := {"ID": "transform-geo", "description":"Governance based transformations must take place in the geography where the data is stored", "version": "0.1"}
    cluster_restrict := {"property": "metadata.region", "values": [input.request.dataset.geography]}
    decision := {"policy": policy, "restrictions": {"clusters": [cluster_restrict]}}
}

# configure the scope of the read capability
config[{"capability": "read", "decision": decision}] {
    input.request.usage.read == true
    policy := {"ID": "read-scope", "description":"Deploy read at the workload scope", "version": "0.1"}
    decision := {"policy": policy, "restrictions": {"modules": [{"property": "capabilities.scope", "values" : ["workload"]}]}}
}

# configure where the read capability will be deployed
config[{"capability": "read", "decision": decision}] {
    input.request.usage.read == true
    policy := {"ID": "read-location", "description":"Deploy read in the workload cluster", "version": "0.1"}
    cluster_restrict := {"property": "name", "values": [ input.workload.cluster.name ] }
    decision := {"policy": policy, "restrictions": {"clusters": [cluster_restrict]}}
}

# allow implicit copies by default
config[{"capability": "copy", "decision": decision}] {
    input.request.usage.read == true
    policy := {"ID": "copy-default", "description":"Implicit copies are allowed in read scenarios", "version": "0.1"}
    decision := {"policy": policy}
}

# configure when implicit copies should be made
config[{"capability": "copy", "decision": decision}] {
    input.request.usage.read == true
    input.request.dataset.geography != input.workload.cluster.metadata.region
    count(input.actions) > 0
    cluster_restrict := {"property": "metadata.region", "values": [input.request.dataset.geography]}
    account_restrict := {"property": "region", "values": [input.workload.cluster.metadata.region]}
    policy := {"ID": "copy-remote", "description":"Implicit copies should be used if the data is in a different region than the compute, and transformations are required", "version": "0.1"}
    decision := {"policy": policy, "deploy": "True", "restrictions": {"clusters": [cluster_restrict], "storageaccounts": [account_restrict]}}
}

# restrict storage for copy
config[{"capability": "copy", "decision": decision}] {
    input.request.usage.copy == true
    input.request.dataset.geography != input.workload.cluster.metadata.region
    account_restrict := {"property": "storage-cost", "range": {"max": 90}}
    policy := {"ID": "copy-restrict-storage", "description":"Use cheaper storage", "version": "0.1"}
    decision := {"policy": policy, "restrictions": {"storageaccounts": [account_restrict]}}
}
