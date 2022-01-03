ROOT_DIR := ../..
DOCKER_NAME = opa-connector

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk

KUBE_NAMESPACE ?= fybrik-system

NEED_TEST_COVERAGE_STATISTICS=1
NEED_TEST_COVERAGE_STATISTICS_IN_HTML=0

CATALOG_PROVIDER_NAME ?= "katalog"
CATALOG_CONNECTOR_URL ?= "http://katalog-connector:50090"

docker-all: docker-build docker-push 

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o opa-connector main.go
	docker build . -t $(IMG)
	rm opa-connector

.PHONY: source-build
source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o opa-connector main.go

.PHONY: deploy
deploy:
	helm install fybrik-openpolicyagent $(ROOT_DIR)/charts/fybrik \
		--set coordinator.enabled=true --set worker.enabled=true --set manager.enabled=true \
		--set opaServer.enabled=true --set openPolicyAgentConnector.enabled=true \
		--set openPolicyAgentConnector.image=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/$(DOCKER_NAME):${DOCKER_TAGNAME} \
		--set coordinator.catalog=${CATALOG_PROVIDER_NAME} \
		--set coordinator.catalogConnectorURL=${CATALOG_CONNECTOR_URL} \
		--set coordinator.policyManagerConnectorURL=http://$(DOCKER_NAME):80 \
		-n $(KUBE_NAMESPACE)

.PHONY: undeploy
undeploy:
	helm uninstall fybrik-openpolicyagent -n $(KUBE_NAMESPACE)


.PHONY: build
build:
	go build -o opa-connector

.PHONY: test
test:
ifeq ($(NEED_TEST_COVERAGE_STATISTICS),1)
	go test -v ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
    ifeq ($(NEED_TEST_COVERAGE_STATISTICS_IN_HTML),1)
		go tool cover -html=coverage.out
    endif
	rm coverage.out
else
	go test -v ./...
endif

.PHONY: clean
clean: opaserver-clean
	rm -f opa-connector && rm -f output-openpolicyagentconnector.log

.PHONY: run
run: opaserver
	nohup ./opa-connector > output-openpolicyagentconnector.log &  2>&1 && echo "$$!" >> pids.txt && sleep 5

.PHONY: terminate
terminate: opaserver-terminate
	kill -9 $$( cat pids.txt ); rm -f pids.txt

# OPA server targets

OPA_PATH=${ABSTOOLBIN}/opa
OPA_DEFAULT_POLICY_FILE=${ROOT_DIR}/third_party/opa/data-and-policies/user_policies.rego
OPA_DEFAULT_POLICY_FOLDER=${ROOT_DIR}/third_party/opa/data-and-policies/data_policies

.PHONY: opaserver
opaserver:
	{ $(OPA_PATH) run --server $(OPA_DEFAULT_POLICY_FILE) $(OPA_DEFAULT_POLICY_FOLDER) > output-opa.log 2>&1 & echo $$! >> opa_pids.txt; }

.PHONY: opaserver-terminate
opaserver-terminate:
	-kill -9 $$( cat opa_pids.txt ) && rm -f opa_pids.txt

.PHONY: opaserver-clean
opaserver-clean:
	rm -f output-opa.log


-include ${ROOT_DIR}/.env
.EXPORT_ALL_VARIABLES:
