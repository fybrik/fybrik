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


.PHONY: helm-chart-push
helm-chart-push:
	$(MAKE) -C fybrik-template helm-chart-push

.PHONY: helm-uninstall
helm-uninstall:
	$(MAKE) -C fybrik-template helm-uninstall
