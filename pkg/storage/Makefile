ROOT_DIR := ../..
DOCKER_NAME = storage-manager

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk
include $(ROOT_DIR)/hack/make-rules/version.mk

.PHONY: all
all: docker-build docker-push

# Overwrite docker-build from docker.mk
.PHONY: docker-build
docker-build: source-build
	docker build . -t ${IMG} -f Dockerfile --build-arg tag=${BASE_IMAGE_TAG}
	rm -rf bin

.PHONY: source-build
source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build $(LDFLAGS) -o bin/storagemanager .

.PHONY: run
run:
	go run main.go run

.PHONY: test
test:
	go test $(TEST_OPTIONS) ./...
