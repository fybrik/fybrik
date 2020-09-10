ROOT_DIR := ..
include $(ROOT_DIR)/Makefile.env

DOCKER_NAME ?= secret-provider
DOCKER_TAGNAME ?= latest
include $(ROOT_DIR)/.mk/docker.mk

.PHONY: deploy
deploy:
	cd deployment && ./deploy-secret-provider.sh deploy || echo "failed ./deploy-secret-provider.sh deploy"

.PHONY: undeploy
undeploy:
	cd deployment && ./deploy-secret-provider.sh undeploy || echo "failed ./deploy-secret-provider.sh undeploy"

.PHONY: configure-vault
configure-vault:
	cd deployment && ./deploy-vault.sh configure || echo "failed to configure vault"

.PHONY: vault-demo-secrets
vault-demo-secrets:
	cd deployment && ./deploy-vault.sh populate_demo_secrets || echo "failed to write demo secrets to vault"
