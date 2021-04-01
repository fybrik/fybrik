ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

#.PHONY: deploy
#deploy: $(TOOLBIN)/kubectl 
#	$(TOOLBIN)/kubectl create configmap taxonomy-configmap --from-file=module.structs.schema.json --from-file=module.values.schema.json --from-file=catalog.structs.schema.json --from-file=catalog.values.schema.json --from-file=policymanager.structs.schema.json --from-file=policymanager.values.schema.json

#.PHONY: undeploy
#undeploy: $(TOOLBIN)/kubectl 
#	$(TOOLBIN)/kubectl delete configmap taxonomy-configmap

.PHONY: test
test:
	go test ./... -coverprofile cover.out
