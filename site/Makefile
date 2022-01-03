ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk


CONNECTORS_PROTO := $(ROOT_DIR)/pkg/connectors/protos

.PHONY: generate
generate: $(TOOLBIN)/crdoc $(TOOLBIN)/protoc $(TOOLBIN)/protoc-gen-doc
	PATH=$(TOOLBIN) crdoc --template ./templates/crd/main.tmpl --resources $(ROOT_DIR)/charts/fybrik-crd/templates --output ./docs/reference/crds.md
	PATH=$(TOOLBIN) protoc -I=$(CONNECTORS_PROTO) --doc_out=./docs/reference/ --doc_opt=./templates/proto/protoc.tmpl,connectors.md $(sort $(wildcard ./$(CONNECTORS_PROTO)/*.proto))
	go mod tidy

generate-openapi-doc:
	rm -rf $(ROOT_DIR)/site/docs/reference/openapigenerator-doc-datacatalog
	rm -rf $(ROOT_DIR)/site/docs/reference/openapigenerator-doc-policymanager
	rm -rf ${PWD}/out
	rm -rf ${PWD}/connectors
	rm -rf ${PWD}/local

	mkdir -p $(ROOT_DIR)/site/docs/reference/openapigenerator-doc-datacatalog
	mkdir -p $(ROOT_DIR)/site/docs/reference/openapigenerator-doc-policymanager
	mkdir -p  ${PWD}/local/connectors/api/
	mkdir -p ${PWD}/local
	mkdir -p  ${PWD}/local/templates/openapigenerator

	cp $(ROOT_DIR)/connectors/api/datacatalog.spec.yaml ${PWD}/local/connectors/api/
	cp $(ROOT_DIR)/connectors/api/policymanager.spec.yaml  ${PWD}/local/connectors/api/
	cp $(ROOT_DIR)/charts/fybrik/files/taxonomy/datacatalog.json  ${PWD}/local/connectors/api/datacatalog.json
	cp $(ROOT_DIR)/charts/fybrik/files/taxonomy/policymanager.json  ${PWD}/local/connectors/api/policymanager.json
	cp $(ROOT_DIR)/charts/fybrik/files/taxonomy/taxonomy.json  ${PWD}/local/connectors/api/taxonomy.json
	cp -r ${PWD}/templates/openapigenerator/.  ${PWD}/local/templates/openapigenerator

	sed -i 's/..\/..\/charts\/fybrik\/files\/taxonomy\/datacatalog.json\#\/definitions\/GetAssetRequest/datacatalog.json#\/definitions\/GetAssetRequest/g' ./local/connectors/api/datacatalog.spec.yaml
	sed -i 's/..\/..\/charts\/fybrik\/files\/taxonomy\/datacatalog.json#\/definitions\/GetAssetResponse/datacatalog.json#\/definitions\/GetAssetResponse/g' ./local/connectors/api/datacatalog.spec.yaml

	sed -i 's/..\/..\/charts\/fybrik\/files\/taxonomy\/policymanager.json\#\/definitions\/GetPolicyDecisionsRequest/policymanager.json#\/definitions\/GetPolicyDecisionsRequest/g' ./local/connectors/api/policymanager.spec.yaml
	sed -i 's/..\/..\/charts\/fybrik\/files\/taxonomy\/policymanager.json\#\/definitions\/GetPolicyDecisionsResponse/policymanager.json#\/definitions\/GetPolicyDecisionsResponse/g' ./local/connectors/api/policymanager.spec.yaml

	docker run --user $(shell id -u):$(shell id -g) --rm \
	-v ${PWD}/local:/local openapitools/openapi-generator-cli generate \
	-i /local/connectors/api/datacatalog.spec.yaml \
	-g markdown \
	-t /local/templates/openapigenerator \
	-o /local/out/datacatalog

	docker run --user $(shell id -u):$(shell id -g) --rm \
	-v ${PWD}/local:/local openapitools/openapi-generator-cli generate \
	-i /local/connectors/api/policymanager.spec.yaml \
	-g markdown \
	-t /local/templates/openapigenerator \
	-o /local/out/policymanager

	cp -r ${PWD}/local/out/datacatalog/. $(ROOT_DIR)/site/docs/reference/openapigenerator-doc-datacatalog
	cp -r ${PWD}/local/out/policymanager/. $(ROOT_DIR)/site/docs/reference/openapigenerator-doc-policymanager

	rm -rf ${PWD}/local


.PHONY: run
run: generate
	python -m mkdocs serve
