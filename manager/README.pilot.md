[![Build Status](https://travis.ibm.com/data-mesh-research/datamesh-cp.svg?token=SFs8yc7zrXxhyzzSs8R8&branch=master)](https://travis.ibm.com/data-mesh-research/datamesh "Travis")

# Pilot

A Kubernetes operator that receives M4DApplication CRD and creates Blueprint CRD. It interacts with Policy Compiler, Catalog Connector and Credential Manager services and leverages M4DModules to determine what components are available for inclusion in the Blueprint.

## Tools

This project uses [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder), follow its [getting started](https://book.kubebuilder.io/quick-start.html) guide

## Contributing

See [CONTRIBUTING](https://github.com/ibm/the-mesh-for-data/blob/master/CONTRIBUTING.md) for details on submitting patches and the contribution workflow.

## Development

Assuming KUBECONFIG is pointing to your cluster you can now run project

```
make
make uninstall
make install

## Deploy modules that implement governance actions
kubectl -f apply samples/modules/<module you want to deploy>

make run
```
## Run using stubs locally - Policy Compiler and Data Catalog Connector

```

### In terminal 1
Install vault - https://learn.hashicorp.com/vault/getting-started/install  
Run vault: vault server -dev 
copy vault root token from results printed by vault

### In terminal 2
make
make uninstall
make install

Set vault env params:
export VAULT_PORT=8200
export VAULT_ADDRESS=http://127.0.0.1:${VAULT_PORT}
export VAULT_TOKEN=s.iJK0al8yaLJxiIN1bwOiuQzL // vault root token copied from terminal 1

### In terminal 3
go run test/services/policycompiler/serverpolicycompiler.go

### In terminal 4
go run test/services/datacatalog/datacatalogstub.go

### In terminal 2
export CATALOG_CONNECTOR_URL=localhost:50085
export CREDENTIALS_CONNECTOR_URL=localhost:50085
export VAULT_DATASET_MOUNT=v1/sys/mounts/m4d/dataset-creds
export VAULT_USER_MOUNT=v1/sys/mounts/m4d/user-creds
export VAULT_DATASET_HOME=m4d/dataset-creds/
export VAULT_USER_HOME=m4d/user-creds/
export CONNECTION_TIMEOUT=120
export MAIN_POLICY_MANAGER_CONNECTOR_URL=localhost:50090
export MAIN_POLICY_MANAGER_NAME="MOCK" 
export USE_EXTENSIONPOLICY_MANAGER=false
make run
```


## Deploy M4DApplication
```
cd manager/testdata/e2e
kubectl apply -f module-implicit-copy-db2wh-to-s3.yaml
kubectl apply -f module-implicit-copy-kafka-to-s3-stream.yaml
kubectl apply -f module-read.yaml
kubectl apply -f bucket-available.yaml
kubectl apply -f m4dapplication.yaml -n default
```


### Alternative vault approach - use instance in open shift
If you want to use a vault instance that is already running in the remote open shift cluster, then do the following:
    export VAULT_PORT=8200
    export VAULT_ADDRESS=http://127.0.0.1:${VAULT_PORT}
    kubectl port-forward service/vault -n vault ${VAULT_PORT}:8200 &
    export VAULT_TOKEN=$(kubectl get secrets vault-unseal-keys -n vault -o jsonpath={.data.vault-root} | base64 --decode)
