ROOT_DIR:=../..
include $(ROOT_DIR)/Makefile.env
include $(ROOT_DIR)/hack/make-rules/tools.mk

WITH_OPENSHIFT ?= false

.PHONY: deploy
deploy: $(TOOLBIN)/kubectl $(TOOLBIN)/helm
	@echo "Installing argocd ..."
	helm repo add argo https://argoproj.github.io/argo-helm
	helm repo update
	helm install argo argo/argo-cd \
		--namespace argocd \
		--create-namespace \
		--set configs.params."server.insecure"=true \
		--set server.ingress.enabled=true \
		--set openshift.enabled=$(WITH_OPENSHIFT) \
		--version v${ARGOCD_HELM_VERSION}

.PHONY: undeploy
undeploy: $(TOOLBIN)/kubectl $(TOOLBIN)/helm
	@echo "Unistalling argocd ..."
	kubectl delete crd applications.argoproj.io
	kubectl delete crd applicationsets.argoproj.io
	kubectl delete crd appprojects.argoproj.io  
	helm uninstall argo --namespace argocd

.PHONY: deploy-wait
deploy-wait: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl wait --for=condition=ready --all pod -n argocd --timeout=120s
