# IT Config Policies and Data Plane Optimization

## What are IT config policies?

IT config policies are the mechanism via which the organization may influence the construction of the data plane, taking into account infrastructure capabilities and costs. Fybrik takes into account the workload context, the data metadata, the data governance policies and the configuration policies when defining the data plane. IT config policies influence what capabilities should be deployed (e.g. read, copy), in which clusters they should be deployed, and selection of the most appropriate module that implements the capability.

## Input to policies

The `input` object includes general application data such as workload cluster and application properties, as well as dataset details (user requirements, metadata).

Available properties:

- `cluster.name`: name of the workload cluster
- `cluster.metadata.region`: region of the workload cluster
- `properties`: application/workload properties defined in FybrikApplication, e.g. `properties.intent`
- `request.metadata`: asset metadata as defined in catalog taxonomy, e.g `request.metadata.geography`
- `usage`: a set of boolean properties associated with data use: `usage.read`, `usage.write`, `usage.copy`

## Syntax 

Policies are written in rego files. Each file declares a package `adminconfig`.

Rules are written in the following syntax: `config[{"capability": capability, "decision": decision}]` where

`capability` represents a required module capability, such as "read", "write", "transform" and "copy".

`decision` is a JSON structure that matches `Decision` defined above. 

```
{ 
	"policy": {"ID": <id>, "description": <description>, "version": <version>}, 
	"deploy": <"True", "False">,
	"restrictions": {
		"modules": <list of restrictions>,
		"clusters": <list of restrictions>,
        "storageaccounts": <list of restrictions>,
	},
}
```
`restriction` restricts a `property` to either a set of `values` or a value in a given `range`.

For example, the policy above restricts the choice of clusters and modules for a read capability by narrowing the choice of deployment clusters to the workload cluster, and restricting the module type to service.

```
config[{"capability": "read", "decision": decision}] {
    input.request.usage == "read"
    policy := {"ID": "read-location", "description":"Deploy read in the workload cluster", "version": "0.1"}
    cluster_restrict := {"property": "name", "values": [ input.workload.cluster.name ] }
    module_restrict := {"property": "type", "values": ["service"]}
    decision := {"policy": policy, "restrictions": {"clusters": [cluster_restrict], "modules": [module_restrict]}}
}
```


`policy` provides policy metadata: unique ID, human-readable description and version

