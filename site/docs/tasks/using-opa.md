# Using OPA

There are [several ways](https://www.openpolicyagent.org/docs/latest/management/) to manage policies and data of the OPA service. 

One simple approach is to use [OPA kube-mgmt](https://github.com/open-policy-agent/kube-mgmt) and manage Rego policies in Kubernetes `Configmap` resources. By default Fybrik installs OPA with kube-mgmt enabled. 

This task shows how to use OPA with kube-mgmt.

!!! warning 
    
    Due to size limits you must ensure that each configmap is smaller than 1MB when base64 encoded.

## Using a configmap YAML

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

## Using a Rego file

You can use `kubectl` to create a configmap from a Rego file. To create a configmap named `<policy-name>` from a Rego file in path `<policy-name.rego>`:

```bash
kubectl create configmap <policy-name> --from-file=main=<policy-name.rego> -n fybrik-system
kubectl label configmap <policy-name> openpolicyagent.org/policy=rego -n fybrik-system
```

Delete the policy with `kubectl delete configmap <policy-name> -n fybrik-system`.
