# Vault Connector
For each of the tasks related to the vault connector, execute the corresponding commands from this directory.

* Verify and Export environment variables: `source  ../../pkg/policy-compiler/policy-compiler.env`

* Build the connector: `make build`

* Build and push the docker images: `make docker-build docker-push`

* Run the connectors locally: `make run`

* Terminate the connector: `make terminate`

* Clean the docker images: `make docker-rmi`

* Clean the connector build: `make clean`

* Run the connector test cases: `make test`
