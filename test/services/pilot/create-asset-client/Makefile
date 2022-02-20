ROOT_DIR := ../../../..
DOCKER_NAME = create-asset-client

include $(ROOT_DIR)/hack/make-rules/docker.mk

include $(ROOT_DIR)/Makefile.env

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o create-asset-client
	docker build . -t $(IMG)
	rm create-asset-client

.PHONY: build
build:
	go build

.PHONY: run
run:
	nohup ./create-asset-client > output-create-asset-client.log &  2>&1 && sleep 10

.PHONY: terminate
terminate:

.PHONY: clean
clean:
	rm -f create-asset-client && rm -f output-create-asset-client.log

.PHONY: test
test:

-include ${ROOT_DIR}/.env
.EXPORT_ALL_VARIABLES:


.PHONY: deploy
deploy: $(TOOLBIN)/kustomize
	./deploy.sh

.PHONY: undeploy
undeploy:
	./deploy.sh undeploy
