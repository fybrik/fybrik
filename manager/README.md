# `manager`

Kubernetes [custom resources and controllers](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) of Fybrik.

The `manager` binary includes all of the controllers that this project defines but you need to select which of the 
controllers to run by passing one or more of the following command line arguments:
- `enable-all-controllers` to enable all controllers
- `enable-application-controller` to enable the controller for `FybrikApplication`
- `enable-blueprint-controller` to enable the controller for `Blueprint`


## Run and debug locally

Beyond testing, you may run and debug the manager outside the cluster using the following instructions.

### Prepare the fybrik environment
Use one of below methods to prepare the fybrik environment.

1. Install Fybrik using the [Quick Start](https://fybrik.io/dev/get-started/quickstart/) guide.

2. Follow [the instructions](../pipeline/README.md) to use tekton pipeline to deploy the fybrik components to your 
existing cluster.

### Run required components

Components such as connectors need to be running before you run the manager.
This can be done by one of these options:
1. Running components locally directly (no instructions provided)
2. Running components in a cluster 

For option 2, the Helm installation allows you to pick which components to install. 
Follow the [installation guide](https://fybrik.io/dev/get-started/quickstart/) as usual,the fybrik crd needs to be 
installed before installing fybrik as is, but in the Helm installation for the control plane 
add `--set manager.enabled=false` to skip the deployment of the manager. For example:

```bash
helm install fybrik charts/fybrik --set global.tag=master --set manager.enabled=false -n fybrik-system --wait
```

If your are using the local development images please use the `0.0.0` tag:

```bash
helm install fybrik charts/fybrik --set global.tag=0.0.0 --set manager.enabled=false -n fybrik-system --wait
```

### Expose running components

Components such as connectors need to be reachable over localhost.
If you chose to run these components in a cluster you can use port-forward.
For example:

```bash
kubectl -n fybrik-system port-forward svc/katalog-connector 49152:80 &
kubectl -n fybrik-system port-forward svc/opa-connector 49153:80 &
```

### Set configuration environment variables

The main configuration map is not available when running locally.
Therefore, you need to define configuration as environment variables.

Create `.env` file in the root folder of the project. For example:

```bash
ClusterName="thegreendragon"
Zone="hobbiton"
VaultAuthPath="kind"
Region="theshire"
VAULT_ADDRESS="http://vault.fybrik-system:8200"
MAIN_POLICY_MANAGER_NAME="opa"
MAIN_POLICY_MANAGER_CONNECTOR_URL="http://localhost:49153"
CATALOG_PROVIDER_NAME="katalog"
CATALOG_CONNECTOR_URL="http://localhost:49152"
VAULT_MODULES_ROLE="module"
ENABLE_WEBHOOKS="false"
VAULT_ENABLED="true"
DATA_DIR="/tmp"
```

If the manager works with a Razee service, you also need to add the following environment variables:

```bash
RAZEE_URL=<Razee access point> # e.g. "http://localhost:3333/graphql"
# you should define either RAZEE_USER/RAZEE_PASSWORD or API_KEY, if both are defined, RAZEE_USER/RAZEE_PASSWORD will be used.
RAZEE_USER=<Razee user>
RAZEE_PASSWORD=<Razee password> 
API_KEY=<Razee api key>
```

If you plan to run manager from the command line,
then run the following to export all of the variables:

```bash
set -a; . .env; set +a
```

### Copy taxonomy JSON files and config policies locally
```bash
cp -R ../charts/fybrik/files/taxonomy /tmp/
cp -R ../charts/fybrik/files/adminconfig /tmp/
```

### Run the manager

### From command line

You can now run the manager from the `manager` folder using one of these options:
1. `make run`
2. `go run main.go --enable-all-controllers --metrics-bind-addr=0 --health-probe-addr=127.0.0.1:8088`

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
            "args": ["--metrics-bind-addr=0", "--health-probe-addr=127.0.0.1:8088", "--enable-all-controllers"]
        }
    ]
}
```

## Directory structure 

The rest of this README describes the directory structure.

### `apis`

Holds the Customer Resource Definitions (CRDs) of the project:
- `app.fybrik.io/v1beta1`: Includes `FybrikApplication`, administrator APIs `FybrikModule` and `FybrikBucket`, and internal CRDs `Blueprint` and `Plotter`.

### `controllers`

Holds the customer controllers of the project:
- `controllers/app` holds the controllers for `app.fybrik.io` APIs `FybrikApplication`, `Blueprint` and `Plotter`.

### `testdata`

Includes resources that are used in unit tests and in integration tests.
