ROOT_DIR:=../../
include $(ROOT_DIR)/Makefile.env

.PHONY: docker-build-and-push
docker-build-and-push:
	$(MAKE) -C datacatalog all
	$(MAKE) -C policycompiler all
