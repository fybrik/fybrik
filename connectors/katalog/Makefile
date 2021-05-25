ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk
DOCKER_NAME=katalog-connector

all: docker-build docker-push

.PHONY: generate
generate: $(TOOLBIN)/oapi-codegen $(TOOLBIN)/crdoc $(TOOLBIN)/openapi2crd
	$(TOOLBIN)/oapi-codegen -generate "types,skip-prune" -package taxonomy -o pkg/taxonomy/taxonomy.gen.go ./manifests/taxonomy.yaml
	$(TOOLBIN)/oapi-codegen -generate "types,skip-prune" -import-mapping=taxonomy.yaml:github.com/ibm/the-mesh-for-data/connectors/katalog/pkg/taxonomy -package api -o pkg/api/spec.gen.go ./manifests/spec.yaml
	$(TOOLBIN)/openapi2crd manifests/spec.yaml -g katalog.m4d.ibm.com/v1alpha1/Asset -o $(ROOT_DIR)/charts/m4d-crd/templates/katalog.m4d.ibm.com_assets.yaml
	$(TOOLBIN)/crdoc --template ./docs/main.tmpl -c ./docs/toc.yaml -r $(ROOT_DIR)/charts/m4d-crd/templates/ -o ./docs/README.md

docker-build: generate
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o katalog main.go
	docker build . -t ${IMG} -f Dockerfile
	rm katalog

.PHONY: run
run: generate
	go run main.go run

.PHONY: test
test:
	go test -v ./...