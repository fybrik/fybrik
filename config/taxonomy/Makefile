ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

.DEFAULT_GOAL := all
all: clean objects/taxonomy.json generate-base generate-client

objects/taxonomy.json:
	go run $(ROOT_DIR)/main.go taxonomy compile --base ./base/base.yaml --out ./objects/taxonomy.json
	go run $(ROOT_DIR)/main.go taxonomy validate ./objects/policy_manager_request.json
	go run $(ROOT_DIR)/main.go taxonomy validate ./objects/policy_manager_response.json

.PHONY: generate-base
generate-base: $(TOOLBIN)/openapi-generator-cli objects/taxonomy.json
	mkdir -p $(ROOT_DIR)/pkg/taxonomy/model/base
	$(TOOLBIN)/openapi-generator-cli generate -i ./codegen.spec.yaml -g go -o $(ROOT_DIR)/pkg/taxonomy/model/base --global-property=models,modelDocs=false \
		--additional-properties=packageName=base \
		--enable-post-process-file \
		--type-mappings=object=interface{}

.PHONY: generate-client
generate-client: $(TOOLBIN)/openapi-generator-cli objects/taxonomy.json
	mkdir -p $(ROOT_DIR)/pkg/connectors/openapiclient
	$(TOOLBIN)/openapi-generator-cli generate -g go --additional-properties=packageName=openapiclient,isGoSubmodule=false \
		--global-property=apis,supportingFiles,apiDocs=false -o $(ROOT_DIR)/pkg/connectors/openapiclient \
		--type-mappings=PolicyManagerRequest=base.PolicyManagerRequest,PolicyManagerResponse=base.PolicyManagerResponse \
		--import-mappings=base.PolicyManagerRequest=github.com/mesh-for-data/mesh-for-data/pkg/taxonomy/model/base,base.PolicyManagerResponse=github.com/mesh-for-data/mesh-for-data/pkg/taxonomy/model/base \
		-i codegen.spec.yaml
	rm -f $(ROOT_DIR)/pkg/connectors/openapiclient/go.mod $(ROOT_DIR)/pkg/connectors/openapiclient/go.sum $(ROOT_DIR)/pkg/connectors/openapiclient/.travis.yml
	rm -f $(ROOT_DIR)/pkg/connectors/openapiclient/README.md $(ROOT_DIR)/pkg/connectors/openapiclient/git_push.sh
	rm -r $(ROOT_DIR)/pkg/connectors/openapiclient/api 

.PHONY: clean
clean:
	rm -f objects/taxonomy.json
	rm -rf $(ROOT_DIR)/pkg/taxonomy/model/base
	rm -rf $(ROOT_DIR)/pkg/connectors/openapiclient