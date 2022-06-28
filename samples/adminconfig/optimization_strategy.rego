package adminconfig

# minimize storage cost for copy scenarios
optimize[decision] {
    input.request.usage == "copy"
    policy := {"ID": "save-cost", "description":"Save storage costs", "version": "0.1"}
    decision := {"policy": policy, "strategy": [{"attribute": "storage-cost", "directive": "min"}]}
}

# minimize distance, minimize storage cost for read scenarios
optimize[decision] {
    input.request.usage == "read"
    policy := {"ID": "general-strategy", "description":"focus on higher performance while saving storage costs", "version": "0.1"}
    optimize_distance := {"attribute": "distance", "directive": "min", "weight": "0.8"}
    optimize_storage := {"attribute": "storage-cost", "directive": "min", "weight": "0.2"}
    decision := {"policy": policy, "strategy": [optimize_distance,optimize_storage]}
}
