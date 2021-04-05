ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

DOCKER_NAME ?= secret-provider
DOCKER_TAGNAME ?= latest
include $(ROOT_DIR)/hack/make-rules/docker.mk

.PHONY: docker-build
docker-build:
	docker build . -t ${IMG}

.PHONY: deploy
deploy:
	cd deploy && ./deploy.sh deploy

.PHONY: undeploy
undeploy:
	cd deploy && ./deploy.sh undeploy

.PHONY: configure-vault
configure-vault:
	cd deploy && ./deploy.sh configure_path || echo "failed to configure vault"

.PHONY: vault-demo-secrets
vault-demo-secrets:
	cd tests && ./test.sh populate_demo_secrets || echo "failed to write demo secrets to vault"
