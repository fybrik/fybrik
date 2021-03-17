ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
include ${ROOT_DIR}/pkg/policy-compiler/policy-compiler.env

.PHONY: build
build:
	$(MAKE) -C $(ROOT_DIR)/connectors build
	$(MAKE) -C $(ROOT_DIR)/test/services/pilot/catalog-credential-mock build
	$(MAKE) -C $(ROOT_DIR)/test/services/pilot/policy-manager-mock build

.PHONY: run
run:
	$(MAKE) -C $(ROOT_DIR)/connectors run
	$(MAKE) -C $(ROOT_DIR)/test/services/pilot/catalog-credential-mock run
	$(MAKE) -C $(ROOT_DIR)/test/services/pilot/policy-manager-mock run

.PHONY: terminate
terminate:
	$(MAKE) -C $(ROOT_DIR)/connectors terminate
	$(MAKE) -C $(ROOT_DIR)/test/services/pilot/catalog-credential-mock terminate
	$(MAKE) -C $(ROOT_DIR)/test/services/pilot/policy-manager-mock terminate

.PHONY: clean
clean:
	$(MAKE) -C $(ROOT_DIR)/connectors clean
	$(MAKE) -C $(ROOT_DIR)/test/services/pilot/catalog-credential-mock clean
	$(MAKE) -C $(ROOT_DIR)/test/services/pilot/policy-manager-mock clean

.PHONY: policy-compiler-test
policy-compiler-test:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	#go tool cover -html=coverage.out

.PHONY: test
test: policy-compiler-test
	$(MAKE) -C $(ROOT_DIR)/connectors test
	$(MAKE) -C $(ROOT_DIR)/test/services/pilot/catalog-credential-mock test
	$(MAKE) -C $(ROOT_DIR)/test/services/pilot/policy-manager-mock test
