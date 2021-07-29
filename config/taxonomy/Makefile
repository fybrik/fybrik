ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

.DEFAULT_GOAL := all

all: clean compile-taxonomy-schema validate

compile-taxonomy-schema:
	go run $(ROOT_DIR)/main.go taxonomy compile --base ./base/base.yaml --out ./objects/taxonomy.json \
		--codegen

validate:
	go run $(ROOT_DIR)/main.go taxonomy validate ./objects/policy_manager_request.json
	go run $(ROOT_DIR)/main.go taxonomy validate ./objects/policy_manager_response.json

compile-codegen-cleanup: clean $(TOOLBIN)/openapi-generator-cli compile-taxonomy-schema
	mkdir -p $(ROOT_DIR)/pkg/taxonomy/model/base
	$(TOOLBIN)/openapi-generator-cli generate -i ./codegen.spec.yaml -g go -o $(ROOT_DIR)/pkg/taxonomy/model/base --global-property=models,modelDocs=false \
	--additional-properties=packageName=base \
	--enable-post-process-file \
	--type-mappings=object=interface{}
	rm -f objects/taxonomy.json

clean:
	rm -f objects/taxonomy.json
	rm -rf $(ROOT_DIR)/pkg/taxonomy/model/base
