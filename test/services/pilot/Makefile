ROOT_DIR := ../../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/.mk/docker.mk

MOCK_DIRS := \
	catalog-credential-mock \
	policy-manager-mock

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
  dockerall:: ; cd $1 && make docker-all
endef

define docker-build-target
  dockerbuild:: ; cd $1 && make docker-build
endef

define docker-push-target
  dockerpush:: ; cd $1 && make docker-push
endef

define docker-rmi-target
  dockerrmi:: ; cd $1 && make docker-rmi
endef

$(foreach dir,$(MOCK_DIRS),$(eval $(call build-target,$(dir))))
$(foreach dir,$(MOCK_DIRS),$(eval $(call clean-target,$(dir))))
$(foreach dir,$(MOCK_DIRS),$(eval $(call test-target,$(dir))))
$(foreach dir,$(MOCK_DIRS),$(eval $(call run-target,$(dir))))
$(foreach dir,$(MOCK_DIRS),$(eval $(call terminate-target,$(dir))))
$(foreach dir,$(MOCK_DIRS),$(eval $(call docker-all-target,$(dir))))
$(foreach dir,$(MOCK_DIRS),$(eval $(call docker-build-target,$(dir))))
$(foreach dir,$(MOCK_DIRS),$(eval $(call docker-push-target,$(dir))))
$(foreach dir,$(MOCK_DIRS),$(eval $(call docker-rmi-target,$(dir))))



.PHONY: deploy
deploy:
	ROOT_DIR=$(ROOT_DIR) ./deploy_mocks.sh

.PHONY: undeploy
undeploy:
	ROOT_DIR=$(ROOT_DIR) ./deploy_mocks.sh undeploy
