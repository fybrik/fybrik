ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk


CONNECTORS_PROTO := $(ROOT_DIR)/pkg/connectors/protos

.PHONY: generate
generate: $(TOOLBIN)/crdoc $(TOOLBIN)/protoc $(TOOLBIN)/protoc-gen-doc
	PATH=$(TOOLBIN) crdoc --template ./templates/crd/main.tmpl --resources $(ROOT_DIR)/charts/fybrik-crd/templates --output ./docs/reference/crds.md
	PATH=$(TOOLBIN) protoc -I=$(CONNECTORS_PROTO) --doc_out=./docs/reference/ --doc_opt=./templates/proto/protoc.tmpl,connectors.md $(sort $(wildcard ./$(CONNECTORS_PROTO)/*.proto))
	go mod tidy

generate-openapi-doc:
	$(TOOLBIN)/openapi-generator-cli generate -i $(ROOT_DIR)/connectors/api/datacatalog.spec.yaml -g markdown -o  $(ROOT_DIR)/site/docs/reference/openapigenerator-doc-datacatalog
	$(TOOLBIN)/openapi-generator-cli generate -i $(ROOT_DIR)/connectors/api/policymanager.spec.yaml -g markdown -o $(ROOT_DIR)/site/docs/reference/openapigenerator-doc-policymanager

.PHONY: run
run: generate
	python -m mkdocs serve
