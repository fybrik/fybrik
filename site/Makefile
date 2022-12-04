ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

IGNORE := $(shell bash -c "sed -n /=/p  ${ROOT_DIR}/hack/tools/requirements.sh | sed 's/=/:=/' | sed 's/^/export /' > makeenv")
include makeenv
OPENAPI_GENERATOR_IMG := openapitools/openapi-generator-cli:v$(OPENAPI_GENERATOR_VERSION)

ABS_ROOT_DIR := $(abspath ${ROOT_DIR})
OPENAPI_GENERATOR_IMG := openapitools/openapi-generator-cli:v$(OPENAPI_GENERATOR_VERSION)

.PHONY: generate
generate: clean generate-crd-docs generate-connectors-docs

.PHONY: generate-crd-docs
generate-crd-docs: $(TOOLBIN)/crdoc
	mkdir tmp_site
	cp $(ROOT_DIR)/charts/fybrik-crd/templates/* tmp_site
	cp $(ROOT_DIR)/charts/fybrik-crd/charts/asset-crd/templates/* tmp_site
	cp $(ROOT_DIR)/charts/fybrik/charts/storageaccount-crd/templates/* tmp_site
	-$(TOOLBIN)/crdoc -v
	$(TOOLBIN)/crdoc --template ./templates/crd/main.tmpl --resources tmp_site --output ./docs/reference/crds.md
	rm -r tmp_site


.PHONY: generate-connectors-docs
generate-connectors-docs:
	rm -rf ./docs/reference/connectors-*

	docker run --user $(shell id -u):$(shell stat -c %g "${ABS_ROOT_DIR}") --rm -v "${ABS_ROOT_DIR}":/workdir \
			${OPENAPI_GENERATOR_IMG} generate \
			-i /workdir/connectors/api/datacatalog.spec.yaml \
			-g markdown \
			-t /workdir/site/templates/openapigenerator \
			-o  /workdir/site/docs/reference/connectors-datacatalog


	docker run --user $(shell id -u):$(shell stat -c %g "${ABS_ROOT_DIR}") --rm -v "${ABS_ROOT_DIR}":/workdir \
			${OPENAPI_GENERATOR_IMG} generate \
			-i /workdir/connectors/api/policymanager.spec.yaml \
			-g markdown \
			-t /workdir/site/templates/openapigenerator \
			-o  /workdir/site/docs/reference/connectors-policymanager

clean:
	rm makeenv

.PHONY: run
run: generate
	@echo "\nServing on http://127.0.0.1:8000/"
	mkdocs serve --quiet
