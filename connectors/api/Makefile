ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk
IGNORE := $(shell bash -c "sed -n /=/p  ${ROOT_DIR}/hack/tools/requirements.env | sed 's/=/:=/' | sed 's/^/export /' > makeenv")
include makeenv

ABS_ROOT_DIR := $(abspath ${ROOT_DIR})
OPENAPI_GENERATOR_IMG := openapitools/openapi-generator-cli:v${OPENAPI_GENERATOR_VERSION}

.PHONY: generate-client-datacatalog
generate-client-datacatalog: clean
	mkdir -p $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient

	docker run --user $(shell id -u):$(shell stat -c %g "${ABS_ROOT_DIR}") --rm -v "${ABS_ROOT_DIR}":/workdir \
		${OPENAPI_GENERATOR_IMG} generate \
		-i /workdir/connectors/api/datacatalog.spec.yaml \
		-g go \
		--additional-properties=packageName=openapiclient,isGoSubmodule=false \
		--global-property=apis,supportingFiles,apiDocs=false \
		-o  /workdir/pkg/connectors/datacatalog/openapiclient

	rm -f $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/go.mod $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/go.sum $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/.travis.yml
	rm -f $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/README.md $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/git_push.sh
	rm -f $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/.gitignore $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/.openapi-generator-ignore
	rm -r $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/api

	cd $(ROOT_DIR) && make verify

.PHONY: generate-client-policymanager
generate-client-policymanager: clean
	mkdir -p $(ROOT_DIR)/pkg/connectors/policymanager/openapiclient

	docker run --user $(shell id -u):$(shell stat -c %g "${ABS_ROOT_DIR}") --rm -v "${ABS_ROOT_DIR}":/workdir \
		${OPENAPI_GENERATOR_IMG} generate \
		-i /workdir/connectors/api/policymanager.spec.yaml \
		-g go \
		--additional-properties=packageName=openapiclient,isGoSubmodule=false \
		--global-property=apis,supportingFiles,apiDocs=false \
		-o  /workdir/pkg/connectors/policymanager/openapiclient

	rm -f $(ROOT_DIR)/pkg/connectors/policymanager/openapiclient/go.mod $(ROOT_DIR)/pkg/connectors/policymanager/openapiclient/go.sum $(ROOT_DIR)/pkg/connectors/policymanager/openapiclient/.travis.yml
	rm -f $(ROOT_DIR)/pkg/connectors/policymanager/openapiclient/README.md $(ROOT_DIR)/pkg/connectors/policymanager/openapiclient/git_push.sh
	rm -f $(ROOT_DIR)/pkg/connectors/policymanager/openapiclient/.gitignore $(ROOT_DIR)/pkg/connectors/policymanager/openapiclient/.openapi-generator-ignore
	rm -r $(ROOT_DIR)/pkg/connectors/policymanager/openapiclient/api

	cd $(ROOT_DIR) && make verify

.PHONY: generate-client-storagemanager
generate-client-storagemanager: clean
	mkdir -p $(ROOT_DIR)/pkg/connectors/storagemanager/openapiclient

	docker run --user $(shell id -u):$(shell stat -c %g "${ABS_ROOT_DIR}") --rm -v "${ABS_ROOT_DIR}":/workdir \
		${OPENAPI_GENERATOR_IMG} generate \
		-i /workdir/connectors/api/storagemanager.spec.yaml \
		-g go \
		--additional-properties=packageName=openapiclient,isGoSubmodule=false \
		--global-property=apis,supportingFiles,apiDocs=false \
		-o  /workdir/pkg/connectors/storagemanager/openapiclient

	rm -f $(ROOT_DIR)/pkg/connectors/storagemanager/openapiclient/go.mod $(ROOT_DIR)/pkg/connectors/storagemanager/openapiclient/go.sum $(ROOT_DIR)/pkg/connectors/storagemanager/openapiclient/.travis.yml
	rm -f $(ROOT_DIR)/pkg/connectors/storagemanager/openapiclient/README.md $(ROOT_DIR)/pkg/connectors/storagemanager/openapiclient/git_push.sh
	rm -f $(ROOT_DIR)/pkg/connectors/storagemanager/openapiclient/.gitignore $(ROOT_DIR)/pkg/connectors/storagemanager/openapiclient/.openapi-generator-ignore
	rm -r $(ROOT_DIR)/pkg/connectors/storagemanager/openapiclient/api

	cd $(ROOT_DIR) && make verify

clean:
	rm makeenv
