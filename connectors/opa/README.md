# OPA Connector

All commands in this document should be executed from this directory unless explicitly stated otherwise.

## Testing

```bash
make test
```

## Build and push the connector image

Set the following environment variables to point to a container registry: `DOCKER_USERNAME`, `DOCKER_PASSWORD`, `DOCKER_HOSTNAME` (defaults to "ghcr.io"), `DOCKER_NAMESPACE` (defaults to "mesh-for-data"), `DOCKER_TAGNAME` (defaults to "latest").
Then run:

```bash
make docker-build docker-push
```

Cleanup with `make docker-rmi`


## Running in a cluster

The connector is deployed by default as part of the Mesh for Data Helm chart. To use a [locally built image](#build-and-push-the-connector-image) add the following to the installation of the m4d chart:

```bash
--set opaConnector.image=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/opa-connector:${DOCKER_TAGNAME}
```

To run independantly of manager you need to set some environment variables:

1. `CATALOG_CONNECTOR_URL`: A URL to a catalog connector
2. `CONNECTION_TIMEOUT`: Connection timeout in seconds
3. `KUBE_NAMESPACE`: target namespace (defaults to "m4d-system")

We recommend to create a file named `.env` in the root directory of the project and set all variables there. For example:

```s
CATALOG_CONNECTOR_URL="katalog-connector:80"
CONNECTION_TIMEOUT=120
KUBE_NAMESPACE="m4d-system"
```

Deploy an OPA server and a connector to it:

```bash
make deploy
```

Cleanup with `make undeploy`

## Running locally

### Run OPA server

```bash
make opaserver
```

Termnate with `make opaserver-terminate` and cleanup with `make opaserver-clean`.

### Build the connector

```bash
make build
```

Cleanup with `make clean`


### Run the connector

Set environment variables:

1. `OPA_SERVER_URL`: a URL to a running OPA service
2. `CATALOG_CONNECTOR_URL`: A URL to a catalog connector
3. `CONNECTION_TIMEOUT`: Connection timeout in seconds
4. `PORT_OPA_CONNECTOR`: port to bind to (defaults to 50082)

We recommend to create a file named `.env` in the root directory of the project and set all variables there. For example:

```s
OPA_SERVER_URL="localhost:8181"
CATALOG_CONNECTOR_URL="localhost:50084"
CONNECTION_TIMEOUT=120
PORT_OPA_CONNECTOR=50082
```

Run the connector:

```bash
make run
```

Termnate with `make terminate`.

Alternatively run directly with `go run main.go` after exporting all required environment variables.
