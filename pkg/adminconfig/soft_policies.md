Soft policies and infrastructure attributes
===========================================

## Summary

Soft policies are used for data plane optimization. A soft policy focuses on one or several infrastructure metrics, e.g. storage cost or bandwidth between two regions.
It provides a general directive such as minimize or maximize, and a relative weght given the workload and dataset information.  

### Infrastructure attributes

Infrastructure attributes are defined in `infrastructure.json` in /tmp/adminconfig directory. 

```
type InfrastructureElement struct {
	// Attribute name, e.g. storage-cost, defined in the taxonomy
	Attribute   taxonomy.Attribute    `json:"attribute"`
	// Description
	Description string                `json:"description,omitempty"`
	// Value type (numeric, string, boolean)
	Type        AttributeType         `json:"type"`
	// Attribute value
	Value       string                `json:"value"`
	// Units (defined in the taxonomy)
	Units       taxonomy.Units        `json:"units,omitempty"`
	// A resource defined by the attribute ("storageaccount","module","cluster")
	Object 		string				  `json:"object,omitempty"`
	// A reference to the resource instance, e.g. storage account name
	Instance    string                `json:"instance,omitempty"`
	// A scale of values (minimum and maximum) when applicable
	Scale       *RangeType            `json:"scale,omitempty"`
	// A list of arguments defining a specific metric, e.g. regions for a bandwidth
	Arguments   []string			  `json:"arguments,omitempty"`
}
```

Attribute examples:

```
{
    "attribute": "storage-cost",
    "description": "theshire object store",
    "value": "90",
    "type": "numeric",
    "units": "US Dollar per TB per month",
    "object": "storageaccount",
    "instance": "account-theshire"
}

{
    "attribute": "bandwidth",
    "description": "bandwidth between neverland and theshire",
    "value": "5",
    "type": "numeric",
    "units": "Mbp/s",
    "arguments": ["theshire","neverland"]
}
```

## Soft policy

### Syntax

Policies are written in rego files. Each file declares a package `adminconfig`.

Rules are written in the following syntax: `optimize[decision}]` where

`decision` is a JSON object that includes a `DecisionPolicy` and a list of `AttributeOptimization` objects defined above. 

```
{ 
    "attribute": <infrastructure attribute name>,
    "directive": <"min", "max">,
    "weight": <a number between 0 and 1> 
}
```

### Examples

```
package adminconfig

optimize[decision] {
    input.workload.properties.priority == "high"
    policy := {"ID": "001", "description":"Focus on high performance"}
    decision := {"policy": policy, "strategy": [{"attribute": "bandwidth", "directive": "max"}]}
}

optimize[decision] {
    input.workload.properties.priority == "medium"
    input.workload.properties.stage == "PROD"
    policy := {"ID": "002", "description":"Save storage costs and minimize latency"}
    optimize_storage := {"attribute": "storage-cost", "directive": "min", "weight": "0.6"}
    optimize_latency := {"attribute": "bandwidth", "directive": "max", "weight": "0.4"}
    decision := {"policy": policy, "strategy": [{optimize_storage, optimize_latency}]}
}
```

### Weights

- Each attribute defined in a specific policy has a weight. When a policy refers to multiple attributes, 
an optimization strategy for the given context is defined by prioritizing some attributes over the others according to the defined weights.  
- A default value is 1.0 
- All weights in a decision are normalized, i.e. weights 1, 0.5 and 0.5 result in weights 0.5, 0.25 and 0.25 respectively.

### Conflict resolution

In case of more than one optimization strategy for a given context, the first one takes precedence, and the others are ignored. 
In the future version there will be a priority assigned to a rule, and the strategy with the highest priority will be chosen.