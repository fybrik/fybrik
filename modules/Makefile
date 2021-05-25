ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env
NAME := m4d-template

.PHONY: all
all:
	@echo
	@echo "### to create m4d-newchart run ###"
	@echo "cp -ar m4d-template m4d-newchart"
	@echo "egrep -lRZ 'm4d-template' . | xargs -0 -l sed -i -e 's/m4d-template/m4d-newchart/g'"

.PHONY: helm
helm:
	$(MAKE) -C m4d-template helm-all
	$(MAKE) -C m4d-implicit-copy-batch helm-all
	$(MAKE) -C m4d-implicit-copy-stream helm-all


.PHONY: helm-chart-push
helm-chart-push:
	$(MAKE) -C m4d-template helm-chart-push
	$(MAKE) -C m4d-implicit-copy-batch helm-chart-push
	$(MAKE) -C m4d-implicit-copy-stream helm-chart-push

.PHONY: helm-uninstall
helm-uninstall:
	$(MAKE) -C m4d-template helm-uninstall
	$(MAKE) -C m4d-implicit-copy-batch helm-uninstall
	$(MAKE) -C m4d-implicit-copy-stream helm-uninstall

# Tests if implicit copy definitions can be installed
.PHONY: test
test:
	$(TOOLBIN)/kubectl apply -f implicit-copy-batch-module.yaml
	$(TOOLBIN)/kubectl apply -f implicit-copy-stream-module.yaml
	$(TOOLBIN)/kubectl delete -f implicit-copy-batch-module.yaml
	$(TOOLBIN)/kubectl delete -f implicit-copy-stream-module.yaml
