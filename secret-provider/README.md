# SecretProvider

## API
#### Definitions
- `secret_name` is a path to a secret in vault. For example: `/v1/secert/cos`.
- `role` is a role to assume. For example: `demo`. As part of authentication in vault, the user assumes a role, and supplies credentials for authentication (a JWT token in our case). Each role has a set of policies atteched to it, which dictates its capabilities inside vault (for example read certain secrets, etc).
- `jwt` is a valid k8s service account token, for authentication against vault (for assuming the `role`).

#### API Endpoints
- `/get-secret?secret_name=<secret_name>&role=<role>` - Returns a raw secret from vault. For example, a JDBC password. Returns it as json, the same json stored inside vault as the secrt (as we are using the KV secret backend of vault).
- `/get-iam-token/secret_name=<secret_name>&role=<role>` - Returns an IAM JWT token. Behind the scene the server first authenticates to vault to obtain an API key, which is then used to authenticate against the IAM.

### Deploy and Test
```bash
# Deploy Vault and the Secret-Provider
OPENSHIFT=<1 / 0> make deploy

# Deploy the sleep pod (deploys to the default ns)
kubectl apply -f deploy/testing/sleep.yaml

# Wait for all the pods in the m4d-system namespace to become ready before moving on to the next step.

# Configure Vault:
#   Configure k8s auth method
#   Create a role for the secret-provider, for authorization using the k8s auth method
#   Enable the kv engine
#   Create a policy to govern the /v1/secret path

# PORT_TO_FORWARD - port number for port-forwording, as part of configuring Vault
# DATA_PROVIDER_USERNAME - username of the data-provider, for the userpass auth method. Default is data_provider.
# DATA_PROVIDER_PASSWORD - password of the data-provider, for the userpass auth method. Default is password.
PORT_TO_FORWARD=<port-number> DATA_PROVIDER_USERNAME=<username> DATA_PROVIDER_PASSWORD=<password> make configure-vault

# Store some secrets inside vault
# This also stores api-key inside vault, assuming APIKEY is an environment varibale
PORT_TO_FORWARD=<port-number> make vault-demo-secrets

# Exec to the sleep pod and curl the get-secret endpoint
kubectl exec $(kubectl get pod -l app=sleep -o jsonpath={.items..metadata.name}) -it -- curl 'http://secret-provider.m4d-system:5555/get-secret?role=demo&secret_name=%2Fv1%2Fsecret%2Fcos' -w "\n"
{"api_key": <api-key>}

# Exec to the sleep pod and curl the get-iam-token endpoint
kubectl exec $(kubectl get pod -l app=sleep -o jsonpath={.items..metadata.name})  -it -- curl 'http://secret-provider.m4d-system:5555/get-iam-token?role=demo&secret_name=%2Fv1%2Fsecret%2Fcos' -w "\n"
<iam-token>

# Now interact with the "external" endpoint which mimics external vault installation, assuming the Vault CLI is installed
export VAULT_ADDR="http://127.0.0.1:<PORT_TO_FORWARD>"

# Defaults:
# DATA_PROVIDER_USERNAME = data_provider
# DATA_PROVIDER_PASSWORD = password
vault login -method=userpass username=<DATA_PROVIDER_USERNAME> password=<DATA_PROVIDER_PASSWORD>

# Get allowed secrets
vault kv get external/some-secret
<some data>

# Try to access secrets which are not allowed
vault kv get secret/some-secret
<some error message>

# Undeploy
OPENSHIFT=<1 / 0> make undeploy
```
