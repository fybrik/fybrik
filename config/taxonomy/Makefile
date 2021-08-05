ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

.DEFAULT_GOAL := all
all: clean $(ROOT_DIR)/charts/fybrik/files/taxonomy/taxonomy.json generate-base generate-client

charts/taxonomy.json:
	go run $(ROOT_DIR)/main.go taxonomy compile --base ./base/base.yaml --out $(ROOT_DIR)/charts/fybrik/files/taxonomy/taxonomy.json
	go run $(ROOT_DIR)/main.go taxonomy validate $(ROOT_DIR)/charts/fybrik/files/taxonomy/policy_manager_request.json
	go run $(ROOT_DIR)/main.go taxonomy validate $(ROOT_DIR)/charts/fybrik/files/taxonomy/policy_manager_response.json
	go run $(ROOT_DIR)/main.go taxonomy validate $(ROOT_DIR)/charts/fybrik/files/taxonomy/fybrik_application.json

.PHONY: generate-base
generate-base: $(TOOLBIN)/openapi-generator-cli charts/taxonomy.json
	mkdir -p $(ROOT_DIR)/pkg/taxonomy/model/base
	$(TOOLBIN)/openapi-generator-cli generate -i ./codegen.spec.yaml -g go -o $(ROOT_DIR)/pkg/taxonomy/model/base --global-property=models,modelDocs=false \
		--additional-properties=packageName=base \
		--enable-post-process-file \
		--type-mappings=object=interface{}

.PHONY: generate-client
generate-client: $(TOOLBIN)/openapi-generator-cli charts/taxonomy.json
	mkdir -p $(ROOT_DIR)/pkg/connectors/openapiclient
	$(TOOLBIN)/openapi-generator-cli generate -g go --additional-properties=packageName=openapiclient,isGoSubmodule=false \
		--global-property=apis,supportingFiles,apiDocs=false -o $(ROOT_DIR)/pkg/connectors/openapiclient \
		--type-mappings=PolicyManagerRequest=base.PolicyManagerRequest,PolicyManagerResponse=base.PolicyManagerResponse \
		--import-mappings=base.PolicyManagerRequest=github.com/mesh-for-data/mesh-for-data/pkg/taxonomy/model/base,base.PolicyManagerResponse=github.com/mesh-for-data/mesh-for-data/pkg/taxonomy/model/base \
		-i codegen.spec.yaml
	rm -f $(ROOT_DIR)/pkg/connectors/openapiclient/go.mod $(ROOT_DIR)/pkg/connectors/openapiclient/go.sum $(ROOT_DIR)/pkg/connectors/openapiclient/.travis.yml
	rm -f $(ROOT_DIR)/pkg/connectors/openapiclient/README.md $(ROOT_DIR)/pkg/connectors/openapiclient/git_push.sh
	rm -r $(ROOT_DIR)/pkg/connectors/openapiclient/api 

.PHONY: generate-server
generate-server: $(TOOLBIN)/openapi-generator-cli objects/taxonomy.json
	mkdir -p $(ROOT_DIR)/connectors/open_policy_agent
	$(TOOLBIN)/openapi-generator-cli generate -g go-gin-server \
		--additional-properties=packageName=openapiserver,serverPort=8081,sourceFolder=openapiserver \
		--global-property=apis,supportingFiles -o $(ROOT_DIR)/connectors/open_policy_agent \
		-i codegen.spec.yaml
	rm -f $(ROOT_DIR)/connectors/open_policy_agent/go.mod $(ROOT_DIR)/connectors/open_policy_agent/.openapi-generator-ignore  
	rm -f $(ROOT_DIR)/connectors/open_policy_agent/Dockerfile $(ROOT_DIR)/connectors/open_policy_agent/go/README.md 
	rm -r $(ROOT_DIR)/connectors/open_policy_agent/api $(ROOT_DIR)/connectors/open_policy_agent/.openapi-generator

.PHONY: clean
clean:
	rm -f $(ROOT_DIR)/charts/fybrik/files/taxonomy/taxonomy.json
	rm -rf $(ROOT_DIR)/pkg/taxonomy/model/base
	rm -rf $(ROOT_DIR)/pkg/connectors/openapiclient
	mkdir -p $(ROOT_DIR)/connectors/tmp
	cp  $(ROOT_DIR)/connectors/open_policy_agent/go/model.go  $(ROOT_DIR)/connectors/tmp
	rm -rf $(ROOT_DIR)/connectors/open_policy_agent
	mkdir -p $(ROOT_DIR)/connectors/open_policy_agent/go
	cp  $(ROOT_DIR)/connectors/tmp/model.go $(ROOT_DIR)/connectors/open_policy_agent/go
	rm -rf $(ROOT_DIR)/connectors/tmp