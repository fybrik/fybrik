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
	go run $(ROOT_DIR)/main.go taxonomy validate $(ROOT_DIR)/charts/fybrik/files/taxonomy/fybrik_module.json

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
		-i codegen.spec.yaml
	rm -f $(ROOT_DIR)/pkg/connectors/openapiclient/go.mod $(ROOT_DIR)/pkg/connectors/openapiclient/go.sum $(ROOT_DIR)/pkg/connectors/openapiclient/.travis.yml
	rm -f $(ROOT_DIR)/pkg/connectors/openapiclient/README.md $(ROOT_DIR)/pkg/connectors/openapiclient/git_push.sh
	rm -r $(ROOT_DIR)/pkg/connectors/openapiclient/api 

.PHONY: generate-server
generate-server: clean-server $(TOOLBIN)/openapi-generator-cli charts/taxonomy.json
	mkdir -p $(ROOT_DIR)/connectors/opa
	$(TOOLBIN)/openapi-generator-cli generate -g go-gin-server \
		--additional-properties=packageName=openapiserver,serverPort=8081,sourceFolder=openapiserver \
		--global-property=apis,supportingFiles -o $(ROOT_DIR)/connectors/opa \
		-i codegen.spec.yaml
	rm -f $(ROOT_DIR)/connectors/opa/go.mod $(ROOT_DIR)/connectors/opa/.openapi-generator-ignore  
	rm -f $(ROOT_DIR)/connectors/opa/Dockerfile $(ROOT_DIR)/connectors/opa/go/README.md 
	rm -r $(ROOT_DIR)/connectors/opa/api $(ROOT_DIR)/connectors/opa/.openapi-generator

.PHONY: clean
clean:
	rm -f $(ROOT_DIR)/charts/fybrik/files/taxonomy/taxonomy.json
	rm -rf $(ROOT_DIR)/pkg/taxonomy/model/base
	mkdir -p $(ROOT_DIR)/pkg/connectors/tmp
	cp  $(ROOT_DIR)/pkg/connectors/openapiclient/model.go  $(ROOT_DIR)/pkg/connectors/tmp
	rm -rf $(ROOT_DIR)/pkg/connectors/openapiclient
	mkdir -p $(ROOT_DIR)/pkg/connectors/openapiclient
	cp  $(ROOT_DIR)/pkg/connectors/tmp/model.go $(ROOT_DIR)/pkg/connectors/openapiclient
	rm -rf $(ROOT_DIR)/pkg/connectors/tmp
	

.PHONY: clean-server
clean-server:
	mkdir -p $(ROOT_DIR)/connectors/tmp
	cp  $(ROOT_DIR)/connectors/opa/go/model.go  $(ROOT_DIR)/connectors/tmp
	rm -rf $(ROOT_DIR)/connectors/opa
	mkdir -p $(ROOT_DIR)/connectors/opa/go
	cp  $(ROOT_DIR)/connectors/tmp/model.go $(ROOT_DIR)/connectors/opa/go
	rm -rf $(ROOT_DIR)/connectors/tmp