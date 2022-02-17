ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

# charts/taxonomy.json:
# 	go run $(ROOT_DIR)/main.go taxonomy compile --base ./base/base.yaml --out $(ROOT_DIR)/charts/fybrik/files/taxonomy/taxonomy.json
# go run $(ROOT_DIR)/main.go taxonomy validate $(ROOT_DIR)/charts/fybrik/files/taxonomy/policy_manager_request.json

.PHONY: generate-client-datacatalog
generate-client-datacatalog: $(TOOLBIN)/openapi-generator-cli
	mkdir -p $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient
	$(TOOLBIN)/openapi-generator-cli generate -g go --additional-properties=packageName=openapiclient,isGoSubmodule=false \
		--global-property=apis,supportingFiles,apiDocs=true -o $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient \
		-i datacatalog.spec.yaml
	rm -f $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/go.mod $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/go.sum $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/.travis.yml
	rm -f $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/README.md $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/git_push.sh
	rm -f $(ROOT_DIR)/pkg/connectors/datacatalog2/openapiclient/.gitignore $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/.openapi-generator-ignore $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/.openapi-generator/FILES
	rm -r $(ROOT_DIR)/pkg/connectors/datacatalog/openapiclient/api
