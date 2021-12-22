ROOT_DIR := ../..
TAXONOMY_FOLDER = charts/fybrik/files/taxonomy
DATA_CATALOG_FILE_NAME = data_catalog.json
DOCKER_NAME = katalog-connector

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk

all: docker-build docker-push

.PHONY: generate
generate: $(TOOLBIN)/controller-gen
	$(TOOLBIN)/controller-gen --version
	$(TOOLBIN)/controller-gen object:headerFile=$(ROOT_DIR)/hack/boilerplate.go.txt,year=$(shell date +%Y) paths=./pkg/v1alpha1/...
	$(TOOLBIN)/controller-gen crd output:crd:artifacts:config=$(ROOT_DIR)/charts/fybrik-crd/templates/ paths=./pkg/v1alpha1/...

.PHONY: generate-data-catalog-json
generate-data-catalog-json: generate
	cd ../..; go run main.go taxonomy generate \
	-i charts/fybrik-crd/templates/katalog.fybrik.io_assets.yaml \
	-o $(TAXONOMY_FOLDER)/$(DATA_CATALOG_FILE_NAME) \
	-t "Common definitions for Data Catalog"

.PHONY: generate-server
generate-server: $(TOOLBIN)/openapi-generator-cli
	mkdir -p $(ROOT_DIR)/connectors/katalog/server
	$(TOOLBIN)/openapi-generator-cli generate -g go-gin-server \
		--additional-properties=packageName=openapiserver,serverPort=8081,sourceFolder=openapiserver \
		--global-property=apis,supportingFiles -o $(ROOT_DIR)/connectors/katalog \
		-i manifests/taxonomy.spec.yaml
	rm -f $(ROOT_DIR)/connectors/katalog/server/go.mod $(ROOT_DIR)/connectors/katalog/server/.openapi-generator-ignore
	rm -f $(ROOT_DIR)/connectors/katalog/server/Dockerfile $(ROOT_DIR)/connectors/katalog/server/go/README.md
	rm -r $(ROOT_DIR)/connectors/katalog/server/api $(ROOT_DIR)/connectors/katalog/server/.openapi-generator

.PHONY: clean-server
clean-server:
	mkdir -p $(ROOT_DIR)/connectors/katalog/server/go
	cp  $(ROOT_DIR)/connectors/katalog/server/go/model.go $(ROOT_DIR)/connectors/katalog/server/go
	rm -rf $(ROOT_DIR)/connectors/katalog/server/go/model.go

docker-build: generate source-build
	docker build . -t ${IMG}
	rm katalog

.PHONY: source-build
source-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o katalog main.go

.PHONY: run
run: generate
	go run main.go run

.PHONY: test
test:
	go test -v ./...
