ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

CONNECTORS := \
	egeria \
	opa \
	vault

define build-target
  build:: ; cd $1 && make build
endef

define clean-target
  clean:: ; cd $1 && make clean
endef

define test-target
  test:: ; cd $1 && make test
endef

define run-target
  run:: ; cd $1 && make run
endef

define terminate-target
  terminate:: ; cd $1 && make terminate
endef

define docker-all-target
  docker-all:: ; cd $1 && make docker-all
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

$(foreach dir,$(CONNECTORS),$(eval $(call build-target,$(dir))))
$(foreach dir,$(CONNECTORS),$(eval $(call clean-target,$(dir))))
$(foreach dir,$(CONNECTORS),$(eval $(call test-target,$(dir))))
$(foreach dir,$(CONNECTORS),$(eval $(call run-target,$(dir))))
$(foreach dir,$(CONNECTORS),$(eval $(call terminate-target,$(dir))))
$(foreach dir,$(CONNECTORS),$(eval $(call docker-all-target,$(dir))))
$(foreach dir,$(CONNECTORS),$(eval $(call docker-build-target,$(dir))))
$(foreach dir,$(CONNECTORS),$(eval $(call docker-push-target,$(dir))))
$(foreach dir,$(CONNECTORS),$(eval $(call docker-rmi-target,$(dir))))

.PHONY: deploy undeploy charts
deploy undeploy charts:
	$(MAKE) -C vault $@
	$(MAKE) -C opa $@
	$(MAKE) -C egeria $@
