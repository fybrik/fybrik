ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env
NAME := fybrik-template

.PHONY: all
all:
	@echo
	@echo "### to create fybrik-newchart run ###"
	@echo "cp -ar fybrik-template fybrik-newchart"
	@echo "egrep -lRZ 'fybrik-template' . | xargs -0 -l sed -i -e 's/fybrik-template/fybrik-newchart/g'"

.PHONY: helm
helm:
	$(MAKE) -C fybrik-template helm-all
	$(MAKE) -C fybrik-implicit-copy-batch helm-all
	$(MAKE) -C fybrik-implicit-copy-stream helm-all


.PHONY: helm-chart-push
helm-chart-push:
	$(MAKE) -C fybrik-template helm-chart-push
	$(MAKE) -C fybrik-implicit-copy-batch helm-chart-push
	$(MAKE) -C fybrik-implicit-copy-stream helm-chart-push

.PHONY: helm-uninstall
helm-uninstall:
	$(MAKE) -C fybrik-template helm-uninstall
	$(MAKE) -C fybrik-implicit-copy-batch helm-uninstall
	$(MAKE) -C fybrik-implicit-copy-stream helm-uninstall

# Tests if implicit copy definitions can be installed
.PHONY: test
test:
	$(TOOLBIN)/kubectl apply -f implicit-copy-batch-module.yaml
	$(TOOLBIN)/kubectl apply -f implicit-copy-stream-module.yaml
	$(TOOLBIN)/kubectl delete -f implicit-copy-batch-module.yaml
	$(TOOLBIN)/kubectl delete -f implicit-copy-stream-module.yaml
