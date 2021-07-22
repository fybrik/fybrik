ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
export RELEASE_NAME=fybrik
export CRD_RELEASE_NAME=fybrik-crd
export NAMESPACE=fybrik-system
export Fybrik_CHART_PATH=$(ROOT_DIR)/charts/fybrik
export CRD_CHART_PATH=$(ROOT_DIR)/charts/fybrik-crd

.PHONY: all
all: helm-template generate

.PHONY: helm-template
helm-template: $(TOOLBIN)/yq $(TOOLBIN)/helm
	$(TOOLBIN)/helm template $(RELEASE_NAME) -n $(NAMESPACE) $(Fybrik_CHART_PATH) | $(TOOLBIN)/yq e 'select(.kind=="Namespace" or .kind=="MutatingWebhookConfiguration" or .kind=="ValidatingWebhookConfiguration" or .kind=="ClusterRole" or .kind=="ClusterRoleBinding")' - > fybrik.yaml
	$(TOOLBIN)/helm template $(CRD_RELEASE_NAME) -n $(NAMESPACE) $(CRD_CHART_PATH) > fybrik-crd.yaml

.PHONY: generate
generate:
	python -m pip install -r requirements.txt
	python filterClusterScoped.py
