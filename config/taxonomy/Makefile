ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

.DEFAULT_GOAL := all

all: clean compile-taxonomy-schema validate model-codegen

compile-taxonomy-schema:
	go run $(ROOT_DIR)/main.go taxonomy compile --base ./base/base.yaml --out ./objects/taxonomy.json \
		--codegen

validate:
	go run ../../main.go taxonomy validate ./objects/policy_manager_request.json
	go run ../../main.go taxonomy validate ./objects/policy_manager_response.json

model-codegen: $(TOOLBIN)/openapi-generator-cli
	mkdir -p ../../pkg/taxonomy/model
	$(TOOLBIN)/openapi-generator-cli generate -i ./codegen.spec.yaml -g go -o ../../pkg/taxonomy/model --global-property=models \
	--additional-properties=legacyDiscriminatorBehavior=false \
	--enable-post-process-file \
	--type-mappings=object=interface{}

clean:
	rm ./objects/taxonomy.json
	rm ../../pkg/taxonomy/model/model_*.go
	rm -rf ../../pkg/taxonomy/model/docs