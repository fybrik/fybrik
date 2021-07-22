ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

.DEFAULT_GOAL := all

all: clean compile-taxonomy-schema validate compile-codegen-cleanup codegen-openapi-rest-client

compile-taxonomy-schema:
	go run $(ROOT_DIR)/main.go taxonomy compile --base ./base/base.yaml --out ./objects/taxonomy.json \
		--codegen

validate:
	go run $(ROOT_DIR)/main.go taxonomy validate ./objects/policy_manager_request.json
	go run $(ROOT_DIR)/main.go taxonomy validate ./objects/policy_manager_response.json

compile-codegen-cleanup: $(TOOLBIN)/openapi-generator-cli
	mkdir -p $(ROOT_DIR)/pkg/taxonomy/model/base
	$(TOOLBIN)/openapi-generator-cli generate -i ./codegen.spec.yaml -g go -o $(ROOT_DIR)/pkg/taxonomy/model/base --global-property=models,modelDocs=false \
	--additional-properties=packageName=base \
	--enable-post-process-file \
	--type-mappings=object=interface{}
# rm -f objects/taxonomy.json

# Code generated due to the below command is MODIFIED to suit our purposes
# below code is the based for implementing the http client

codegen-openapi-rest-client: clean $(TOOLBIN)/openapi-generator-cli
	mkdir -p $(ROOT_DIR)/pkg/connectors/openapiclient
	$(TOOLBIN)/openapi-generator-cli generate -g go --additional-properties=packageName=openapiclient,isGoSubmodule=false \
	--global-property=apis,supportingFiles -o $(ROOT_DIR)/pkg/connectors/openapiclient \
	--import-mappings=PolicyManagerRequest=github.com/mesh-for-data/mesh-for-data/pkg/taxonomy/model/base \
	--git-user-id "mesh-for-data" \
    --git-repo-id "mesh-for-data/pkg/connectors/openapiclient" \
	-i codegen.spec.yaml

clean:
	rm -f objects/taxonomy.json
	rm -rf $(ROOT_DIR)/pkg/taxonomy/model/base
	rm -rf $(ROOT_DIR)/pkg/connectors/openapiclient