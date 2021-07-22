# Lifecycle Management

To install OPA, execute
```bash
helm install fybrik-opa  ../../charts/fybrik --set manager.enabled=false --set opaServer.enabled=true
```
To view the OPA installation template, execute
```bash
helm template fybrik-opa  ../../charts/fybrik --set manager.enabled=false --set opaServer.enabled=true
```

To uninstall OPA, execute
```bash
helm uninstall fybrik-opa
```


# Policy Management

### Add a policy to OPA
```bash
make loadpolicy ARGS=<POLICYFOLDER>

Example: make loadpolicy ARGS=data-and-policies/user-created-policy-1
```

### Remove a policy from OPA
```bash
make unloadpolicy ARGS=<POLICYFOLDER>

Example: make unloadpolicy ARGS=data-and-policies/user-created-policy-1
```

### Add a policy data folder to OPA
```bash
make loaddata ARGS=<POLICYDATAFOLDER>

Example: make loaddata ARGS=data-and-policies/fybrik-external-data
```

### Remove a policy data folder from OPA
```bash
make unloaddata ARGS=<POLICYDATAFOLDER>

Example: make unloaddata ARGS=data-and-policies/fybrik-external-data
```

# Example

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
