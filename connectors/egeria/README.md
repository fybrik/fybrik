# Egeria Connector

For each of the tasks related to the egeria connector, execute the corresponding commands from this directory.


## Build the connector

```bash
make build
```

Cleanup with `make clean`

## Build and push the connector image

Set the following environment variables to point to a container registry: `DOCKER_USERNAME`, `DOCKER_PASSWORD`, `DOCKER_HOSTNAME` (defaults to "ghcr.io"), `DOCKER_NAMESPACE` (defaults to "mesh-for-date"), `DOCKER_TAGNAME` (defaults to "latest").
Then run:

```bash
make docker-build docker-push
```

Cleanup with `make clean docker-rmi`

## Running tests

```bash
make test
```

## Running locally

To run the connector locally the `EGERIA_SERVER_URL` environment variable needs to be set to a URL of a running Egeria server. We recommend to create a file named `.env` in the root directory of the project and set the variable there. For example:

```bash
EGERIA_SERVER_URL=https://localhost:9443
```

To run the connector locally:

```bash
make run
```

To terminate the process:

```bash
make terminate
```

Note that by default the connector binds to port 50084.
You can change it by setting the `PORT_EGERIA_CONNECTOR` environment variable (e.g., adding it to `.env`).
