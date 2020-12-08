export K8S_VERSION:=${K8S_VERSION}
export K8S_CLUSTER:=${K8S_CLUSTER}
export KUBECONFIG:=${KUBECONFIG}

.PHONY: kind-setup
kind-setup: $(TOOLBIN)/kind $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./create_kind.sh

.PHONY: kind-setup-multi
kind-setup-multi: $(TOOLBIN)/kind $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./create_kind.sh multi

.PHONY: kind-cleanup
kind-cleanup: $(TOOLBIN)/kind $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./create_kind.sh cleanup

.PHONY: kind
kind: kind-cleanup kind-setup
