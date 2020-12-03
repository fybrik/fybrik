ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/.mk/cluster.mk

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl $(TOOLBIN)/istioctl
	./deploy.sh

.PHONY: deploy-wait
deploy-wait: 
	./deploy.sh wait

.PHONY: undeploy
undeploy: $(TOOLBIN)/kubectl $(TOOLBIN)/istioctl
	./deploy.sh cleanup || true

.PHONY: status
status: $(TOOLBIN)/kubectl $(TOOLBIN)/istioctl
	./deploy.sh status
