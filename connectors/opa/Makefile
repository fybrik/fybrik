ROOT_DIR := ../..
DOCKER_NAME = opa-connector

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk
include $(ROOT_DIR)/hack/make-rules/version.mk
include ${ROOT_DIR}/hack/tools/requirements.env

KUBE_NAMESPACE ?= fybrik-system
POLICY_FILES ?=

.PHONY: all
all: docker-build docker-push

.PHONY: docker-build
docker-build: source-build
	docker build . -t ${IMG} -f Dockerfile --build-arg tag=${BASE_IMAGE_TAG}
	rm opa-connector

.PHONY: source-build
source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build $(LDFLAGS) -o opa-connector .

.PHONY: run
run:
	go run .

.PHONY: test
test:
	go test $(TEST_OPTIONS) ./...

.PHONY: opaserver
opaserver: $(TOOLBIN)/opa
	${TOOLBIN}/opa run --server ${POLICY_FILES}

.PHONY: deploy
deploy:
	helm install fybrik-openpolicyagent $(ROOT_DIR)/charts/fybrik \
		--set manager.enabled=false --set coordinator.enabled=false --set worker.enabled=false \
		--set opaServer.enabled=true --set opaConnector.enabled=true \
		--set opaConnector.image=${DOCKER_HOSTNAME}/${DOCKER_NAMESPACE}/$(DOCKER_NAME):${DOCKER_TAGNAME} \
		-n $(KUBE_NAMESPACE) \
		--wait

.PHONY: undeploy
undeploy:
	helm uninstall fybrik-openpolicyagent -n $(KUBE_NAMESPACE)

-include ${ROOT_DIR}/.env
.EXPORT_ALL_VARIABLES:
