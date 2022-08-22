# Infrastructure attributes

When writing configuration policies, infrastructure metrics and costs may also be taken into account in order to optimize the generated data plane. 
For example, selection of a storage account may be based on a storage cost, selection of a cluster may provide a restriction on cluster capacity, and so on. 
Collection of the metrics and their dynamic update is beyond the scope of Fybrik. One may develop or use 3rd party solutions for monitoring and updating these infrastructure metrics.
Infrastructure attributes are stored in the `/tmp/adminconfig/infrastructure.json` directory of the manager pod. 

### Metric metadata

Prior to defining an infrastructure attribute, the corresponding metric should be defined, providing information about the attribute value, e.g. the measurement units and the scale of possible values.
Several attributes may share the same metric, e.g. `rate` can be defined for both the `error-rate` and the `load-rate`.
Example of a metric:
```
"name": "rate",
"type": "numeric",
"units": "%",
"scale": {"min": 0, "max": 100}
```

### How to define infrastructure attributes

An infrastructure attribute is defined by a JSON object that includes the following fields:

- `attribute` - name of the infrastructure attribute, should be defined in the taxonomy
- `description` 
- `metricName` - a reference to the [metric](#metric-metadata)
- `value` - the actual value of the attribute
- `object` - a resource the attribute relates to (storageaccount, module, cluster)
- `instance` - a reference to the resource instance, e.g. storage account name

The infrastructure attributes are associated with resources managed by Fybrik: FybrikStorageAccount, FybrikModule and cluster (defined in the `cluster-metadata` config map). The valid values for the attribute `object` field are `storageaccount`, `module` and `cluster`, respectively.

For example, the following attribute defines the storage cost of the "account-theshire" storage account. 

```
{
    "attribute": "storage-cost",
    "description": "theshire object store",
    "value": "90",
    "metricName": "cost",
    "object": "storageaccount",
    "instance": "account-theshire"
}
```

### Add a new attribute definition to the taxonomy

See [metric taxonomy](https://github.com/fybrik/fybrik/blob/master/samples/taxonomy/example/infrastructure/attributepair.yaml) for an example how to define an attribute and the corresponding measurement units. 
 
### Usage of infrastructure attributes in configuration policies

An infrastructure attribute can be used as the `property` value in [configuration policies](../concepts/config-policies.md#configuration-policies). For example, the following policy restricts 
the storage account selection using the `storage-cost` infrastructure attribute:
```
# restrict storage costs to a maximum of $95 when copying the data
config[{"capability": "copy", "decision": decision}] {
    input.request.usage == "copy"
    input.request.dataset.geography != input.workload.cluster.metadata.region
    account_restrict := {"property": "storage-cost", "range": {"max": 95}}
    policy := {"ID": "copy-restrict-storage", "description":"Use cheaper storage", "version": "0.1"}
    decision := {"policy": policy, "restrictions": {"storageaccounts": [account_restrict]}}
}
```

### Usage of infrastructure attributes for optimization

Attributes can also be used for [optimizing](../concepts/config-policies.md#optimization-goals) a control plane. For example, the following rule reduces storage costs:

```
# minimize storage cost
optimize[decision] {
    policy := {"ID": "save-cost", "description":"Save storage costs", "version": "0.1"}
    decision := {"policy": policy, "strategy": [{"attribute": "storage-cost", "directive": "min"}]}
}
```