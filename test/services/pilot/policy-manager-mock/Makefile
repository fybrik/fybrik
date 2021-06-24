ROOT_DIR := ../../../..
DOCKER_NAME = plcy-mgr-mock

include $(ROOT_DIR)/hack/make-rules/docker.mk

.PHONY: build
build:
	go build

docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o policy-manager-mock
	docker build . -t $(IMG)
	rm policy-manager-mock

.PHONY: run
run:
	nohup ./policy-manager-mock > output-policy-manager-mock.log &  2>&1 && sleep 10

.PHONY: terminate
terminate:

.PHONY: clean
clean:
	rm -f policy-manager-mock && rm -f output-policy-manager-mock.log

.PHONY: test
test:

-include ${ROOT_DIR}/.env
.EXPORT_ALL_VARIABLES:
