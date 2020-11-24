ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
DOCKER_NAME=opa-connector
DOCKER_CONTEXT=$(ROOT_DIR)
DOCKER_FILE=$(DOCKER_CONTEXT)/connectors/opa/Dockerfile
include $(ROOT_DIR)/hack/make-rules/docker.mk

OPA_PATH=${ABSTOOLBIN}/opa
OPA_DEFAULT_POLICY_FILE=${ROOT_DIR}/third_party/opa/opa-policy.rego

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o opa-connector opa_connector.go evaluate_opa.go
	docker build . -t $(IMG) -f Dockerfile
	rm opa-connector

.PHONY: opaserver
opaserver:
	{ $(OPA_PATH) run --server $(OPA_DEFAULT_POLICY_FILE) > output-opa.log 2>&1 & echo $$! >> opa_pids.txt; }

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

.PHONY: clean
clean: opaserver-clean
	rm -f opa_connector && rm -f output-opaconnector.log

.PHONY: run
run: opaserver
	nohup ./opa_connector > output-opaconnector.log &  2>&1 && echo "$$!" >> pids.txt && sleep 5

.PHONY: terminate
terminate: opaserver-terminate
	kill -9 $$( cat pids.txt ); rm -f pids.txt

