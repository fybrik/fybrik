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
loadpolicy: $(TOOLBIN)/kubectl deploy
	./deploy.sh loadpolicy $(ARGS)

.PHONY: loaddata
loaddata: $(TOOLBIN)/kubectl deploy
	./deploy.sh loaddata $(ARGS)

.PHONY: unloadpolicy
unloadpolicy: $(TOOLBIN)/kubectl
	./deploy.sh unloadpolicy $(ARGS)

.PHONY: unloaddata
unloaddata: $(TOOLBIN)/kubectl
	./deploy.sh unloaddata $(ARGS)

