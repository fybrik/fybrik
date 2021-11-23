ROOT_DIR := ../..
TAXONOMY_FOLDER = charts/fybrik/files/taxonomy
DATA_CATALOG_FILE_NAME = data_catalog.json
DOCKER_NAME = katalog-connector

include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk

all: docker-build docker-push

#.PHONY: generate
#generate: $(TOOLBIN)/oapi-codegen $(TOOLBIN)/openapi2crd
#	$(TOOLBIN)/oapi-codegen -generate "types,skip-prune" -package taxonomy -o pkg/taxonomy/taxonomy.gen.go ./manifests/taxonomy.yaml
#	$(TOOLBIN)/oapi-codegen -generate "types,skip-prune" -import-mapping=taxonomy.yaml:fybrik.io/fybrik/connectors/katalog/pkg/taxonomy -package api -o pkg/api/spec.gen.go ./manifests/spec.yaml
#	$(TOOLBIN)/openapi2crd manifests/spec.yaml -g katalog.fybrik.io/v1alpha1/Asset -o $(ROOT_DIR)/charts/fybrik-crd/templates/katalog.fybrik.io_assets.yaml

.PHONY: generate
generate: $(TOOLBIN)/controller-gen
	$(TOOLBIN)/controller-gen --version
	$(TOOLBIN)/controller-gen object:headerFile=$(ROOT_DIR)/hack/boilerplate.go.txt,year=$(shell date +%Y) paths=./pkg/api/...
	$(TOOLBIN)/controller-gen crd output:crd:artifacts:config=$(ROOT_DIR)/charts/fybrik-crd/templates/ paths=./pkg/api/...

.PHONY: generate-data-catalog-json
generate-data-catalog-json: generate
	cd ../..; go run main.go taxonomy generate \
	-i charts/fybrik-crd/templates/katalog.fybrik.io_assets.yaml \
	-o $(TAXONOMY_FOLDER)/$(DATA_CATALOG_FILE_NAME) \
	-t "Common definitions for Data Catalog"

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
