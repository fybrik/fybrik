# Policy Compiler Service
For each of the tasks related to the policy compiler service, including the policy manager connectors, execute the corresponding commands from this directory.

* Verify and Export environment variables: `source  policy-compiler.env`

* Build the service and connectors: `make build`

* Build the docker images: `make docker`

* Run the service and connectors locally: `make run`

* Terminate the service and connectors: `make terminate`

* Clean the docker images: `make docker-clean`

* Clean the policy-compiler service and connectors build: `make clean`

* Run the service and connectors test cases: `make test`
