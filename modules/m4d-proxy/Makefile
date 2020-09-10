ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env

NAME ?= bucket1
BUCKET ?= m4d-objectstorage-secret-provider-test
ENDPOINT ?= s3.eu-de.cloud-object-storage.appdomain.cloud

CHART ?= m4d-proxy

CONFIG := \
	--set proxy.host=${BUCKET}.${ENDPOINT}

.PHONY: debug
debug:
	helm install --debug --dry-run ${NAME} ../${CHART} $(CONFIG)

.PHONY: install
install:
	helm install ${NAME} ../${CHART} $(CONFIG)

.PHONY: uninstall
uninstall:
	helm uninstall ${NAME}

.PHONY: list
list:
	helm list

.PHONY: template
template:
	helm template ${NAME} ../${CHART} $(CONFIG)

.PHONY: status
status:
	kubectl get all -l "app.kubernetes.io/instance=${NAME}"
	kubectl get istio-io -l "app.kubernetes.io/instance=${NAME}"
