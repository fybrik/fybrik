##### Setup

Option 1 - connect to Vault instance inside a k8s cluster
```
export VAULT_ADDR=http://127.0.0.1:8200
kubectl port-forward -n m4d-system service/vault 8200:8200 &
export VAULT_TOKEN=$(kubectl get secrets vault-unseal-keys -n m4d-system -o jsonpath={.data.vault-root} | base64 --decode)
```

Option 2 - deploy vault locally
```
vault server -dev &
export VAULT_ADDR="http://127.0.0.1:8200"
vault login <root-token>
```

##### Full flow - authentication with k8s service-account
```
Based on: https://learn.hashicorp.com/vault/identity-access-management/vault-agent-k8s

# get environment varibales for configuring vault for communication with the API server
export VAULT_SA_NAME=$(kubectl get sa vault -n m4d-system -o jsonpath="{.secrets[0]['name']}")
export SA_JWT_TOKEN=$(kubectl get secret $VAULT_SA_NAME -n m4d-system -o jsonpath="{.data.token}" | base64 --decode)
export SA_CA_CRT=$(kubectl get secret $VAULT_SA_NAME -n m4d-system -o jsonpath="{.data['ca\.crt']}" | base64 --decode)

vault auth enable kubernetes

vault write auth/kubernetes/config \
        token_reviewer_jwt="$SA_JWT_TOKEN" \
        kubernetes_host="https://kubernetes.default.svc:443" \
        kubernetes_ca_cert="$SA_CA_CRT"

vault policy write test-1 - <<EOF
path "secret/*" {
    capabilities = ["read", "list"]
}
EOF

vault write auth/kubernetes/role/demo \
        bound_service_account_names=secret-provider \
        bound_service_account_namespaces=m4d-system \
        policies=test-1 \
        ttl=24h

# List enabled auth methods
vault auth list

# Read the configuration of given role
vault read auth/kubernetes/role/demo

# Read the configuration of the kubernetes authentication method
vault read auth/kubernetes/config
```

##### Put and read secrets
```
vault secrets enable -path=secret -version=1 kv

vault kv put secret/cos api_key=abcdefgh12345678
vault kv put secret/db2 username=tomer password=s3cr3t

vault kv get secret/db2
vault kv get secret/cos

# List the secrets under the secret path
vault kv list secret
```

##### Teardown
```
# If used kubectl port-forward delete it in the end with:
kill -9 %%
```