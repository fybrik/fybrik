ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
DOCKER_NAME=opa-connector
DOCKER_CONTEXT=$(ROOT_DIR)
DOCKER_FILE=$(DOCKER_CONTEXT)/connectors/opa/Dockerfile
include $(ROOT_DIR)/hack/make-rules/docker.mk

OPA_PATH=${ABSTOOLBIN}/opa
OPA_DEFAULT_POLICY_FILE=${ROOT_DIR}/third_party/opa/data-and-policies/user_policies.rego
OPA_DEFAULT_POLICY_FOLDER=${ROOT_DIR}/third_party/opa/data-and-policies/data_policies
NEED_TEST_COVERAGE_STATISTICS=1
NEED_TEST_COVERAGE_STATISTICS_IN_HTML=0

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o opa-connector main.go
	docker build . -t $(IMG) -f Dockerfile
	rm opa-connector

.PHONY: opaserver
opaserver:
	{ $(OPA_PATH) run --server $(OPA_DEFAULT_POLICY_FILE) $(OPA_DEFAULT_POLICY_FOLDER) > output-opa.log 2>&1 & echo $$! >> opa_pids.txt; }

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

IMAGE := opa-connector

kustomization.yaml: kustomization.yaml.in
	cp $< $@

.PHONY: deploy
deploy: kustomization.yaml
	kustomize edit set image $(IMAGE)=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/$(IMAGE):${DOCKER_TAGNAME}
	kustomize build . | kubectl apply -f -

.PHONY: undeploy
undeploy: kustomization.yaml
	kustomize edit set image $(IMAGE)=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/$(IMAGE):${DOCKER_TAGNAME}
	kustomize build . | kubectl delete -f -

.PHONY: charts
charts: kustomization.yaml
	cp opa_connector.yaml ${ROOT_DIR}/charts/m4d/templates/.
