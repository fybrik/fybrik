# Configuration Policies

## What are configuration policies?

Configuration policies are the mechanism via which the organization may influence the construction of the data plane, taking into account infrastructure capabilities and costs. Fybrik takes into account the workload context, the data metadata, the data governance policies and the configuration policies when defining the data plane. The configuration policies influence what capabilities should be deployed (e.g. read, copy), in which clusters they should be deployed, and selection of the most appropriate module that implements the capability.

## Input to policies

The `input` object includes general application data such as workload cluster and application properties, as well as dataset details (user requirements, metadata, required actions).

Available properties:
- `cluster.name`: name of the workload cluster
- `cluster.metadata.region`: region of the workload cluster
- `properties`: application/workload properties defined in FybrikApplication, e.g. `properties.intent`
- `request.metadata`: asset metadata as defined in catalog taxonomy, e.g `request.metadata.geography`
- `usage`: a set of boolean properties associated with data use: `usage.read`, `usage.write`, `usage.copy`

## Syntax 

Policies are written in rego files. Each file declares a package `adminconfig`.

Rules are written in the following syntax: `config[{capability: decision}]` where

`capability` represents a required module capability, such as "read", "write", "transform" and "copy".

`decision` is a JSON structure that matches `Decision` defined above. 

```
{ 
	"policy": {"ID": <id>, "policySetID": <setId>, "description": <description>}, 
	"deploy": <true, false>,
	"restrictions": {
		"modules": <map {key, list-of-values}>,
		"clusters": <map {key, list-of-values}>,
	},
}
```

For example, the policy above restricts the choice of clusters and modules for a read capability by narrowing the choice of deployment clusters to the workload cluster, and restricting the module type to service.

```
config[{"read": decision}] {
    input.request.usage.read == true
    policy := {"ID": "read-ID", "description":"Deploy read as a service in the workload cluster"}
    clusters := { "name" : [ input.workload.cluster.name ] }
    modules := { "type": ["service"]}
    decision := {"policy": policy, "restrictions": {"clusters": clusters, "modules": modules}}
}
```


`policy` provides policy metadata: unique ID, human-readable description and `policySetID` (see ### Policy Set ID)

`restrictions` provides restrictions for `modules`, `clusters` and `storageaccounts`.
Each restriction provides a list of allowed values for a property of module/cluster/storageaccount object. For example, to restrict a module type to either "service" or "plugin", we'll use "type" as a key, and [ "service","plugin ] as a list of allowed values.
Properties of a module can be found inside [`FybrikModule`](../reference/crds.md#fybrikmodule) Spec.
Properties of a storage account are listed inside [`FybrikStorageAccount`](../reference/crds.md#fybrikstorageaccount).
Cluster is not a custom resource. It has the following properties:
- name: cluster name
- metadata.region: cluster region
- metadata.zone: cluster zone

`deploy` receives "true"/"false" values. These values indicate whether the capability should or should not be deployed. If not specified in the policy, it's up to Fybrik to decide on the capability deployment.

### Policy Set ID

Fybrik supports evaluating different sets of policies for different FybrikApplications. It is possible to define a policy for a specific `policySetID` which will be trigered only if it matches the `policySetID` defined in FybrikApplication. 
If a policy does not specify a policy set id, it will be considered as relevant for all FybrikApplications.
In a similar way, all policies are relevant for a FybrikApplication that does not specify a policy set id, to support a use-case of a single policy set for all.

### Out of the box policies

Out of the box policies come with the fybrik deployment. They define the deployment of basic capabilities, such as read and write. 
```
package adminconfig

config[{"read": decision}] {
    read_request := input.request.usage.read
    policy := {"ID": "read-default", "description":"Read capability is requested for read workloads"}
    decision := {"policy": policy, "deploy": read_request}
}

config[{"write": decision}] {
    write_request := input.request.usage.write 
    policy := {"ID": "write-default", "description":"Write capability is requested for workloads that write data"}
    decision := {"policy": policy, "deploy": write_request}
}
```

### Extended policies

The extended policies define advanced deployment requirements, such as where read or transform modules should run, what should be the scope of module deployments, and more. 

The policies below are provided as a sample and can be updated for the production deployment.

```
package adminconfig

# configure where transformations take place
config[{"transform": decision}] {
    policy := {"ID": "transform-geo", "description":"Governance based transformations must take place in the geography where the data is stored"}
    clusters := { "metadata.region" : [ input.request.dataset.geography ] }
    decision := {"policy": policy, "restrictions": {"clusters": clusters}}
}

# configure the scope of the read capability
config[{"read": decision}] {
    input.request.usage.read == true
    policy := {"ID": "read-scope", "description":"Deploy read at the workload scope"}
    decision := {"policy": policy, "restrictions": {"modules": {"capabilities.scope" : ["workload"]}}}
}

# configure where the read capability will be deployed
config[{"read": decision}] {
    input.request.usage.read == true
    policy := {"ID": "read-location", "description":"Deploy read in the workload cluster"}
    clusters := { "name" : [ input.workload.cluster.name ] }
    decision := {"policy": policy, "restrictions": {"clusters": clusters}}
}

# allow implicit copies by default
config[{"copy": decision}] {
    input.request.usage.read == true
    policy := {"ID": "copy-default", "description":"Implicit copies are allowed in read scenarios"}
    decision := {"policy": policy}
}

# configure when implicit copies should be made
config[{"copy": decision}] {
    input.request.usage.read == true
    input.request.dataset.geography != input.workload.cluster.metadata.region
    count(input.actions) > 0
    clusters := { "metadata.region" : [ input.request.dataset.geography ] }
    policy := {"ID": "copy-remote", "description":"Implicit copies should be used if the data is in a different region than the compute, and transformations are required"}
    decision := {"policy": policy, "deploy": true, "restrictions": {"clusters": clusters}}
}

```

### How to add / modify / delete policies

Updating policies is done by updating `fybrik-adminconfig` config map in the controller plane.

The steps below demonstrate how to add a new rego file (samples/adminconfig/quickstart-policies.rego)

```
mkdir -p /tmp/adminconfig
cp samples/adminconfig/quickstart-policies.rego /tmp/adminconfig/
cp charts/fybrik/files/adminconfig/* /tmp/adminconfig
kubectl create configmap fybrik-adminconfig --from-file=/tmp/adminconfig -o yaml --dry-run=client | kubectl replace -n fybrik-system -f -
rm -rf /tmp/adminconfig
```
   
