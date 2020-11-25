ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk

KUBE_NAMESPACE ?= m4d-system

.PHONY: deploy
deploy:
	./deploy.sh deploy

.PHONY: deploy-wait
deploy-wait: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl wait --for=condition=ready -n ${KUBE_NAMESPACE} pod/vault-0 --timeout=180s

.PHONY: undeploy
undeploy:
	./deploy.sh undeploy

.PHONY: wait_for_vault
wait_for_vault:
	./deploy.sh wait_for_vault
