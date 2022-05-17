package adminconfig

# require copies for production workloads
config[{"capability": "copy", "decision": decision}] {
    input.request.usage == "read"
    input.request.dataset.geography != input.workload.cluster.metadata.region
    input.workload.properties.stage == "PROD"
    policy := {"ID": "copy-default", "description":"Copy remote assets in production workloads", "version": "0.1"}
    decision := {"policy": policy, "deploy": "True"}
}
