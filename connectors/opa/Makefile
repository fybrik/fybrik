ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
DOCKER_NAME=opa-connector
DOCKER_CONTEXT=$(ROOT_DIR)
DOCKER_FILE=$(DOCKER_CONTEXT)/connectors/opa/Dockerfile
include $(ROOT_DIR)/.mk/docker.mk

OPA_PATH=${ABSTOOLBIN}/opa

#should work with single file or folder
OPA_DEFAULT_POLICIES=${ROOT_DIR}/third_party/opa/opa-policy.rego

.PHONY: opaserver
opaserver:
	{ $(OPA_PATH) run --server $(OPA_DEFAULT_POLICIES) > output-opa.log 2>&1 & echo $$! >> opa_pids.txt; }

.PHONY: opaserver-terminate
opaserver-terminate:
	-kill -9 $$( cat opa_pids.txt ) && rm -f opa_pids.txt

.PHONY: opaserver-clean
opaserver-clean:
	rm -f output-opa.log

.PHONY: build
build:
	go build -o opa_connector

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean: opaserver-clean
	rm -f opa_connector && rm -f output-opaconnector.log

.PHONY: run
run: opaserver
	nohup ./opa_connector > output-opaconnector.log &  2>&1 && echo "$$!" >> pids.txt && sleep 5

.PHONY: terminate
terminate: opaserver-terminate
	kill -9 $$( cat pids.txt ); rm -f pids.txt

