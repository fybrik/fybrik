Component for Config Policy Decisions
==========================

## Summary

Design a component that will evaluate config policies and provide the manager with deployment decisions, such as what capabilities should be deployed, while restricting the choice of the clusters, modules and other resources used for the deployment. The evaluation process is based on the workload in use by the fybrikapplication, asset metadata and infrastructure (e.g. available clusters).

## Goals

1. Design an interface between the config policy evaluator and the manager.

2. Design an interface between the config policy evaluator and OPA

## Interface with the manager

### Input (`adminconfig.EvaluatorInput`)

A dynamic input constructed per a FybrikApplication, per a single dataset.
It provides general application data such as workload cluster and application properties, as well as dataset details (user requirements, metadata, required actions.
```
// WorkloadInfo holds workload details such as the cluster where the workload is running,  
// and additional properties defined in the taxonomy, e.g. workload type
type WorkloadInfo struct {
	// Unique fybrikapplication id used for logging
	UID string `json:"uid"`
	// Policy set id to allow evaluation of a specific set of policies per fybrikapplication
	PolicySetID string `json:"policySetID"`
	// Cluster where the user workload is running
	Cluster multicluster.Cluster `json:"cluster"`
	// Application/workload properties
	Properties api.ApplicationDetails `json:"properties,omitempty"`
}

// DataRequest is a request to use a specific asset
type DataRequest struct {
	// asset identifier
	DatasetID string `json:"datasetID"`
	// requested interface
	Interface api.InterfaceDetails `json:"interface"`
	// requested usage, e.g. "read": true, "write": false
	Usage map[api.DataFlow]bool `json:"usage"`
	// Asset metadata
	Metadata *assetmetadata.DataDetails `json:"dataset"`
}

// EvaluatorInput is an input to Configuration Policies Evaluator.
// Used to evaluate configuration policies.
type EvaluatorInput struct {
	// Workload configuration
	Workload WorkloadInfo `json:"workload"`
	// Requirements for asset usage
	Request DataRequest `json:"request"`
	// Governance Actions for reading data (relevant for read scenarios only)
	GovernanceActions []model.Action `json:"actions"`
}
```
On top of the dynamic input based on a specific FybrikApplication, an `Infrastructure` object is used by the config policy evaluator to get deployment infrastructure information: available clusters, available storage, bandwidth metrics, etc. Evaluator uses `InfrastructureManager` to obtain the `Infrastructure` data. In the initial implementation the infrastructure metadata will be stored in a config file, and manually updated. In the future updates may be done by automated processes outside of fybrik.


In the current implementation, `Infrastructure` is defined as:
```
type Infrastructure struct {
	// Clusters available for deployment
	Clusters []multicluster.Cluster `json:"clusters"`
}
```
Future versions will include additional attributes (storage, bandwidth, ...). The attributes will be defined in the taxonomy. 


### Output (`adminconfig.EvaluatorOutput`)

The result of evaluating config policies on a given dataset contains deployment `Decision` for each capablility, after merging all relevant decisions for the capability. The result of the evaluation should always have at least one capability returned - that which is associated with the flow (read, write, copy ...). There may be additional capabilities returned (indicating additional capabilities which should/could/should not be included in the data path) based on the policies.
 A conflict during evaluation result is marked by `Valid == false`
 For each capability, the decision will provide the manager with the following information:
 - whether the capability should be deployed, should not be deployed, or is allowed to be deployed based on manager decision.
 - a restriction on the clusters or regions where the capability will be deployed
 - a restriction on the modules deploying this capability (e.g. deployment scope)
 
```
// EvaluatorOutput is an output of ConfigurationPoliciesEvaluator.
// Used by manager to decide which modules are deployed and in which cluster.
type EvaluatorOutput struct {
	// Valid is true when there is no conflict between the decisions, and false otherwise
	Valid bool
	// Dataset identifier
	DatasetID string
	// Unique fybrikapplication id used for logging
	UID string `json:"uid"`
	// Policy set id used in the evaluation
	PolicySetID string `json:"policySetID"`
	// Decisions per capability (after being merged)
	ConfigDecisions DecisionPerCapabilityMap
}
```

`Decision` is a result of evaluating a configuration policy which satisfies the predicates of the policy.
`Deploy` is a deployment decision (require, forbid or allow).
`DeploymentRestrictions` restricts the choice of the modules to be deployed, deployment clusters, storage accounts, etc.
`Policy` combines IDs and descriptions of all policies that have been applied.

```
// DecisionPolicy is a justification for a policy that consists of a unique id, id of a policy set and a human readable desciption
type DecisionPolicy struct {
	ID          string `json:"ID"`
	PolicySetID string `json:"policySetID"`
	Description string `json:"description,omitempty"`
}

// Deployment restrictions on modules, clusters and additional resources that will be added in the future
type Restrictions struct {
	// Restrictions on clusters used for deployment
	Clusters []string `json:"clusters,omitempty"`
	// Restrictions on modules of the type “key”: “value” when the key is a module property (e.g. scope, type) and the value is an allowed value (e.g. asset, plugin)
	ModuleRestrictions map[string]string `json:"modules,omitempty"`
}

// Decision is a result of evaluating a configuration policy which satisfies the specified predicates
type Decision struct {
	// a decision regarding deployment: True = require, False = forbid, Unknown = allow
	Deploy corev1.ConditionStatus `json:"deploy,omitempty"`
	// Deployment restrictions on modules, clusters and additional resources
	DeploymentRestrictions Restrictions `json:"restrictions,omitempty"`
	// Descriptions of policies that have been used for evaluation
	Policy DecisionPolicy `json:"policy,omitempty"`
}

type DecisionPerCapabilityMap map[api.CapabilityType]Decision
```

### Functionality (`adminconfig.EvaluatorInterface`)

Any implementation of the config policy evaluator should implement this interface.

```
// EvaluatorInterface is an interface for config policies' evaluator
type EvaluatorInterface interface {
	SetupWithInfrastructureManager(mgr *InfrastructureManager)
	Evaluate(in *EvaluatorInput) (EvaluatorOutput, error)
}
```

`SetupWithInfrastructureManager` attaches `InfrastructureManager` to the evaluator to obtain the `Infrastructure` object.
`Evaluate` evaluates config policies based on `EvaluatorInput` and returns `EvaluatorOutput`.

## Interface with OPA

Configuration policies are written in Rego language and are evaluated using OPA (Open Policy Agent).
Interaction between the evaluator and OPA is done using internal OPA golang packages (see https://pkg.go.dev/github.com/open-policy-agent/opa/rego#Rego.Eval)
OPA supports different ways of evaluating policies, such as communicating with a deployed server using REST APIs, or a GO library which can be integrated into the solution. We have chosen to take the second approach for two reasons: the evaluation has less overhead than the REST API because all the communication happens in the same operating-system process, and second, we do not see this as being a pluggable component in the fybrik architecture.

Information that passes from config policies' evaluator to OPA has three main ingredeients:

- The `data` json object with infrastructure details, such as available clusters, available object stores, etc. Infrastructure is known at the deployment time and is not changed frequently. The `Infrastructure` object is obtained using `InfrastructureManager` that has interfaces and clients that can access various kubernetes resources, such as configmaps and custom resources. 

- `Rego Modules` with policies created by IT administrator. Modules exist at the deployment time and are unlikely to be changed frequently. 

- Dynamic `input` that has the details of the workload, asset metadata and the user request. Inputs are constructed on each reconciliation of FybrikApplication.

Interaction with OPA is done in two steps. First, a `PreparedEvalQuery` is created using the `data` json and compiled `modules`. This should be done upon changes in policies and/or infrastructure details. Then, the query is evaluated on `input` during each time the plotter object needs to be generated.

The mechanism of tracking the changes in infrastructure and policies is TBD. In the first implementation, infrastructure and policies are loaded at the deployment start and are assumed not to be changed. The future version will track the changes and recompile the query upon the change.

After the query is evaluated, it is parsed into a list of decisions per capability. Then, all decisions on the same capability are merged into one. For example, a decision to deploy read at the workload scope and a decision to deploy read in the workload cluster will result in a single decision to deploy read at the workload scope in the workload cluster. A decision to deploy a capability in clusterA, merged with a decision to deploy this capability in 
any available cluster, will result in a decision to deploy in clusterA.

If the merge process does not succeed to provide a consistent solution, `Valid` will be set to false in `EvaluatorOutput`, and plotter generation will fail.

### Policies

#### Syntax

Policies are written in rego files. Each file declares a package `adminconfig`.

Rules are written in the following syntax: `config[{capability: decision}]` where

`capability` represents a required module capability, such as "read", "write", "transform" and "copy".

`decision` is a JSON structure that matches `Decision` defined above. 

```
{ 
	"policy": {"ID": <id>, "policySetID": <setId>, "description": <description>}, 
	"deploy": <true, false>,
	"restrictions": {
		"modules": <map {key, value}>,
		"clusters": <list of cluster names>,
	},
}
```

#### Policy Set ID

We want to support evaluating different sets of policies for different FybrikApplications. Our recomendation would be to deploy multiple rego files when each file includes policies for a single policy set. A policy decision will return a policy set id along with policy id and desciption. Manager will then match the policy set id of received decisions with the one defined in FybrikApplication and will ignore those that do not match.

If a policy does not specify a policy set id, it will be considered as relevant for all FybrikApplications.
In a similar way, all policies a re relevant for a FybrikApplication that does not specify a policy set id, to support a use-case of a single policy set for all.

#### Out of the box policies

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

#### Extended policies

The extended policies define advanced deployment requirements, such as where read or transform modules should run, what should be the scope of module deployments, and more. 

The policies below are provided as a sample and should be replaced for the production deployment.

```
package adminconfig

config[{"transform": decision}] {
    policy := {"ID": "transform-geo", "description":"Governance based transformations must take place in the geography where the data is stored"}
    clusters := [ data.clusters[i].name | data.clusters[i].metadata.region == input.request.dataset.geography ]
    decision := {"policy": policy, "restrictions": {"clusters": clusters}}
}

config[{"read": decision}] {
    input.request.usage.read == true
    policy := {"ID": "read-scope", "description":"Deploy read at the workload scope"}
    decision := {"policy": policy, "restrictions": {"modules": {"capabilities.scope" : "workload"}}}
}

config[{"read": decision}] {
    input.request.usage.read == true
    policy := {"ID": "read-location", "description":"Deploy read in the workload cluster"}
    decision := {"policy": policy, "restrictions": {"clusters": [ input.workload.cluster.name]}}
}

config[{"copy": decision}] {
    input.request.usage.copy == true
    policy := {"ID": "copy-request", "description":"Copy capability is requested by the user"}
    decision := {"policy": policy, "deploy": true}
}

config[{"copy": decision}] {
    input.request.usage.read == true
    input.request.dataset.geography != input.workload.cluster.region
    count(input.actions) > 0
    clusters :=  [ data.clusters[i].name | data.clusters[i].metadata.region == input.request.dataset.geography ]
    policy := {"ID": "copy-remote", "description":"Implicit copies should be used if the data is in a different region than the compute, and transformations are required"}
    decision := {"policy": policy, "deploy": true, "restrictions": {"clusters": clusters}}
}

config[{"copy": decision}] {
    input.request.usage.read == true
    policy := {"ID": "copy-default", "description":"Implicit copies are allowed in read scenarios"}
    decision := {"policy": policy}
}
```

#### Mechanism for loading policies

Stage 1: policies are provided via files in /tmp/adminconfig/ directory during the control plane deployment.

Stage 2: dynamic load of policies from a configmap. TBD - design a mechanism to track the changes in policies and recompile.

