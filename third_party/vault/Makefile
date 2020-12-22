ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/docker.mk

KUBE_NAMESPACE ?= m4d-system

.PHONY: deploy
deploy:
	./deploy.sh deploy

.PHONY: undeploy
undeploy:
	./deploy.sh undeploy

.PHONY: deploy-wait
deploy-wait:
	# We're using old-school while b/c we can't wait on object that haven't been created, and we can't know for sure that the statefulset had been created so far
	# See https://github.com/kubernetes/kubernetes/issues/75227
	while [[ $$(kubectl get -n ${KUBE_NAMESPACE} pods -l statefulset.kubernetes.io/pod-name=vault-0 -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do \
	    echo "waiting for vault pod to become ready"; \
	    sleep 5; \
	done
