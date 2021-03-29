ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk


CONNECTORS_PROTO := $(ROOT_DIR)/pkg/connectors/protos

.PHONY: generate
generate: $(TOOLBIN)/crdoc $(TOOLBIN)/protoc $(TOOLBIN)/protoc-gen-doc
	PATH=$(TOOLBIN) crdoc --template ./templates/main.tmpl --resources $(ROOT_DIR)/charts/m4d-crd/templates --output ./docs/reference/crds.md
	PATH=$(TOOLBIN) protoc -I=$(CONNECTORS_PROTO) --doc_out=./docs/reference/ --doc_opt=./templates/protoc.tmpl,connectors.md $(wildcard ./$(CONNECTORS_PROTO)/*.proto)

.PHONY: run
run: generate
	mkdocs serve
