ROOT_DIR := ../..
DOCKER_NAME = opa-connector

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk

KUBE_NAMESPACE ?= fybrik-system

NEED_TEST_COVERAGE_STATISTICS=1
NEED_TEST_COVERAGE_STATISTICS_IN_HTML=0

docker-build: source-build
	docker build . -t $(IMG)
	rm opa-connector

.PHONY: deploy
deploy:
	helm install fybrik-opa $(ROOT_DIR)/charts/fybrik \
		--set coordinator.enabled=false --set worker.enabled=false --set manager.enabled=false \
		--set opaServer.enabled=true --set opaConnector.enabled=true \
		--set opaConnector.image=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/$(DOCKER_NAME):${DOCKER_TAGNAME} \
		--set coordinator.catalogConnectorURL=${CATALOG_CONNECTOR_URL}
		-n $(KUBE_NAMESPACE)

.PHONY: source-build
source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o opa-connector main.go

.PHONY: undeploy
undeploy:
	helm uninstall fybrik-opa -n $(KUBE_NAMESPACE)


.PHONY: build
build:
	go build -o opa_connector

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
	rm -f opa_connector && rm -f output-opaconnector.log

.PHONY: run
run: opaserver
	nohup ./opa_connector > output-opaconnector.log &  2>&1 && echo "$$!" >> pids.txt && sleep 5

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
