# Using OPA for Data Governance
[Open Policy Agent](https://www.openpolicyagent.org/) may be used as a data governance policy engine with Fybrik via the [connector](https://fybrik.io/dev/concepts/connectors/) mechanism.

When OPA is used for data governance, it is deployed as a stand-alone service.  Policies are defined in rego and uploaded to OPA.

For more details on OPA policies please refer to [OPA documentation](https://www.openpolicyagent.org/docs/latest/policy-language/) in particulate to [the basics](https://www.openpolicyagent.org/docs/latest/policy-language/#the-basics) section which explains how a policy is evaluated.

## Fybrik Default Policies

Fybrik ***denys by default*** any request if no rule is triggered. This behavior can be changed to ***allow by default*** by creating the following rule and upload it to OPA using methods described in this page:

```yaml
package dataapi.authz

rule [{}] { true }
```

You can also add conditions like

```yaml
 rule[{}] {
  // conditions here
 }
```

The verdict `allow` will be reached only if the conditions hold, and no other rule has been triggered, e.g. a rule requiring column redaction.

## Managing OPA policies

There are [several ways](https://www.openpolicyagent.org/docs/latest/management/) to manage policies and data of the OPA service. 

One simple approach is to use [OPA kube-mgmt](https://github.com/open-policy-agent/kube-mgmt) and manage Rego policies in Kubernetes `Configmap` resources. By default, Fybrik installs OPA with kube-mgmt enabled. 

The following two sections show how to use OPA with kube-mgmt.

!!! warning 
    
    Due to size limits you must ensure that each configmap is smaller than 1MB when base64 encoded.

### Using a configmap YAML

1. Create a configmap with a Rego policy and a `openpolicyagent.org/policy=rego` label in the `fybrik-system` namespace:
    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: <policy-name>
      namespace: fybrik-system
      labels:
        openpolicyagent.org/policy: rego
    data:
      main: |
      <you rego policy here>
    ```
1. Apply the configmap:
    ```bash
    kubectl apply -f <policy-name>.yaml
    ```
1. To remove the policy just remove the configmap:
   ```bash
    kubectl delete -f <policy-name>.yaml
   ```

### Using a Rego file

You can use `kubectl` to create a configmap from a Rego file. To create a configmap named `<policy-name>` from a Rego file in path `<policy-name.rego>`:

```bash
kubectl create configmap <policy-name> --from-file=main=<policy-name.rego> -n fybrik-system
kubectl label configmap <policy-name> openpolicyagent.org/policy=rego -n fybrik-system
```

Delete the policy with `kubectl delete configmap <policy-name> -n fybrik-system`.


### Using `opaServer.bootstrapPolicies` field

Another method to upload policies to OPA is to write them as `opaServer.bootstrapPolicies` field in [values.yaml](https://raw.githubusercontent.com/fybrik/charts/master/charts/fybrik/values.yaml) file used for the Fybrik deployment.
In this approach the policies are uploaded upon OPA startup.

```bash
opaServer:
  # Bootstrap policies to load upon startup
  bootstrapPolicies:
    allowSamplePolicy: |-
      package dataapi.authz

      rule [{}] { true }
```
