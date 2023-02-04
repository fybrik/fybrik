include $(TOOLS_DIR)/requirements.env
SKIP_INSTALL_CHECK ?= true

define post-install-check
	$(SKIP_INSTALL_CHECK) || go mod tidy
	$(SKIP_INSTALL_CHECK) || git diff --exit-code -- go.mod
endef

INSTALL_TOOLS += $(TOOLBIN)/yq
.PHONY: $(TOOLBIN)/yq
$(TOOLBIN)/yq:
	cd $(TOOLS_DIR); ./install_yq.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/controller-gen
$(TOOLBIN)/controller-gen:
	GOBIN=$(ABSTOOLBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CONTROLLER_GEN_VERSION) 
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/dlv
$(TOOLBIN)/dlv:
	GOBIN=$(ABSTOOLBIN) go install github.com/go-delve/delve/cmd/dlv@v$(DLV_VERSION)
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/helm
.PHONY: $(TOOLBIN)/helm
$(TOOLBIN)/helm:
	cd $(TOOLS_DIR); ./install_helm.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/golangci-lint
$(TOOLBIN)/golangci-lint:
	GOBIN=$(ABSTOOLBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION)
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/kubebuilder
.PHONY: $(TOOLBIN)/kubebuilder $(TOOLBIN)/etcd $(TOOLBIN)/kube-apiserver $(TOOLBIN)/kubectl
$(TOOLBIN)/kubebuilder $(TOOLBIN)/etcd $(TOOLBIN)/kube-apiserver $(TOOLBIN)/kubectl: $(TOOLBIN)/yq
	cd $(TOOLS_DIR); ./install_kubebuilder.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/kustomize
.PHONY: $(TOOLBIN)/kustomize
$(TOOLBIN)/kustomize:
	cd $(TOOLS_DIR); ./install_kustomize.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/kind
$(TOOLBIN)/kind:
	GOBIN=$(ABSTOOLBIN) go install sigs.k8s.io/kind@v$(KIND_VERSION)
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/istioctl
.PHONY: $(TOOLBIN)/istioctl
$(TOOLBIN)/istioctl:
	cd $(TOOLS_DIR); ./install_istio.sh
	$(call post-install-check)

# INSTALL_TOOLS += $(TOOLBIN)/oc
.PHONY: $(TOOLBIN)/oc
$(TOOLBIN)/oc:
	cd $(TOOLS_DIR); ./install_oc.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/misspell
$(TOOLBIN)/misspell:
	GOBIN=$(ABSTOOLBIN) go install github.com/client9/misspell/cmd/misspell@v$(MISSPELL_VERSION)
	$(call post-install-check)

$(TOOLBIN)/license_finder:
	gem install license_finder -v 6.5.0 --bindir=$(ABSTOOLBIN)
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/opa
.PHONY: $(TOOLBIN)/opa
$(TOOLBIN)/opa:
	cd $(TOOLS_DIR); ./install_opa.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/solver
.PHONY: $(TOOLBIN)/solver
$(TOOLBIN)/solver:
	cd $(TOOLS_DIR); ./install_or_tools.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/vault
.PHONY: $(TOOLBIN)/vault
$(TOOLBIN)/vault:
	cd $(TOOLS_DIR); ./install_vault.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/oapi-codegen
$(TOOLBIN)/oapi-codegen:
	GOBIN=$(ABSTOOLBIN) go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v$(OAPI_CODEGEN_VERSION)
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/crdoc
$(TOOLBIN)/crdoc:
	GOBIN=$(ABSTOOLBIN) go install fybrik.io/crdoc@v$(CRDOC_VERSION)
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/json-schema-generator
.PHONY: $(TOOLBIN)/json-schema-generator
$(TOOLBIN)/json-schema-generator:
	cd $(TOOLS_DIR); ./install_json-schema-generator.sh
	$(call post-install-check)

.PHONY: install-tools
install-tools: $(INSTALL_TOOLS)
	go mod tidy
	ls -l $(TOOLS_DIR)/bin

.PHONY: uninstall-tools
uninstall-tools:
	find $(TOOLBIN) -mindepth 1 -delete
