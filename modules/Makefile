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
	$(MAKE) -C m4d-db2wh helm-all
	$(MAKE) -C m4d-kafka helm-all
	$(MAKE) -C m4d-s3-to-s3 helm-all


.PHONY: helm-chart-push
helm-chart-push:
	$(MAKE) -C m4d-template helm-chart-push
	$(MAKE) -C m4d-db2wh helm-chart-push
	$(MAKE) -C m4d-kafka helm-chart-push
	$(MAKE) -C m4d-s3-to-s3 helm-chart-push

.PHONY: helm-uninstall
helm-uninstall:
	$(MAKE) -C m4d-template helm-uninstall
	$(MAKE) -C m4d-db2wh helm-uninstall
	$(MAKE) -C m4d-kafka helm-uninstall
	$(MAKE) -C m4d-s3-to-s3 helm-uninstall
