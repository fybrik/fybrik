ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

.PHONY: run-integration-tests
run-integration-tests: export ROOT_DIR := .
run-integration-tests:
	cd .. && $(MAKE) kind
	helm install -g m4d-crd
	kubectl create namespace m4d-system
	cd ../third_party/cert-manager && $(MAKE) deploy deploy-wait
	cd ../third_party/vault && $(MAKE) deploy wait_for_vault
	helm install -g m4d
