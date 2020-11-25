export K8S_VERSION:=${K8S_VERSION}
export K8S_CLUSTER:=${K8S_CLUSTER}
export KUBECONFIG:=${KUBECONFIG}

.PHONY: minikube-setup
minikube-setup: $(TOOLBIN)/minikube $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./create_minikube.sh

.PHONY: minikube-cleanup
minikube-cleanup: $(TOOLBIN)/minikube $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./create_minikube.sh cleanup

.PHONY: kind-setup
kind-setup: $(TOOLBIN)/kind $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./create_kind.sh

.PHONY: kind-setup-multi
kind-setup-multi: $(TOOLBIN)/kind $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./create_kind.sh multi

.PHONY: kind-cleanup
kind-cleanup: $(TOOLBIN)/kind $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./create_kind.sh cleanup

.PHONY: istio-setup
istio-setup: $(TOOLBIN)/kubectl $(TOOLBIN)/istioctl
	cd $(TOOLS_DIR); ./create_istio.sh

.PHONY: istio-cleanup
istio-cleanup: $(TOOLBIN)/kubectl $(TOOLBIN)/istioctl
	cd $(TOOLS_DIR); ./create_istio.sh cleanup

.PHONY: istio-status
istio-status: $(TOOLBIN)/kubectl
	cd $(TOOLS_DIR); ./create_istio.sh status

.PHONY: kind
kind: kind-cleanup kind-setup istio-setup istio-status

.PHONY: minikube
minikube: minikube-cleanup minikube-setup istio-setup istio-status