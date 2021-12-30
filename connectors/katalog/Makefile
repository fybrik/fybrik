ROOT_DIR := ../..
TAXONOMY_FOLDER = charts/fybrik/files/taxonomy
DATA_CATALOG_FILE_NAME = data_catalog.json
DOCKER_NAME = katalog-connector

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk

all: docker-build docker-push

docker-build: generate source-build
	docker build . -t ${IMG}
	rm -rf bin

.PHONY: source-build
source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o bin/katalog main.go

.PHONY: run
run: generate
	go run main.go run

.PHONY: test
test:
	go test -v ./...
