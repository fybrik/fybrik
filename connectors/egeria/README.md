# Egeria Connector

All commands in this document should be executed from this directory unless explicitly stated otherwise.


## Testing

```bash
make test
```

## Build and push the connector image

Set the following environment variables to point to a container registry: `DOCKER_USERNAME`, `DOCKER_PASSWORD`, `DOCKER_HOSTNAME` (defaults to "ghcr.io"), `DOCKER_NAMESPACE` (defaults to "fybrik"), `DOCKER_TAGNAME` (defaults to "latest").
Then run:

```bash
make docker-build docker-push
```

Cleanup with `make clean docker-rmi`

## Running in a cluster

The connector can be deployed as part of the Fybrik Helm chart by adding the following to the installation of the fybrik chart (replace the serverURL if the value is different):

```bash
--set coordinator.catalog=egeria --set egeriaConnector.serverURL="https://egeria-platform.egeria-catalog:9443"
```

To use a [locally built image](#build-and-push-the-connector-image) add the following to the installation of the fybrik chart:

```bash
--set egeriaConnector.image=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/egr-connector:${DOCKER_TAGNAME}
```

To run independantly of manager you need to set some environment variables:

1. `EGERIA_SERVER_URL`: A URL to a running egeria server
2. `KUBE_NAMESPACE`: target namespace (defaults to "fybrik-system")

We recommend to create a file named `.env` in the root directory of the project and set all variables there. For example:

```s
EGERIA_SERVER_URL="https://egeria-platform.egeria-catalog:9443"
KUBE_NAMESPACE="fybrik-system"
```

Deploy the connector:

```bash
make deploy
```

Cleanup with `make undeploy`

## Running locally

### Build the connector

```bash
make build
```

Cleanup with `make clean`

### Run the connector

Set environment variables:

1. `EGERIA_SERVER_URL`: a URL to a running Egeria server
2. `PORT_EGERIA_CONNECTOR`: port to bind to (defaults to 50084)

We recommend to create a file named `.env` in the root directory of the project and set all variables there. For example:

```s
EGERIA_SERVER_URL="https://localhost:9443"
PORT_EGERIA_CONNECTOR=50084
```

Run the connector:

```bash
make run
```

Termnate with `make terminate`.
