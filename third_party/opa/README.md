# Example

### Add a policy

```bash
kubectl create configmap sample_policies.rego --from-file=sample_policies.rego -n m4d-system -o yaml --dry-run=client | kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego | kubectl apply -f -

kubectl create configmap sample_policies.rego  --from-file=sample_policies.rego --from-file=data_policies/action_struct.rego --from-file=data_policies/auditing_struct.rego --from-file=data_policies/helper_functions.rego --from-file=data_policies/input_reader.rego --from-file=data_policies/taxonomies_unification.rego --from-file=data_policies/verify_correct_input.rego --from-file=data_policies/medical_taxonomies.json -n m4d-system -o yaml --dry-run=client | kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego | kubectl apply -f -

kubectl create configmap sample_policies.rego  --from-file=main=sample_policies.rego --from-file=main=data_policies/action_struct.rego --from-file=main=data_policies/auditing_struct.rego --from-file=main=data_policies/helper_functions.rego --from-file=main=data_policies/input_reader.rego --from-file=main=data_policies/taxonomies_unification.rego --from-file=main=data_policies/verify_correct_input.rego --from-file=main=data_policies/medical_taxonomies.json -n m4d-system -o yaml --dry-run=client | kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego | kubectl apply -f -



kubectl create configmap sample_policies.rego  --from-file=data_policies -n m4d-system -o yaml --dry-run=client | kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego  --local openpolicyagent.org/data=opa | kubectl apply -f -






kubectl create configmap sample_policies.rego --from-file=sample_policies.rego -n m4d-system -o yaml --dry-run=client | kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego | kubectl apply -f -

kubectl create configmap actionstruct.rego --from-file=data_policies/action_struct.rego -n m4d-system -o yaml --dry-run=client | kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego | kubectl apply -f -

kubectl create configmap auditingstruct.rego --from-file=data_policies/auditing_struct.rego -n m4d-system -o yaml --dry-run=client | kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego | kubectl apply -f -

kubectl create configmap helperfunctions.rego --from-file=data_policies/helper_functions.rego -n m4d-system -o yaml --dry-run=client | kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego | kubectl apply -f -

kubectl create configmap inputreader.rego --from-file=data_policies/input_reader.rego -n m4d-system -o yaml --dry-run=client | kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego | kubectl apply -f -

kubectl create configmap taxonomiesunification.rego --from-file=data_policies/taxonomies_unification.rego -n m4d-system -o yaml --dry-run=client | kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego | kubectl apply -f -

kubectl create configmap verifycorrectinput.rego --from-file=data_policies/verify_correct_input.rego -n m4d-system -o yaml --dry-run=client | kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego | kubectl apply -f -

kubectl create configmap medicaltaxonomies.json --from-file=data_policies/medical_taxonomies.json -n m4d-system -o yaml --dry-run=client | kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego | kubectl apply -f -









-- latest try 
kubectl kustomize ./ |kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/policy=rego > newpolicy.yaml 

kubectl kustomize ./ |kubectl label -f- --dry-run=client -o yaml --local openpolicyagent.org/data=opa > newjson.yaml 


```

Or directly with `kubectl` using `policy.yaml`:
```bash
kubectl apply -f policy.yaml -n m4d-system
```

### Port forward OPA

```bash
kubectl port-forward deployment/opa 8181
```

### Send an OPA query

```bash
curl localhost:8181/v1/data/katalog/example/verdict -d @input.json -H 'Content-Type: application/json'
```

The expected output is
```json
{"result":[{"action":"RedactColumn","columns":["nameOrig"],"name":"Redact PII columns for CBA"}]}
```
