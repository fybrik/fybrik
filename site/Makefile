ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

ABS_ROOT_DIR := $(abspath ${ROOT_DIR})
OPENAPI_GENERATOR_IMG := openapitools/openapi-generator-cli:v5.3.1
CONNECTORS_PROTO := $(ROOT_DIR)/pkg/connectors/protos

.PHONY: generate
generate: generate-crd-docs generate-connectors-docs

.PHONY: generate-crd-docs
generate-crd-docs: $(TOOLBIN)/crdoc
	PATH=$(TOOLBIN) crdoc --template ./templates/crd/main.tmpl --resources $(ROOT_DIR)/charts/fybrik-crd/templates --output ./docs/reference/crds.md
	go mod tidy

.PHONY: generate-connectors-docs
generate-connectors-docs:
	rm -rf ./docs/reference/connectors-*

	docker run --user $(shell id -u):$(shell id -g) --rm -v "${ABS_ROOT_DIR}":/workdir \
			${OPENAPI_GENERATOR_IMG} generate \
			-i /workdir/connectors/api/datacatalog.spec.yaml \
			-g markdown \
			-t /workdir/site/templates/openapigenerator \
			-o  /workdir/site/docs/reference/connectors-datacatalog


	docker run --user $(shell id -u):$(shell id -g) --rm -v "${ABS_ROOT_DIR}":/workdir \
			${OPENAPI_GENERATOR_IMG} generate \
			-i /workdir/connectors/api/policymanager.spec.yaml \
			-g markdown \
			-t /workdir/site/templates/openapigenerator \
			-o  /workdir/site/docs/reference/connectors-policymanager


.PHONY: run
run: generate
	python -m mkdocs serve --quiet
