ROOT_DIR:=../../
include $(ROOT_DIR)/Makefile.env

.PHONY: docker
docker:
	$(MAKE) -C datacatalog all
	$(MAKE) -C policycompiler all
