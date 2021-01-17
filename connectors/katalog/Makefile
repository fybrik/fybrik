ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk
include $(ROOT_DIR)/hack/make-rules/tools.mk
DOCKER_NAME=katalog-connector

all: build docker-build docker-push deploy

.PHONY: generate
generate: $(TOOLBIN)/oapi-codegen $(TOOLBIN)/crdoc $(TOOLBIN)/openapi2crd
	go generate ./...

.PHONY: build
build: generate
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o bin/katalog main.go

.PHONY: clean
clean: docker-rmi
	rm bin/*

.PHONY: deploy
deploy: install/kustomization.yaml
	cd install && kustomize edit set image ${DOCKER_NAME}=${IMG}
	kustomize build install | kubectl apply -f -

.PHONY: undeploy
undeploy:
	kustomize build install | kubectl delete -f -

.PHONY: run
run: generate
	go run main.go run

install/kustomization.yaml: install/kustomization.yaml.in
	cp $< $@
