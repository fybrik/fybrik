ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl apply --validate=false -f https://raw.githubusercontent.com/datashim-io/datashim/master/release-tools/manifests/dlf.yaml

.PHONY: deploy-wait
deploy-wait: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl wait --for=condition=ready pod -n dlf --all --timeout=120s

.PHONY: undeploy
undeploy: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl delete -f https://raw.githubusercontent.com/datashim-io/datashim/master/release-tools/manifests/dlf.yaml
