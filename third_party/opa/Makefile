ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl
	./deploy.sh deploy

.PHONY: undeploy
undeploy: $(TOOLBIN)/kubectl
	./deploy.sh undeploy

.PHONY: loadpolicy
loadpolicy: $(TOOLBIN)/kubectl
	./deploy.sh loadpolicy $(ARGS)

.PHONY: loaddata
loaddata: $(TOOLBIN)/kubectl
	./deploy.sh loaddata $(ARGS)

.PHONY: unloadpolicy
unloadpolicy: $(TOOLBIN)/kubectl
	./deploy.sh unloadpolicy $(ARGS)

.PHONY: unloaddata
unloaddata: $(TOOLBIN)/kubectl
	./deploy.sh unloaddata $(ARGS)

.PHONY: charts
charts:
	rm -rf $(ROOT_DIR)/charts/m4d/files/opa-server/
	mkdir -p $(ROOT_DIR)/charts/m4d/files/opa-server/policy-lib
	cp data-and-policies/meshfordata-policy-lib/** $(ROOT_DIR)/charts/m4d/files/opa-server/policy-lib
	rm $(ROOT_DIR)/charts/m4d/files/opa-server/policy-lib/taxonomies_*
