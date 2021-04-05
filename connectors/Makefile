ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

CONNECTORS := \
	katalog \
	egeria \
	opa \
	vault

define test-target
  test:: ; cd $1 && make test
endef

define docker-build-target
  docker-build:: ; cd $1 && make docker-build
endef

define docker-push-target
  docker-push:: ; cd $1 && make docker-push
endef

define docker-rmi-target
  docker-rmi:: ; cd $1 && make docker-rmi
endef

$(foreach dir,$(CONNECTORS),$(eval $(call test-target,$(dir))))
$(foreach dir,$(CONNECTORS),$(eval $(call docker-build-target,$(dir))))
$(foreach dir,$(CONNECTORS),$(eval $(call docker-push-target,$(dir))))
$(foreach dir,$(CONNECTORS),$(eval $(call docker-rmi-target,$(dir))))

.PHONY: deploy undeploy charts
deploy undeploy charts:
	$(MAKE) -C vault $@
	$(MAKE) -C opa $@
	$(MAKE) -C egeria $@
