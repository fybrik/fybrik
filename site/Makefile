ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk


CONNECTORS_PROTO := $(ROOT_DIR)/pkg/connectors/protos

.PHONY: generate
generate: $(TOOLBIN)/crdoc $(TOOLBIN)/protoc $(TOOLBIN)/protoc-gen-doc
	PATH=$(TOOLBIN) crdoc --template ./templates/crd/main.tmpl --resources $(ROOT_DIR)/charts/fybrik-crd/templates --output ./docs/reference/crds.md
	PATH=$(TOOLBIN) protoc -I=$(CONNECTORS_PROTO) --doc_out=./docs/reference/ --doc_opt=./templates/proto/protoc.tmpl,connectors.md $(wildcard ./$(CONNECTORS_PROTO)/*.proto)
	go mod tidy

.PHONY: run
run: generate
	mkdocs serve
