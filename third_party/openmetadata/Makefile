include Makefile.env

all: prepare-openmetadata-for-fybrik

deploy-openmetadata:
	helm repo add open-metadata https://helm.open-metadata.org/ || true
	helm repo update || true
	kubectl create ns ${INSTALLATION_NAMESPACE} || true
ifeq ($(IBM_OPENSHIFT_INSTALLATION), true)
	oc adm policy add-scc-to-user anyuid -z airflow -n ${INSTALLATION_NAMESPACE}
	oc adm policy add-scc-to-user anyuid -z mysql -n ${INSTALLATION_NAMESPACE}
	oc adm policy add-scc-to-user ibm-privileged-scc -z default -n ${INSTALLATION_NAMESPACE}
	kubectl apply -f ibm-openshift/pvc1.yaml -n ${INSTALLATION_NAMESPACE} || true
	kubectl apply -f ibm-openshift/pvc2.yaml -n ${INSTALLATION_NAMESPACE} || true
	kubectl apply -f ibm-openshift/pvc3.yaml -n ${INSTALLATION_NAMESPACE} || true
	kubectl apply -f ibm-openshift/pvc4.yaml -n ${INSTALLATION_NAMESPACE} || true
else
	sed s/INSTALLATION_NAMESPACE/${INSTALLATION_NAMESPACE}/g pv1.yaml | kubectl apply -n ${INSTALLATION_NAMESPACE} -f - || true
	kubectl apply -f pvc1.yaml -n ${INSTALLATION_NAMESPACE} || true
	sed s/INSTALLATION_NAMESPACE/${INSTALLATION_NAMESPACE}/g pv2.yaml | kubectl apply -n ${INSTALLATION_NAMESPACE} -f - || true
	kubectl apply -f pvc2.yaml -n ${INSTALLATION_NAMESPACE} || true
endif
	@kubectl create secret generic airflow-mysql-secrets --from-literal=airflow-mysql-password=${AIRFLOW_MYSQL_PASSWORD} -n ${INSTALLATION_NAMESPACE} || true
	echo installing openmetadata-dependencies - this might take a while
	@helm install openmetadata-dependencies open-metadata/openmetadata-dependencies --version ${OPENMETADATA_HELM_CHART_VERSION} --values values-deps.yaml --wait --timeout 20m -n ${INSTALLATION_NAMESPACE} \
--set airflow.airflow.image.tag=${OPENMETADATA_AIRFLOW_VERSION} \
--set mysql.initdbScripts."init_openmetadata_db_scripts\.sql"=${INIT_OPENMETADATA_DB_SCRIPT} \
--set mysql.initdbScripts."init_airflow_db_scripts\.sql"=${INIT_AIRFLOW_DB_SCRIPT} \
--set airflow.airflow.users[0].username=${OPENMETADATA_AIRFLOW_USER} \
--set airflow.airflow.users[0].password=${OPENMETADATA_AIRFLOW_PASSWORD} \
--set airflow.airflow.users[0].firstName=${AIRFLOW_ADMIN_FIRST_NAME} \
--set airflow.airflow.users[0].lastName=${AIRFLOW_ADMIN_LAST_NAME} \
--set airflow.airflow.users[0].email=${AIRFLOW_ADMIN_EMAIL} \
--set airflow.externalDatabase.user=${AIRFLOW_MYSQL_USER} \
--set airflow.airflow.config.AIRFLOW__LINEAGE__OPENMETADATA_API_ENDPOINT="${OPENMETADATA_ENDPOINT}" \
--set airflow.externalDatabase.host=${MYSQL_ENDPOINT} \
--set mysql.auth.rootPassword=${MYSQL_ROOT_PASSWORD} || true

	@kubectl create secret generic mysql-secrets --from-literal=openmetadata-mysql-password=${OPENMETADATA_MYSQL_PASSWORD} -n ${INSTALLATION_NAMESPACE} || true
	@kubectl create secret generic airflow-secrets --from-literal=openmetadata-airflow-password=${OPENMETADATA_AIRFLOW_PASSWORD} -n ${INSTALLATION_NAMESPACE} || true
	echo installing openmetadata
	@helm install openmetadata open-metadata/openmetadata --version ${OPENMETADATA_HELM_CHART_VERSION} --wait --timeout 20m -n ${INSTALLATION_NAMESPACE} \
--set global.airflow.host=${AIRFLOW_ENDPOINT} \
--set global.airflow.openmetadata.serverHostApiUrl=${OPENMETADATA_ENDPOINT} \
--set global.database.auth.username=${OPENMETADATA_MYSQL_USER} \
--set global.airflow.auth.username=${OPENMETADATA_AIRFLOW_USER} \
--set extraEnvs[0].name=AUTHORIZER_ADMIN_PRINCIPALS,extraEnvs[0].value='[${OPENMETADATA_USER}:${OPENMETADATA_PASSWORD}]' || true

define JOB_YAML
apiVersion: batch/v1
kind: Job
metadata:
  name: prepare-openmetadata
  namespace: $(INSTALLATION_NAMESPACE)
spec:
  template:
    spec:
      containers:
      - name: prepare-openmetadata
        image: 'ghcr.io/fybrik/openmetadata-connector:0.0.0'
        command: ["/openmetadata-connector", "prepare", "--customization", "/customization.yaml"]
        env:
        - name: OPENMETADATA_ENDPOINT
          value: http://openmetadata.$(INSTALLATION_NAMESPACE):8585/api
        - name: OPENMETADATA_USER
          value: $(OPENMETADATA_USER)
        - name: OPENMETADATA_PASSWORD
          value: $(OPENMETADATA_PASSWORD)
      restartPolicy: Never
  backoffLimit: 4
endef

export JOB_YAML

prepare-openmetadata-for-fybrik: deploy-openmetadata
	@echo "$$JOB_YAML" | kubectl apply -f -
	kubectl wait --for=condition=complete job/prepare-openmetadata -n open-metadata
	kubectl delete job prepare-openmetadata -n open-metadata
