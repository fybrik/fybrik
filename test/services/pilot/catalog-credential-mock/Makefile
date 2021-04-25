ROOT_DIR := ../../../..
DOCKER_NAME = ctlg-cred-mock

include $(ROOT_DIR)/hack/make-rules/docker.mk

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o catalog-credential-mock
	docker build . -t $(IMG)
	rm catalog-credential-mock

.PHONY: build
build:
	go build

.PHONY: run
run:
	nohup ./catalog-credential-mock > output-catalog-credential-mock.log &  2>&1 && sleep 10

.PHONY: terminate
terminate:

.PHONY: clean
clean:
	rm -f catalog-credential-mock && rm -f output-catalog-credential-mock.log

.PHONY: test
test:

-include ${ROOT_DIR}/.env
.EXPORT_ALL_VARIABLES:
