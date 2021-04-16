export RELEASE_NAME=m4d
export NAMESPACE=m4d-system
export M4D_CHART_PATH=../../charts/m4d
export CRD_CHART_PATH=../../charts/m4d-crd

.PHONY: helm-template
helm-template: 
	helm template $(RELEASE_NAME) -n $(NAMESPACE) $(M4D_CHART_PATH) | ../../hack/tools/bin/yq e 'select(.kind=="Namespace" or .kind=="MutatingWebhookConfiguration" or .kind=="ValidatingWebhookConfiguration" or .kind=="ClusterRole" or .kind=="ClusterRoleBinding")' - > m4d.yaml
	helm template $(RELEASE_NAME) -n $(NAMESPACE) $(M4D_CHART_PATH) > m4d-crd.yaml
	python3 -m pip install -r requirements.txt
	python3 filterClusterScoped.py
