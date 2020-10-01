ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

DOCKER_NAME ?= secret-provider
DOCKER_TAGNAME ?= latest
include $(ROOT_DIR)/.mk/docker.mk

.PHONY: deploy
deploy:
	cd deploy && ./deploy.sh deploy

.PHONY: undeploy
undeploy:
	cd deploy && ./deploy.sh undeploy

.PHONY: configure-vault
configure-vault:
	cd deploy && ./deploy-vault.sh configure || echo "failed to configure vault"

.PHONY: vault-demo-secrets
vault-demo-secrets:
	cd deploy && ./deploy-vault.sh populate_demo_secrets || echo "failed to write demo secrets to vault"
