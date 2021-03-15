ROOT_DIR:=../../
include $(ROOT_DIR)/Makefile.env

.PHONY: docker-build
docker-build:
	$(MAKE) -C datacatalog docker-build
	$(MAKE) -C policycompiler docker-build

.PHONY: docker-push
docker-push:
	$(MAKE) -C datacatalog docker-push
	$(MAKE) -C policycompiler docker-push