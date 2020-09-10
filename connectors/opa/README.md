# OPA Connector
For each of the tasks related to the opa connector, execute the corresponding commands from this directory.

* Verify and Export environment variables: `source  ../../pkg/policy-compiler/policy-compiler.env`

* Build the connector: `make build`

* Build the docker images: `make docker-all`

* Run the connectors locally: `make run`

* Terminate the connector: `make terminate`

* Clean the docker images: `make docker-rmi`

* Clean the connector build: `make clean`

* Run the connector test cases: `make test`
