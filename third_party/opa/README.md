# Example

### Deploy OPA
```bash
make deploy
```

### Add a policy
```bash
make loadpolicy ARGS=<POLICYFOLDER>

Example: make loadpolicy ARGS=data-and-policies/user-created-policy-1
```

### Remove a policy
```bash
make unloadpolicy ARGS=<POLICYFOLDER>

Example: make unloadpolicy ARGS=data-and-policies/user-created-policy-1
```

### Add a policy data folder
```bash
make loaddata ARGS=<POLICYDATAFOLDER>

Example: make loaddata ARGS=data-and-policies/meshfordata-external-data
```

### Remove a policy data folder
```bash
make unloaddata ARGS=<POLICYDATAFOLDER>

Example: make unloaddata ARGS=data-and-policies/meshfordata-external-data
```

### UnDeploy OPA
```bash
make undeploy
```

### Port forward OPA

```bash
kubectl port-forward -n  <OPA_NAMESPACE> deployment/opa 8181
```

### Send an OPA query

```bash
curl localhost:8181/v1/data/dataapi/authz/transform -d @input-READ.json -H 'Content-Type: application/json'
```

The expected output is
```json
{"result":[{"action_name":"redact column","arguments":{"column_name":"nameDest::6"},"description":"Single column is obfuscated with XXX instead of values","used_policy":{"description":"test for transactions dataset that redacts some columns by name"}},{"action_name":"redact column","arguments":{"column_name":"nameOrig::3"},"description":"Single column is obfuscated with XXX instead of values","used_policy":{"description":"test for transactions dataset that redacts some columns by name"}}]}
```

```bash
curl localhost:8181/v1/data/dataapi/authz/transform -d @input-WRITE.json -H 'Content-Type: application/json'
```

The expected output is
```json
{"result":[{"action_name":"redact column","arguments":{"column_name":"CUSTOMER_ID"},"description":"Single column is obfuscated with XXX instead of values","used_policy":{"description":"Columns with Confidential tag to be redacted before read action"}}]}
```
