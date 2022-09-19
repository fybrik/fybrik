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
	GOBIN=$(ABSTOOLBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/dlv
$(TOOLBIN)/dlv:
	GOBIN=$(ABSTOOLBIN) go install github.com/go-delve/delve/cmd/dlv@v1.4.1
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/helm
.PHONY: $(TOOLBIN)/helm
$(TOOLBIN)/helm:
	cd $(TOOLS_DIR); ./install_helm.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/golangci-lint
$(TOOLBIN)/golangci-lint:
	GOBIN=$(ABSTOOLBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.44.1
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/kubebuilder
.PHONY: $(TOOLBIN)/kubebuilder $(TOOLBIN)/etcd $(TOOLBIN)/kube-apiserver $(TOOLBIN)/kubectl
$(TOOLBIN)/kubebuilder $(TOOLBIN)/etcd $(TOOLBIN)/kube-apiserver $(TOOLBIN)/kubectl:
	cd $(TOOLS_DIR); ./install_kubebuilder.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/kustomize
.PHONY: $(TOOLBIN)/kustomize
$(TOOLBIN)/kustomize:
	cd $(TOOLS_DIR); ./install_kustomize.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/kind
$(TOOLBIN)/kind:
	GOBIN=$(ABSTOOLBIN) go install sigs.k8s.io/kind@v0.13.0
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
	GOBIN=$(ABSTOOLBIN) go install github.com/client9/misspell/cmd/misspell@v0.3.4
	$(call post-install-check)

$(TOOLBIN)/license_finder:
	gem install license_finder -v 6.5.0 --bindir=$(ABSTOOLBIN)
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/opa
.PHONY: $(TOOLBIN)/opa
$(TOOLBIN)/opa:
	cd $(TOOLS_DIR); ./install_opa.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/fzn-or-tools
.PHONY: $(TOOLBIN)/fzn-or-tools
$(TOOLBIN)/fzn-or-tools:
	cd $(TOOLS_DIR); ./install_or_tools.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/vault
.PHONY: $(TOOLBIN)/vault
$(TOOLBIN)/vault:
	cd $(TOOLS_DIR); ./install_vault.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/oapi-codegen
$(TOOLBIN)/oapi-codegen:
	GOBIN=$(ABSTOOLBIN) go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.4.2
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/openapi-generator-cli
.PHONY: $(TOOLBIN)/openapi-generator-cli
$(TOOLBIN)/openapi-generator-cli:
	cd $(TOOLS_DIR); chmod +x ./install_openapi-generator-cli.sh; ./install_openapi-generator-cli.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/crdoc
$(TOOLBIN)/crdoc:
	GOBIN=$(ABSTOOLBIN) go install fybrik.io/crdoc@v0.6.1
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
