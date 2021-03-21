ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl 
	$(TOOLBIN)/kubectl create configmap taxonomy-configmap --from-file=module_schema.json --from-file=data_catalog_schema.json --from-file=policy_manager_schema.json

.PHONY: undeploy
undeploy: $(TOOLBIN)/kubectl 
	$(TOOLBIN)/kubectl delete configmap taxonomy-configmap

.PHONY: test
test:
	go test ./... -coverprofile cover.out
