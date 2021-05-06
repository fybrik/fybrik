SKIP_INSTALL_CHECK ?= true

define post-install-check
	$(SKIP_INSTALL_CHECK) || go mod tidy
	$(SKIP_INSTALL_CHECK) || git diff --exit-code -- go.mod
endef

INSTALL_TOOLS += $(TOOLBIN)/controller-gen
$(TOOLBIN)/controller-gen:
	GOBIN=$(ABSTOOLBIN) GO111MODULE=on go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.5.0
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/crd-ref-docs
$(TOOLBIN)/crd-ref-docs:
	GOBIN=$(ABSTOOLBIN) GO111MODULE=on go get github.com/elastic/crd-ref-docs@v0.0.5
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/dlv
$(TOOLBIN)/dlv:
	GOBIN=$(ABSTOOLBIN) GO111MODULE=on go get github.com/go-delve/delve/cmd/dlv@v1.4.1
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/helm
$(TOOLBIN)/helm:
	cd $(TOOLS_DIR); ./install_helm.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/golangci-lint
$(TOOLBIN)/golangci-lint:
	GOBIN=$(ABSTOOLBIN) GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.31.0
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/kubebuilder
$(TOOLBIN)/kubebuilder $(TOOLBIN)/etcd $(TOOLBIN)/kube-apiserver $(TOOLBIN)/kubectl:
	cd $(TOOLS_DIR); ./install_kubebuilder.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/kustomize
$(TOOLBIN)/kustomize:
	cd $(TOOLS_DIR); ./install_kustomize.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/kind
$(TOOLBIN)/kind:
	GOBIN=$(ABSTOOLBIN) GO111MODULE=on go get sigs.k8s.io/kind@v0.9.0
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/istioctl
$(TOOLBIN)/istioctl:
	cd $(TOOLS_DIR); ./install_istio.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/protoc
$(TOOLBIN)/protoc:
	cd $(TOOLS_DIR); ./install_protoc.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/protoc-gen-doc
$(TOOLBIN)/protoc-gen-doc:
	GOBIN=$(ABSTOOLBIN) GO111MODULE=on go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@v1.4.1
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/protoc-gen-go
$(TOOLBIN)/protoc-gen-go:
	# FIXME: using `go install` to fix broken deps but needs to be reverted back to `go get`
	# GOBIN=$(ABSTOOLBIN) GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@v1.3.5
	GOBIN=$(ABSTOOLBIN) GO111MODULE=on go install github.com/golang/protobuf/protoc-gen-go
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/protoc-gen-lint
$(TOOLBIN)/protoc-gen-lint:
	GOBIN=$(ABSTOOLBIN) GO111MODULE=on go get github.com/ckaznocha/protoc-gen-lint@v0.2.1
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/protoc-gen-deepcopy
$(TOOLBIN)/protoc-gen-deepcopy:
	GOBIN=$(ABSTOOLBIN) GO111MODULE=on go get istio.io/tools/cmd/protoc-gen-deepcopy@1.8.2
	$(call post-install-check)
	   
# INSTALL_TOOLS += $(TOOLBIN)/oc
$(TOOLBIN)/oc:
	cd $(TOOLS_DIR); ./install_oc.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/addlicense
$(TOOLBIN)/addlicense:
	GOBIN=$(ABSTOOLBIN) GO111MODULE=on go get github.com/google/addlicense
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/misspell
$(TOOLBIN)/misspell:
	GOBIN=$(ABSTOOLBIN) GO111MODULE=on go get github.com/client9/misspell/cmd/misspell@v0.3.4
	$(call post-install-check)

$(TOOLBIN)/license_finder:
	gem install license_finder -v 6.5.0 --bindir=$(ABSTOOLBIN)
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/opa
$(TOOLBIN)/opa:
	cd $(TOOLS_DIR); ./install_opa.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/vault
$(TOOLBIN)/vault:
	cd $(TOOLS_DIR); ./install_vault.sh
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/oapi-codegen
$(TOOLBIN)/oapi-codegen:
	GOBIN=$(ABSTOOLBIN) GO111MODULE=on go get github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.4.2
	$(call post-install-check)

INSTALL_TOOLS += $(TOOLBIN)/openapi2crd
$(TOOLBIN)/openapi2crd:
	curl -sfL https://raw.githubusercontent.com/mesh-for-data/openapi2crd/main/hack/get-openapi2crd.sh | sh -s -- -b $(ABSTOOLBIN) -d v0.1.4

INSTALL_TOOLS += $(TOOLBIN)/crdoc
$(TOOLBIN)/crdoc:
	curl -sfL https://raw.githubusercontent.com/mesh-for-data/crdoc/main/hack/get-crdoc.sh | sh -s -- -b $(ABSTOOLBIN) -d v0.1.2

INSTALL_TOOLS += $(TOOLBIN)/yq
$(TOOLBIN)/yq:
	cd $(TOOLS_DIR); ./install_yq.sh
	$(call post-install-check)

.PHONY: install-tools
install-tools: $(INSTALL_TOOLS)
	go mod tidy

.PHONY: uninstall-tools
uninstall-tools:
	rm -rf $(INSTALL_TOOLS)
