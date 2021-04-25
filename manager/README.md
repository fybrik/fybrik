# `manager`

Kubernetes [custom resources and controllers](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) of Mesh for Data.

The `manager` binary includes all of the controllers that this project defines but you need to select which of the controllers to run by passing one or more of the following command line arguments:
- `enable-all-controllers` to enable all controllers
- `enable-application-controller` to enable the controller for `M4DApplication`
- `enable-blueprint-controller` to enable the controller for `Blueprint`
- `enable-motion-controller` to enable the controllers for `BatchTransfer` and `StreamTransfer`


## Run and debug locally

Beyond testing, you may run and debug the manager outside the cluster using the following instructions.

### Run required components

Components such as connectors need to be running before you run the manager.
This can be done by one of these options:
1. Running components locally directly (no instructions provided)
2. Running components in a cluster 

For option 2, the Helm installation allows you to pick which compoenents to install. 
Follow the installation guide as usual but in the Helm installation for the control plane add `--set manager.enabled=false` to skip the deployment of the manager. For example:

```bash
helm install m4d charts/m4d --set global.tag=latest --set manager.enabled=false -n m4d-system --wait
```

### Expose running components

Components such as connectors need to be reachable over localhost.
If you chose to run these components in a cluster you can use port-forward.
For example:

```bash
kubectl -n m4d-system port-forward svc/katalog-connector 49152:80 &
kubectl -n m4d-system port-forward svc/opa-connector 49153:80 &
```

### Set configuration environment variables

The main configuration map is not available when running locally.
Therefore, you need to define configuration as environment variables.

Create `.env` file in the root folder of the project. For example:

```bash
VAULT_ADDRESS="http://vault.m4d-system:8200"
MAIN_POLICY_MANAGER_NAME="opa"
MAIN_POLICY_MANAGER_CONNECTOR_URL="localhost:49153"
CATALOG_PROVIDER_NAME="katalog"
CATALOG_CONNECTOR_URL="localhost:49152"
CONNECTION_TIMEOUT="120"
USE_EXTENSIONPOLICY_MANAGER="false"
VAULT_MODULES_ROLE="module"
ENABLE_WEBHOOKS="false"
```

If you plan to run manager from the command line,
then run the following to export all of the variables:

```bash
set -a; . .env; set +a
```

### Run the manager

### From command line

You can now run the manager from the `manager` folder using one of these options:
1. `make run`
2. `go run main.go --enable-all-controllers --metrics-bind-addr=0`

### From IDE

If you wish to debug it from an IDE then be sure to configure the environment variables properly as described in the previous step.

Below is a `launch.json` file for VSCode:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/manager/main.go",
            "envFile": "${workspaceFolder}/.env",
            "args": ["--metrics-bind-addr=0", "--enable-all-controllers"]
        }
    ]
}
```

## Directory structure 

The rest of this README describes the directory structure.

### `apis`

Holds the Customer Resource Definitions (CRDs) of the project:
- `app.m4d.ibm.com/v1alpha1`: Includes `M4DApplication`, administrator APIs `M4DModule` and `M4DBucket`, and internal CRDs `Blueprint` and `Plotter`.
- `motion.m4d.ibm.com/v1alpha1`: Includes data movements APIs `BatchTransfer` and `StreamTransfer`. Usually not used directly but rather invoked as a module.

### `controllers`

Holds the customer controllers of the project:
- `controllers/app` holds the controllers for `app.m4d.ibm.com` APIs `M4DApplication`, `Blueprint` and `Plotter`.
- `controllers/motion` holds the controllers for `motion.m4d.ibm.com` APIs `BatchTransfer` and `StreamTransfer`.

### `testdata`

Includes resources that are used in unit tests and in integration tests.