`restrictions` provides restrictions for `modules`, `clusters` and `storageaccounts`.
Each restriction provides a list or a range of allowed values for a property of module/cluster/storageaccount object. For example, to restrict a module type to either "service" or "plugin", we'll use "type" as a property, and [ "service","plugin ] as a list of allowed values.
Properties of a module can be found inside [`FybrikModule`](../reference/crds.md#fybrikmodule) Spec.
Properties of a storage account are listed inside [`FybrikStorageAccount`](../reference/crds.md#fybrikstorageaccount).
Cluster is not a custom resource. It has the following properties:
- name: cluster name
- metadata.region: cluster region
- metadata.zone: cluster zone

`deploy` receives "True"/"False" values. These values indicate whether the capability should or should not be deployed. If not specified in the policy, it's up to Fybrik to decide on the capability deployment.


### Out of the box policies

Out of the box policies come with the fybrik deployment. They define the deployment of basic capabilities, such as read, write, copy and delete. 

```
package adminconfig

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

# delete capability deployment
config[{"capability": "delete", "decision": decision}] {
    input.request.usage == "delete"
    policy := {"ID": "delete-request", "description":"Delete capability is requested by the user", "version": "0.1"}
    decision := {"policy": policy, "deploy": "True"}
}

# do not deploy copy in scenarios different from read or copy
config[{"capability": "copy", "decision": decision}] {
    input.request.usage != "read"
    input.request.usage != "copy"
    policy := {"ID": "copy-disabled", "description":"Copy capability is not requested", "version": "0.1"}
    decision := {"policy": policy, "deploy": "False"}
}

# do not deploy read in other scenarios
config[{"capability": "read", "decision": decision}] {
    input.request.usage != "read"
    policy := {"ID": "read-disabled", "description":"Read capability is not requested", "version": "0.1"}
    decision := {"policy": policy, "deploy": "False"}
}

# do not deploy write in other scenarios
config[{"capability": "write", "decision": decision}] {
    input.request.usage != "write"
    policy := {"ID": "write-disabled", "description":"Write capability is not requested", "version": "0.1"}
    decision := {"policy": policy, "deploy": "False"}
}

# do not deploy delete in other scenarios
config[{"capability": "delete", "decision": decision}] {
    input.request.usage != "delete"
    policy := {"ID": "delete-disabled", "description":"Delete capability is not requested", "version": "0.1"}
    decision := {"policy": policy, "deploy": "False"}
}

```

### Extended policies

The extended policies define advanced deployment requirements, such as where read or transform modules should run, what should be the scope of module deployments, and more. 

The policies below are provided as a sample and can be updated for the production deployment.

```
package adminconfig

# configure where transformations take place
config[{"capability": "transform", "decision": decision}] {
    policy := {"ID": "transform-geo", "description":"Governance based transformations must take place in the geography where the data is stored", "version": "0.1"}
    cluster_restrict := {"property": "metadata.region", "values": [input.request.dataset.geography]}
    decision := {"policy": policy, "restrictions": {"clusters": [cluster_restrict]}}
}

# configure the scope of the read capability
config[{"capability": "read", "decision": decision}] {
    input.request.usage == "read"
    policy := {"ID": "read-scope", "description":"Deploy read at the workload scope", "version": "0.1"}
    decision := {"policy": policy, "restrictions": {"modules": [{"property": "capabilities.scope", "values" : ["workload"]}]}}
}

# configure where the read capability will be deployed
config[{"capability": "read", "decision": decision}] {
    input.request.usage == "read"
    policy := {"ID": "read-location", "description":"Deploy read in the workload cluster", "version": "0.1"}
    cluster_restrict := {"property": "name", "values": [ input.workload.cluster.name ] }
    decision := {"policy": policy, "restrictions": {"clusters": [cluster_restrict]}}
}

# allow implicit copies by default
config[{"capability": "copy", "decision": decision}] {
    input.request.usage == "read"
    policy := {"ID": "copy-default", "description":"Implicit copies are allowed in read scenarios", "version": "0.1"}
    decision := {"policy": policy}
}

```

### How to provide custom policies

In order to deploy Fybrik with customized policies, perform the following steps 

1. Clone the github repository of Fybrik for the required release: `git clone -b releases/<version> https://github.com/fybrik/fybrik.git`
2. Copy the rego files containing customized policies to fybrik/charts/fybrik/files/adminconfig/ folder 
3. Install Fybrik:
```
cd fybrik
helm install fybrik-crd charts/fybrik-crd -n fybrik-system --wait
helm install fybrik charts/fybrik --set global.tag=master --set global.imagePullPolicy=Always -n fybrik-system --wait
```

### How to add start and/or expiry dates to policies

By utilizing the time built-in functions of OPA, an effective date and/or expiry date of a policy can be defined. The related built-in functions are:

``` 
output := time.now_ns() //the current date
output := time.parse_rfc3339_ns(value) //the specified date in RFC3339 format
```
`parse_rfc3339_ns` enables to add the expiry date as well as the the date for the policy to become effective, and `now_ns` captures the date when policies are applied. Through comparisons, it can be acquired whether the current policy is still valid. Below is an example.

```
package adminconfig

# vaild from 2022.1.1, expire on 2022.6.1
config[{"capability": "copy", "decision": decision}] {
    policy := {"ID": "test-1", "description": "forbid making copies", "version": "0.1"}
    nowDate := time.now_ns()
    startDate := time.parse_rfc3339_ns("2022-01-01T00:00:00Z")
    expiration := time.parse_rfc3339_ns("2022-06-01T00:00:00Z")
    nowDate >= startDate
    nowDate < expiration
    decision := {"policy": policy, "deploy": "False"}
}
```
Note that an empty ConfigDecisions map will be returned if the expiration date is exceeded by the time when the policy is applied. 

### How to update policies after Fybrik is already deployed

Updating policies is done by updating `fybrik-adminconfig` config map in the controller plane.

To do that, first, download all files to some directory, e.g. /tmp/adminconfig, after that update the files, and finally, upload them to the config map. The steps below demonstrate how to add a new rego file `samples/adminconfig/quickstart-policies.rego`. 

```
#!/bin/bash
kubectl get cm fybrik-adminconfig -o json > tmp.json
mkdir -p /tmp/adminconfig
files=$(cat tmp.json | jq '.data' | jq -r 'keys[]')
for k in $files; do
    name=".data[\"$k\"]";
    cat tmp.json | jq -r $name > /tmp/adminconfig/$k;
done
cp samples/adminconfig/quickstart_policies.rego /tmp/adminconfig/
kubectl create configmap fybrik-adminconfig --from-file=/tmp/adminconfig -o yaml --dry-run=client | kubectl replace -n fybrik-system -f -
rm -rf /tmp/adminconfig
rm -rf tmp.json
```

## Optimization goals

In a typical Fybrik deployment there may be several possibilities to create a data plane that satisfies the user requirements, governance and configuration policies. Based on the enterprise policy, an IT administrator may affect the choice of the data plane by defining a policy with optimization goals. 
An optimization goal attempts to minimize or maximize a specific [infrastructure attribute](../tasks/infrastructure.md#how-to-define-infrastructure-attributes).
While [IT config policies](#what-are-it-config-policies) are always enforced, the data plane optimization is disabled by default. To enable data-plane optimization, the [Optimizer component](./optimizer.md) must be enabled as explained [here](../tasks/data-plane-optimization.md#enabling-the-optimizer). 

### Syntax 

Optimization rules are written in rego files in a package `adminconfig`.

Rules are written in the following syntax: `optimize[decision]` where

- `decision` is a JSON structure with the following fields:
- `policy` - policy metadata: unique ID, human-readable description and a version
- list of `goals` including attribute name, optimization directive(`min` or `max`) and optionally a weight.

For example, the following rule attempts to minimize storage cost in copy scenarios.

```
# minimize storage cost for copy scenarios
optimize[decision] {
    input.request.usage == "copy"
    policy := {"ID": "save-cost", "description":"Save storage costs", "version": "0.1"}
    decision := {"policy": policy, "strategy": [{"attribute": "storage-cost", "directive": "min"}]}
}
```

### Weights

If more than one goal is provided, they can have a different weight. By default, all weights are equal to 1. 
For example, the rule below defines two goals with weights in the 4:1 ratio meaning that the optimizer try to optimize distance and storage costs, but will give a higher priority to distance.
```
# minimize distance, minimize storage cost for read scenarios
optimize[decision] {
    input.request.usage == "read"
    policy := {"ID": "general-strategy", "description":"focus on higher performance while saving storage costs", "version": "0.1"}
    optimize_distance := {"attribute": "distance", "directive": "min", "weight": "0.8"}
    optimize_storage := {"attribute": "storage-cost", "directive": "min", "weight": "0.2"}
    decision := {"policy": policy, "strategy": [optimize_distance,optimize_storage]}
}
```

