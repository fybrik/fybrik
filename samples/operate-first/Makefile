ROOT_DIR := ../..
include $(ROOT_DIR)/Makefile.env
export RELEASE_NAME=m4d
export NAMESPACE=m4d-system
export M4D_CHART_PATH=$(ROOT_DIR)/charts/m4d
export CRD_CHART_PATH=$(ROOT_DIR)/charts/m4d-crd

.PHONY: helm-template
helm-template: $(TOOLBIN)/yq $(TOOLBIN)/helm
	$(TOOLBIN)/helm template $(RELEASE_NAME) -n $(NAMESPACE) $(M4D_CHART_PATH) | $(TOOLBIN)/yq e 'select(.kind=="Namespace" or .kind=="MutatingWebhookConfiguration" or .kind=="ValidatingWebhookConfiguration" or .kind=="ClusterRole" or .kind=="ClusterRoleBinding")' - > m4d.yaml
	$(TOOLBIN)/helm template $(RELEASE_NAME) -n $(NAMESPACE) $(CRD_CHART_PATH) > m4d-crd.yaml
	python -m pip install -r requirements.txt
	python filterClusterScoped.py

.PHONY: generate
generate:
	python -m pip install -r requirements.txt
	python filterClusterScoped.py

all: helm-crd-hook helm-verify
