Component for Config Policy Decisions
==========================

## Summary

Design a component that will evaluate config policy decisions based on the fybrikapplication spec and infrastructure (e.g. deployed clusters).

## Goals

1. Design an interface between the config policy evaluator and the manager.

2. Design an interface between the config policy evaluator and OPA

## Interface with the manager

### Input (`adminconfig.EvaluatorInput`)

A dynamic input constructed per a FybrikApplication, per a single dataset.
It provides general application data such as workload cluster and application properties, as well as dataset details (user requirements, metadata, required actions.
```
// WorkloadInfo holds workload details such as cluster/region, type, etc.
type WorkloadInfo struct {
	// Cluster where the user workload is running
	Cluster multicluster.Cluster
}

// Request is a request to use a specific asset
type Request struct {
	// asset identifier
	DatasetID string
	// requested interface
	Interface api.InterfaceDetails
	// requested usage, e.g. "read": true, "write": false
	Usage map[api.DataFlow]bool
}

// EvaluatorInput is an input to Configuration Policies Evaluator.
// Used to evaluate configuration policies.
type EvaluatorInput struct {
	// Workload configuration
	Workload WorkloadInfo
	// Application properties
	AppInfo api.ApplicationDetails
	// Asset metadata
	AssetMetadata *assetmetadata.DataDetails
	// Requirements for asset usage
	AssetRequirements Request
	// Governance Actions for reading data (relevant for read scenarios only)
	GovernanceActions []model.Action
}
```
On top of the dynamic input based on a specific FybrikApplication, an `Infrastructure` object is used by the config policy evaluator to get deployment infrastructure information: available clusters, available storage, bandwidth metrics, etc. Evaluator uses `InfrastructureManager` to obtain the `Infrastructure` data.

In the current implementation, `Infrastructure` is defined as:
```
type Infrastructure struct {
	// Clusters available for deployment
	Clusters []multicluster.Cluster
}
```

### Output (`adminconfig.EvaluatorOutput`)

The result of evaluating config policies on a given dataset contains deployment `Decision` for each capablility, after merging all relevant decisions for the capability.
 A conflict during evaluation result is marked by `Valid == false`
 
```
// EvaluatorOutput is an output of ConfigurationPoliciesEvaluator.
// Used by manager to decide which modules are deployed and in which cluster.
type EvaluatorOutput struct {
	// Valid is true when there is no conflict between the decisions, and false otherwise
	Valid bool
	// Dataset identifier
	DatasetID string
	// Decisions per capability (after being merged)
	ConfigDecisions map[api.CapabilityType]Decision
}
```

`Decision` is a result of evaluating a configuration policy which satisfies the predicates of the policy.
`Deploy` is a deployment decision (require, forbid or allow).
`Restrictions` restricts the choice of the modules to be deployed, deployment clusters, storage accounts, etc.
`Justifications` provides a full list of policies that have been evaluated.

```
type Restrictions struct {
	// Restrictions on clusters used for deployment
	Clusters []string
	// Restrictions on modules of the type “key”: “value” when the key is a module property (e.g. scope, type) and the value is an allowed value (e.g. asset, plugin)
	ModuleRestrictions map[string]string
}

type Decision struct {
	// a decision regarding deployment: True = require, False = forbid, Unknown = allow
	Deploy corev1.ConditionStatus
	// Deployment restrictions on modules, clusters and additional resources
	Restrictions DeploymentRestrictions
	// Descriptions of policies that have been triggered during evaluation
	Jusifications []string
}
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

### Policies

#### Syntax

TBD

#### Mechanism for loading policies

TBD - file(s) provided during deployment or a config map?

#### Merging policy decisions

TBD